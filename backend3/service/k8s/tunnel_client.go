package k8s

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"oneops/backend3/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TunnelClient Arthas Tunnel Server 客户端
//
// 架构说明（对齐官方 arthas-tunnel-server 3.x/4.x 实现）：
//   - 7777 Netty WS（/ws）：agent 反连注册（?method=agentRegister）与诊断会话
//     （?method=connectArthas&id=）共用同一端口；会话握手后 tunnel-server 做纯透传
//   - 8080 HTTP：console SPA + 免认证 API。注意：/actuator/** 受 Spring Security formLogin
//     保护，官方 WebSecurityConfig 未启用 httpBasic（Basic 头会被忽略并 302 到 /login），
//     平台不可用；agent 在线探测走免认证的 /api/tunnelAgents?agentId=（不受
//     arthas.enable-detail-pages 开关限制）
//
// 平台侧配置（k8s_diagnostic_config 表，优先级：集群 DB > 全局 DB > 环境变量 > 推导默认）：
//   - arthas.tunnel.url       平台访问 tunnel-server 的 HTTP 地址，如 http://1.2.3.4:8080
//   - arthas.tunnel.session   会话 WS 基地址，如 ws://1.2.3.4:7777（留空按 HTTP 地址同主机:7777 推导）
//
// 环境变量兜底：ARTHAS_TUNNEL_URL / ARTHAS_TUNNEL_SESSION_URL
type TunnelClient struct {
	httpClient *http.Client
}

// NewTunnelClient 创建 tunnel 客户端
func NewTunnelClient() *TunnelClient {
	return &TunnelClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// 诊断平台访问 tunnel-server 的配置 key
const (
	ConfigKeyTunnelURL        = "arthas.tunnel.url"
	ConfigKeyTunnelSessionURL = "arthas.tunnel.session"
	// defaultTunnelSessionPort 会话 WS 默认端口（官方 Netty 端口）
	defaultTunnelSessionPort = "7777"
)

// getTunnelURL 平台访问 tunnel-server 的 HTTP 基地址（集群级 > 全局 > env）
func (s *DiagnosticService) getTunnelURL(clusterID string) string {
	for _, cid := range []string{clusterID, "0"} {
		if cid == "" {
			continue
		}
		if v := s.diagnosticRepo.GetConfigValue(cid, ConfigKeyTunnelURL, ""); v != "" {
			return strings.TrimRight(v, "/")
		}
	}
	if v := os.Getenv("ARTHAS_TUNNEL_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return ""
}

// getTunnelSessionURL 会话 WS 基地址（集群级 > 全局 > env > 按 HTTP 地址推导同主机:7777）
func (s *DiagnosticService) getTunnelSessionURL(clusterID string) string {
	for _, cid := range []string{clusterID, "0"} {
		if cid == "" {
			continue
		}
		if v := s.diagnosticRepo.GetConfigValue(cid, ConfigKeyTunnelSessionURL, ""); v != "" {
			return strings.TrimRight(v, "/")
		}
	}
	if v := os.Getenv("ARTHAS_TUNNEL_SESSION_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	if base := s.getTunnelURL(clusterID); base != "" {
		if u, err := url.Parse(base); err == nil {
			switch u.Scheme {
			case "https":
				u.Scheme = "wss"
			default:
				u.Scheme = "ws"
			}
			u.Host = replacePort(u.Host, defaultTunnelSessionPort)
			u.Path = ""
			return u.String()
		}
	}
	return ""
}

// replacePort 将 host:port 的端口替换为指定端口（无端口时追加）
func replacePort(hostport, port string) string {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	return net.JoinHostPort(host, port)
}

// CheckAgentOnline 探测 agent 是否在线（免认证 API，与 agentId 命名格式无关）
// 官方 endpoint: GET {tunnel}/api/tunnelAgents?agentId=xxx → {"success": true|false}
func (t *TunnelClient) CheckAgentOnline(tunnelURL, agentID string) (bool, error) {
	if tunnelURL == "" {
		return false, fmt.Errorf("未配置 tunnel-server 地址（治理页 arthas.tunnel.url 或环境变量 ARTHAS_TUNNEL_URL）")
	}
	if agentID == "" {
		return false, fmt.Errorf("缺少 agentId")
	}
	req, err := http.NewRequest(http.MethodGet, tunnelURL+"/api/tunnelAgents", nil)
	if err != nil {
		return false, fmt.Errorf("tunnel URL 无效: %w", err)
	}
	q := req.URL.Query()
	q.Set("agentId", agentID)
	req.URL.RawQuery = q.Encode()

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("访问 tunnel-server 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("tunnel-server 返回状态码 %d", resp.StatusCode)
	}
	body, err := readBodyWithLimit(resp, 1<<20)
	if err != nil {
		return false, fmt.Errorf("读取 tunnel-server 响应失败: %w", err)
	}
	var r struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return false, fmt.Errorf("解析 tunnel-server 响应失败: %w", err)
	}
	return r.Success, nil
}

// DialAgentSession 与指定 agent 建立诊断 WebSocket 会话（arthas telnet 协议文本帧透传）
// 官方协议：ws(s)://<tunnel-host>:7777/ws?method=connectArthas&id=<agentId>
// 说明：agent 不在线时 tunnel-server 在握手后立即以 Close(2000, "Can not find arthas agent...")
// 断开，错误会在首次读取时暴露；调用方应先经 CheckAgentOnline / 注册表确认在线
func (t *TunnelClient) DialAgentSession(sessionBase, agentID string) (*websocket.Conn, error) {
	if sessionBase == "" {
		return nil, fmt.Errorf("未配置 tunnel-server 会话地址（治理页 arthas.tunnel.session 或环境变量 ARTHAS_TUNNEL_SESSION_URL）")
	}
	if agentID == "" {
		return nil, fmt.Errorf("缺少 agentId")
	}

	wsURL := strings.TrimRight(sessionBase, "/") + "/ws?method=connectArthas&id=" + url.QueryEscape(agentID)
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("连接 tunnel 会话失败 (%s): %w", wsURL, err)
	}

	logger.Info("已建立 tunnel 诊断会话",
		zap.String("agentId", agentID), zap.String("url", wsURL))
	return conn, nil
}

// readBodyWithLimit 读取响应体并限制大小
func readBodyWithLimit(resp *http.Response, max int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(resp.Body, max))
}
