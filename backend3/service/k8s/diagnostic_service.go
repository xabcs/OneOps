package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/logger"
	repok8s "oneops/backend3/repository/k8s"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticService 诊断服务（Arthas Tunnel 架构）
//
// 设计要点：
//   - 目标发现来自 tunnel-server agent 注册表（应用接入 starter/webhook 注入后自动注册）
//   - 命令目录内置全量 Arthas 命令 + 风险分级（L0-L5），diag_command_override 表可覆盖
//   - one-shot 执行经 tunnel WS 建立临时会话执行单条命令，结束自动 reset 并断开
//   - 交互式终端由 WS 双桥 controller 实现，本服务提供数据与执行原语
type DiagnosticService struct {
	clusterSvc     *K8sClusterService
	diagnosticRepo *repok8s.DiagnosticRepository
	tunnelClient   *TunnelClient
}

// NewDiagnosticService 创建诊断服务
func NewDiagnosticService(clusterSvc *K8sClusterService, diagnosticRepo *repok8s.DiagnosticRepository) *DiagnosticService {
	return &DiagnosticService{
		clusterSvc:     clusterSvc,
		diagnosticRepo: diagnosticRepo,
		tunnelClient:   NewTunnelClient(),
	}
}

// ========== 命令目录（内置默认值，可被 diag_command_override 覆盖） ==========

// CommandCatalogItem 命令目录条目
type CommandCatalogItem struct {
	Command     string `json:"command"`         // 命令名（首 token）
	Category    string `json:"category"`        // 分类
	Description string `json:"description"`     // 说明
	RiskLevel   string `json:"riskLevel"`       // L0-L5
	Streaming   bool   `json:"streaming"`       // 是否流式持续输出（须在终端中使用）
	Usage       string `json:"usage,omitempty"` // 用法示例
}

// builtinCommandCatalog 内置 Arthas 全量命令目录（随官方版本演进可增补）
var builtinCommandCatalog = []CommandCatalogItem{
	// --- 会话与基础 ---
	{"help", "基础", "查看全部命令帮助", "L0", false, "help"},
	{"version", "基础", "查看 Arthas 版本", "L0", false, "version"},
	{"session", "基础", "查看当前会话", "L0", false, "session"},
	{"history", "基础", "查看命令历史", "L0", false, "history"},
	{"pwd", "基础", "查看工作目录", "L0", false, "pwd"},
	{"cls", "基础", "清屏", "L0", false, "cls"},
	{"reset", "会话控制", "还原所有被增强的类（清理字节码增强，鼓励随时使用）", "L0", false, "reset"},
	{"stop", "会话控制", "关闭 Arthas：还原增强并断开 tunnel，agent 将离线且无法重连（需重启 Pod 恢复）。仅清理增强请用 reset", "L3", false, "stop"},
	{"save-log", "会话控制", "导出本次会话全部日志", "L0", false, "save-log"},
	{"options", "基础", "查看/设置会话选项（save-result 等）", "L1", false, "options"},
	{"keymap", "基础", "查看快捷键", "L0", false, "keymap"},

	// --- JVM 运行时（只读） ---
	{"dashboard", "运行时", "实时数据面板（线程/内存/GC），持续刷新", "L1", true, "dashboard"},
	{"thread", "线程", "线程查看；thread -n 5 看 CPU Top；thread -b 找阻塞/死锁；thread <id> 看堆栈", "L1", false, "thread -n 5"},
	{"jvm", "运行时", "JVM 完整信息（内存池/GC/类加载/线程）", "L1", false, "jvm"},
	{"memory", "内存", "各内存区（Eden/Survivor/Old/Metaspace）占用", "L1", false, "memory"},
	{"sysprop", "运行时", "查看/修改系统属性（修改为变更操作，需确认）", "L1", false, "sysprop"},
	{"sysenv", "运行时", "查看 JVM 环境变量", "L1", false, "sysenv"},
	{"vmoption", "运行时", "查看/修改 JVM 选项（修改为变更操作，需确认）", "L1", false, "vmoption"},
	{"perfcounter", "运行时", "查看 JVM 性能计数器", "L1", false, "perfcounter"},
	{"mbean", "运行时", "查看 MBean 信息", "L1", false, "mbean"},
	{"getstatic", "运行时", "查看类静态属性", "L1", false, "getstatic com.example.X fieldName"},

	// --- 内存/资源敏感 ---
	{"heapdump", "内存", "堆转储（STW + 大文件，谨慎使用）", "L3", false, "heapdump --live /tmp/dump.hprof"},
	{"vmtool", "高危", "JVM 底层操作：查询对象实例/forceGc，能力等同任意代码执行", "L5", false, "vmtool --action getInstances --className java.lang.String --limit 10"},

	// --- 任意执行 ---
	{"ognl", "高危", "执行 OGNL 表达式（等同任意代码执行）", "L5", false, "ognl '@System@getProperty(\"java.version\")'"},
	{"logger", "变更", "查看/修改日志级别（热生效，用完建议还原）", "L4", false, "logger --name ROOT --level DEBUG"},

	// --- 类与类加载 ---
	{"sc", "类加载", "查看已加载类信息（来源 jar/ClassLoader）", "L1", false, "sc -d com.example.X"},
	{"sm", "类加载", "查看类的方法签名", "L1", false, "sm com.example.X"},
	{"jad", "类加载", "反编译线上运行代码", "L1", false, "jad com.example.X"},
	{"classloader", "类加载", "查看类加载器继承树与统计", "L1", false, "classloader"},
	{"dump", "类加载", "导出已加载类字节码", "L1", false, "dump com.example.X"},
	{"mc", "高危", "内存编译器（热更新前置步骤）", "L5", false, "mc /tmp/X.java"},
	{"redefine", "高危", "热替换已加载类字节码（不重启修复，极高风险）", "L5", false, "redefine /tmp/X.class"},
	{"retransform", "高危", "还原/重转换字节码（配合 mc 热更新）", "L5", false, "retransform --classPattern com.example.X"},

	// --- 方法级观测（增强类命令） ---
	{"trace", "方法观测", "方法内部调用路径逐层耗时", "L2", true, "trace com.example.Svc method 'cost > 100' -n 5"},
	{"watch", "方法观测", "观察方法入参/返回值/异常（免改代码加日志）", "L2", true, "watch com.example.Svc method '{params,returnObj}' -x 3 -n 5"},
	{"stack", "方法观测", "查看方法被谁调用（调用栈反向）", "L2", true, "stack com.example.Svc method -n 5"},
	{"tt", "方法观测", "时光隧道：记录调用快照，可回放重放", "L2", true, "tt -t com.example.Svc method -n 5"},
	{"monitor", "方法观测", "调用统计（QPS/平均耗时/失败率）", "L2", true, "monitor -c 5 com.example.Svc method"},
	{"line", "方法观测", "观察指定源码行的局部变量", "L2", true, "line com.example.X method:42"},

	// --- Profiler ---
	{"profiler", "性能剖析", "async-profiler 火焰图（start/stop/status）", "L2", false, "profiler start --event cpu --duration 30"},
	{"jfr", "性能剖析", "Java Flight Recorder 录制与查看", "L2", false, "jfr start -d 30"},
}

// 命令默认风险等级（未知命令的兜底分级）
const defaultRiskLevelUnknown = "L2"

// GetCommandCatalog 获取命令目录（内置默认 + 覆盖表合并）
func (s *DiagnosticService) GetCommandCatalog() []CommandCatalogItem {
	overrides, err := s.diagnosticRepo.ListCommandOverrides()
	if err != nil {
		logger.Warn("读取命令风险覆盖表失败，使用内置目录", zap.Error(err))
		return builtinCommandCatalog
	}
	if len(overrides) == 0 {
		return builtinCommandCatalog
	}

	overrideMap := make(map[string]modelk8s.DiagnosticCommandOverride, len(overrides))
	for _, o := range overrides {
		overrideMap[o.Command] = o
	}

	result := make([]CommandCatalogItem, 0, len(builtinCommandCatalog)+len(overrides))
	for _, item := range builtinCommandCatalog {
		if o, ok := overrideMap[item.Command]; ok {
			item.RiskLevel = o.RiskLevel
			delete(overrideMap, item.Command)
		}
		result = append(result, item)
	}
	// 覆盖表中的新命令（内置目录没有的）也加入目录
	for cmd, o := range overrideMap {
		result = append(result, CommandCatalogItem{
			Command:     cmd,
			Category:    "扩展",
			Description: o.Description,
			RiskLevel:   o.RiskLevel,
		})
	}
	return result
}

// ClassifyCommand 对完整命令行分级，返回（首token, 风险级, 是否禁用）
func (s *DiagnosticService) ClassifyCommand(commandLine string) (string, string, bool) {
	tokens := strings.Fields(commandLine)
	if len(tokens) == 0 {
		return "", defaultRiskLevelUnknown, false
	}
	cmd := strings.ToLower(tokens[0])

	overrides, err := s.diagnosticRepo.ListCommandOverrides()
	if err == nil {
		for _, o := range overrides {
			if strings.EqualFold(o.Command, cmd) {
				return cmd, o.RiskLevel, o.RiskLevel == "disabled"
			}
		}
	}
	for _, item := range builtinCommandCatalog {
		if item.Command == cmd {
			return cmd, item.RiskLevel, false
		}
	}
	// 未知命令（官方新版本新增）默认 L2，管理员可在覆盖表重新归类
	return cmd, defaultRiskLevelUnknown, false
}

// ========== 目标发现（tunnel agent 注册表） ==========

// AppSummary 应用分组摘要
type AppSummary struct {
	AppName      string `json:"appName"`
	Total        int    `json:"total"`
	Online       int    `json:"online"`
	AgentVersion string `json:"agentVersion,omitempty"`
}

// ListApps 应用分组列表
func (s *DiagnosticService) ListApps() ([]AppSummary, error) {
	if err := s.SyncAgents(); err != nil {
		logger.Warn("同步 tunnel agent 失败", zap.Error(err))
	}
	agents, err := s.diagnosticRepo.ListAgents("", false)
	if err != nil {
		return nil, fmt.Errorf("查询 agent 注册表失败: %w", err)
	}

	appMap := make(map[string]*AppSummary)
	for _, a := range agents {
		summary, ok := appMap[a.AppName]
		if !ok {
			summary = &AppSummary{AppName: a.AppName}
			appMap[a.AppName] = summary
		}
		summary.Total++
		if a.Online {
			summary.Online++
		}
		if summary.AgentVersion == "" {
			summary.AgentVersion = a.AgentVersion
		}
	}

	result := make([]AppSummary, 0, len(appMap))
	for _, v := range appMap {
		result = append(result, *v)
	}
	return result, nil
}

// ListAppAgents 查询某应用的实例（agent）列表
func (s *DiagnosticService) ListAppAgents(appName string) ([]modelk8s.DiagnosticAgent, error) {
	return s.diagnosticRepo.ListAgents(appName, false)
}

// GetAgent 查询单个 agent
func (s *DiagnosticService) GetAgent(agentID string) (*modelk8s.DiagnosticAgent, error) {
	return s.diagnosticRepo.FindAgentByAgentID(agentID)
}

// SyncAgents 同步 agent 注册表：K8s 按注入 label 发现候选 Pod，再逐个向 tunnel-server 探测在线状态
//
// 背景说明：tunnel-server 官方仅对 /actuator/** 提供 formLogin 保护（未启用 httpBasic，
// Basic 头会被忽略并 302），平台无法拉取全量列表；/api/* 为免认证接口，其中
// /api/tunnelAgents?agentId= 可精确探测单个 agent 是否在线，且不受 detail-pages 开关限制。
// 候选来源因此改为 K8s 侧 label 发现（oneops-arthas-injection=enabled），agentId 按
// 注入约定 POD_NAME-NAMESPACE 推导；注册表中的历史 agent（手动接入未打 label）仅刷新在线状态。
func (s *DiagnosticService) SyncAgents() error {
	clusterIDs, err := s.clusterSvc.ListAllClusterIDs()
	if err != nil {
		return fmt.Errorf("查询集群列表失败: %w", err)
	}
	now := time.Now()
	seen := make(map[string]bool)

	for _, id := range clusterIDs {
		cid := strconv.FormatUint(uint64(id), 10)
		tunnelURL := s.getTunnelURL(cid)
		client, err := s.clusterSvc.GetClient(id)
		if err != nil {
			logger.Warn("获取集群 client 失败，跳过 agent 发现", zap.String("clusterID", cid), zap.Error(err))
			continue
		}
		pods, err := client.CoreV1().Pods("").List(context.Background(),
			metav1.ListOptions{LabelSelector: InjectLabel + "=enabled"})
		if err != nil {
			logger.Warn("列取注入 Pod 失败", zap.String("clusterID", cid), zap.Error(err))
			continue
		}
		for i := range pods.Items {
			pod := &pods.Items[i]
			agentID := pod.Name + "-" + pod.Namespace
			s.upsertAgentRecord(&modelk8s.DiagnosticAgent{
				ClusterID: cid,
				AppName:   appNameOfPod(pod),
				AgentID:   agentID,
				PodName:   pod.Name,
				Namespace: pod.Namespace,
				Online:    s.probeAgent(tunnelURL, agentID),
				LastSeen:  now,
			})
			seen[agentID] = true
		}
	}

	// 注册表已有、但本轮未在集群中发现的 agent（手动接入未打 label / Pod 已删除）仅刷新在线状态
	if agents, err := s.diagnosticRepo.ListAgents("", false); err == nil {
		for i := range agents {
			a := &agents[i]
			if seen[a.AgentID] {
				continue
			}
			a.Online = s.probeAgent(s.getTunnelURL(a.ClusterID), a.AgentID)
			a.LastSeen = now
			s.upsertAgentRecord(a)
		}
	}
	return nil
}

// probeAgent 探测 agent 在线状态（探测失败视为离线并记录告警，不打断同步）
func (s *DiagnosticService) probeAgent(tunnelURL, agentID string) bool {
	ok, err := s.tunnelClient.CheckAgentOnline(tunnelURL, agentID)
	if err != nil {
		logger.Warn("探测 agent 在线状态失败", zap.String("agentId", agentID), zap.Error(err))
		return false
	}
	return ok
}

// upsertAgentRecord 写入/更新注册表（合并既有记录字段，避免覆盖为空值）
func (s *DiagnosticService) upsertAgentRecord(agent *modelk8s.DiagnosticAgent) {
	if agent.AppName == "" {
		agent.AppName = "unknown"
	}
	if existing, err := s.diagnosticRepo.FindAgentByAgentID(agent.AgentID); err == nil && existing != nil {
		agent.ID = existing.ID
		if agent.ClusterID == "" {
			agent.ClusterID = existing.ClusterID
		}
		if agent.PodName == "" {
			agent.PodName = existing.PodName
		}
		if agent.Namespace == "" {
			agent.Namespace = existing.Namespace
		}
		if agent.AgentVersion == "" {
			agent.AgentVersion = existing.AgentVersion
		}
	}
	if err := s.diagnosticRepo.UpsertAgent(agent); err != nil {
		logger.Warn("保存 agent 注册信息失败", zap.String("agentId", agent.AgentID), zap.Error(err))
	}
}

// appNameOfPod 从 Pod 推导应用名：常见 app label 优先，ownerReference 兜底
func appNameOfPod(pod *corev1.Pod) string {
	for _, key := range []string{"app.kubernetes.io/name", "app", "k8s-app"} {
		if v := pod.Labels[key]; v != "" {
			return v
		}
	}
	for _, ref := range pod.OwnerReferences {
		switch ref.Kind {
		case "Deployment", "StatefulSet", "DaemonSet", "Job":
			return ref.Name
		}
	}
	return ""
}

// ========== one-shot 执行 ==========

// OneShotRequest 一次性命令执行请求
type OneShotRequest struct {
	ClusterID string
	AppName   string
	AgentID   string
	Command   string
	Timeout   int // 秒
	UserID    uint
	Username  string
}

// OneShotResult 一次性命令执行结果
type OneShotResult struct {
	Output    string `json:"output"`
	Duration  int    `json:"duration"`
	RiskLevel string `json:"riskLevel"`
	Status    string `json:"status"` // success / error
	Error     string `json:"error,omitempty"`
}

// ExecuteOneShot 经 tunnel 建立临时会话执行单条命令
// 适用只读与简单命令；流式命令（dashboard/trace/watch 等）应使用终端会话
func (s *DiagnosticService) ExecuteOneShot(req *OneShotRequest) (*OneShotResult, error) {
	if req.Command == "" {
		return nil, fmt.Errorf("命令不能为空")
	}
	cmd, riskLevel, disabled := s.ClassifyCommand(req.Command)
	if disabled {
		return nil, fmt.Errorf("命令 %s 已被管理员禁用", cmd)
	}
	if req.Timeout <= 0 {
		req.Timeout = 60
	}

	agent, err := s.GetAgent(req.AgentID)
	if err != nil || agent == nil {
		return nil, fmt.Errorf("agent 不存在或未注册: %s", req.AgentID)
	}
	// 实时探测 agent 在线状态（DB 里的 online 可能是 SyncAgents 上次同步的旧值，
	// agent 可能在那之后掉线；执行前必须实时确认，避免 WS 会话超时）
	tunnelURL := s.getTunnelURL(req.ClusterID)
	online, probeErr := s.tunnelClient.CheckAgentOnline(tunnelURL, req.AgentID)
	if probeErr != nil {
		// 探测本身失败（网络问题），降级用 DB 旧值但给出警告
		logger.Warn("实时探测 agent 在线状态失败，降级使用 DB 缓存",
			zap.String("agentId", req.AgentID), zap.Error(probeErr))
	}
	if probeErr == nil {
		// 探测成功，更新 DB 状态（让后续查询拿到最新值）
		agent.Online = online
		s.upsertAgentRecord(agent)
	}
	if !agent.Online {
		return nil, fmt.Errorf("agent 离线: %s（Pod %s/%s）", req.AgentID, agent.Namespace, agent.PodName)
	}

	sessionURL := s.getTunnelSessionURL(req.ClusterID)
	start := time.Now()
	output, execErr := s.runCommandOverTunnel(sessionURL, req.AgentID, req.Command, req.Timeout)
	duration := int(time.Since(start).Milliseconds())

	result := &OneShotResult{RiskLevel: riskLevel, Duration: duration}
	execution := &modelk8s.DiagnosticExecution{
		ClusterID: req.ClusterID, AppName: agent.AppName,
		Namespace: agent.Namespace, PodName: agent.PodName, AgentID: agent.AgentID,
		Command: req.Command, RiskLevel: riskLevel, Channel: "oneshot",
		UserID: req.UserID, Username: req.Username, Timestamp: time.Now(),
	}
	if execErr != nil {
		result.Status = "error"
		result.Error = execErr.Error()
		execution.ResultStatus = "error"
		execution.ResultError = execErr.Error()
		execution.ResultOutput = output
	} else {
		result.Status = "success"
		result.Output = output
		execution.ResultStatus = "success"
		execution.ResultOutput = output
	}
	if dbErr := s.diagnosticRepo.CreateExecution(execution); dbErr != nil {
		logger.Error("保存诊断执行记录失败", zap.Error(dbErr))
	}
	return result, nil
}

// runCommandOverTunnel 在临时 tunnel 会话上执行命令并收集输出
//
// 设计：用持续运行的 reader goroutine 管理读取，避免 gorilla/websocket 的
// SetReadDeadline readErr 污染问题。reader 在连接生命周期内只启动一次，
// collect 可多次调用，共享同一个 channel。
//
// banner 与命令输出分离：Arthas 在会话建立后发送 banner（以 [arthas@N]$ 提示符
// 结尾），每条命令执行完后也会发送新的提示符。用 [arthas@ 作为 collect 的
// stopMarker，精准判定阶段边界，避免旧实现 idle 2s 太短导致 banner 被误判
// 为"无输出"而泄漏到 command collect（表现为命令输出变成 banner）。
func (s *DiagnosticService) runCommandOverTunnel(sessionURL, agentID, command string, timeoutSec int) (string, error) {
	conn, err := s.tunnelClient.DialAgentSession(sessionURL, agentID)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
		_ = conn.Close()
	}()

	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)

	// 启动持续 reader goroutine（整个连接生命周期内只启动一次）
	reader := newWSReader(conn)
	defer reader.stop()

	// 1) 等待 Arthas 就绪：收集 banner 直到出现 [arthas@ 提示符
	bannerDeadline := time.Now().Add(15 * time.Second)
	if bannerDeadline.After(deadline) {
		bannerDeadline = deadline
	}
	bannerStart := time.Now()
	banner, bannerErr := reader.collect(5*time.Second, bannerDeadline, "[arthas@")
	logger.Info("tunnel banner collect",
		zap.String("agentId", agentID),
		zap.Int("bannerLen", len(banner)),
		zap.Bool("hasPrompt", strings.Contains(banner, "[arthas@")),
		zap.Duration("cost", time.Since(bannerStart)),
		zap.NamedError("err", bannerErr))

	// 2) 发送命令
	//    Arthas WS 会话的输入协议是 JSON（termd HttpTtyConnection.writeToDecoder，
	//    官方源码 alibaba/termd 确认），输出方向才是纯文本：
	//    - {"action":"read","data":"<用户输入>"}：data 可一次带整条命令 + \r 回车
	//    - {"action":"resize","cols":120,"rows":40}：调整终端尺寸
	//    纯文本帧会在 agent 端 JSON 解析失败被静默丢弃（writeToDecoder catch 后
	//    直接 return），表现为命令无响应——这就是之前 banner 能收（输出方向
	//    纯文本）但命令不执行（输入方向要 JSON）的根因
	type relayMsg struct {
		Action string `json:"action"`
		Data   string `json:"data"`
	}
	sendRead := func(data string) error {
		payload, _ := json.Marshal(relayMsg{Action: "read", Data: data})
		return conn.WriteMessage(websocket.TextMessage, payload)
	}
	if err := sendRead(command + "\r"); err != nil {
		return "", fmt.Errorf("发送命令失败: %w", err)
	}
	logger.Info("tunnel 命令已发送(JSON read 协议)",
		zap.String("agentId", agentID), zap.String("command", command))

	// 3) 收集命令输出：直到出现下一个 [arthas@ 提示符（命令完成），或 idle 超时
	cmdStart := time.Now()
	output, err := reader.collect(8*time.Second, deadline, "[arthas@")
	logger.Info("tunnel command collect",
		zap.String("agentId", agentID),
		zap.Int("outputLen", len(output)),
		zap.Bool("hasPrompt", strings.Contains(output, "[arthas@")),
		zap.Duration("cost", time.Since(cmdStart)),
		zap.NamedError("err", err))

	// 4) 执行 reset 清理字节码增强（幂等，JSON read 协议）
	_ = sendRead("reset\r")

	// 5) 剥离末尾的 [arthas@N]$ 提示符（命令完成标志，非命令输出）
	output = trimArthasPrompt(output)

	if err != nil && output == "" {
		return "", err
	}
	return output, nil
}

// wsReader 持续读取 WebSocket 消息的 reader
// 通过 channel 供多次 collect 调用共享，避免 SetReadDeadline 的 readErr 污染
type wsReader struct {
	conn *websocket.Conn
	ch   chan wsReadResult
	done chan struct{}
}

type wsReadResult struct {
	msg []byte
	err error
}

func newWSReader(conn *websocket.Conn) *wsReader {
	r := &wsReader{
		conn: conn,
		ch:   make(chan wsReadResult, 64),
		done: make(chan struct{}),
	}
	go r.readLoop()
	return r
}

func (r *wsReader) readLoop() {
	for {
		_, msg, err := r.conn.ReadMessage()
		select {
		case r.ch <- wsReadResult{msg: msg, err: err}:
		case <-r.done:
			return
		}
		if err != nil {
			return
		}
	}
}

func (r *wsReader) stop() {
	close(r.done)
}

// collect 从 channel 收集消息。
//   - idleTimeout: 两条消息之间最大间隔，超过则视为空闲，返回已收集内容
//   - deadline:    绝对截止时间
//   - stopMarker:  检测到该子串立即返回（如 "[arthas@" 提示符，标志 banner 结束或命令完成）
//
// 为何需要 stopMarker：Arthas 在 banner 结尾与每条命令执行完后都会发送 [arthas@N]$
// 提示符。用提示符作为"阶段结束"的可靠标志，比固定 idle 超时更精准——既不会在 banner
// 延迟到达时空等（旧 2s idle 在 agent 初始化慢时会误判"无 banner"，导致 banner 消息
// 被后续 command collect 消费，命令输出变成 banner），也不会在命令完成后还傻等 5s。
func (r *wsReader) collect(idleTimeout time.Duration, deadline time.Time, stopMarker string) (string, error) {
	var sb strings.Builder
	lastReceive := time.Now()

	for {
		if time.Now().After(deadline) {
			if sb.Len() > 0 {
				return sb.String(), nil
			}
			return "", fmt.Errorf("命令执行超时")
		}

		remaining := idleTimeout - time.Since(lastReceive)
		if remaining < 0 {
			if sb.Len() > 0 {
				return sb.String(), nil
			}
			return "", fmt.Errorf("命令执行超时")
		}

		select {
		case res := <-r.ch:
			if res.err != nil {
				if sb.Len() > 0 {
					return sb.String(), nil
				}
				return "", fmt.Errorf("tunnel 会话读取失败: %w", res.err)
			}
			lastReceive = time.Now()
			sb.Write(res.msg)
			// 检测到停止标记（如 [arthas@ 提示符）立即返回，避免误消费下一阶段内容
			if stopMarker != "" && strings.Contains(sb.String(), stopMarker) {
				return sb.String(), nil
			}
			if sb.Len() > 4<<20 {
				return sb.String(), nil
			}
		case <-time.After(remaining):
			if sb.Len() > 0 {
				return sb.String(), nil
			}
			return "", fmt.Errorf("命令执行超时")
		}
	}
}

// trimArthasPrompt 剥离输出末尾的 [arthas@N]$ 提示符
// Arthas 命令完成后会发送新的提示符，这不是命令输出的一部分
func trimArthasPrompt(s string) string {
	idx := strings.LastIndex(s, "[arthas@")
	if idx < 0 {
		return s
	}
	return strings.TrimRight(s[:idx], " \t\r\n")
}

// ========== 历史与总览 ==========

// GetExecutions 分页查询执行记录
func (s *DiagnosticService) GetExecutions(q repok8s.ExecutionQuery) ([]modelk8s.DiagnosticExecution, int64, error) {
	return s.diagnosticRepo.FindExecutionsWithPagination(q)
}

// OverviewData 总览数据
type OverviewData struct {
	Apps        int64                   `json:"apps"`        // 接入应用数
	AgentTotal  int64                   `json:"agentTotal"`  // agent 总数
	AgentOnline int64                   `json:"agentOnline"` // 在线 agent 数
	ExecStats   *repok8s.ExecutionStats `json:"execStats"`   // 执行统计
	AppList     []AppSummary            `json:"appList"`     // 应用列表
}

// GetOverview 总览
func (s *DiagnosticService) GetOverview() (*OverviewData, error) {
	if err := s.SyncAgents(); err != nil {
		logger.Warn("同步 tunnel agent 失败", zap.Error(err))
	}
	total, online, err := s.diagnosticRepo.CountAgents()
	if err != nil {
		return nil, err
	}
	stats, err := s.diagnosticRepo.GetExecutionStats()
	if err != nil {
		return nil, err
	}
	apps, err := s.ListApps()
	if err != nil {
		return nil, err
	}
	return &OverviewData{
		Apps: int64(len(apps)), AgentTotal: total, AgentOnline: online,
		ExecStats: stats, AppList: apps,
	}, nil
}

// ========== 会话数据访问（WS controller 使用） ==========

// GetActiveSessions 活跃会话列表
func (s *DiagnosticService) GetActiveSessions() ([]modelk8s.DiagnosticSession, error) {
	return s.diagnosticRepo.ListActiveSessions()
}

// GetSessionByID 按 ID 查会话
func (s *DiagnosticService) GetSessionByID(id uint) (*modelk8s.DiagnosticSession, error) {
	return s.diagnosticRepo.FindSessionByID(id)
}

// RecordTerminalCommand 终端会话内命令入库（由 WS 桥在识别到用户回车行时调用）
func (s *DiagnosticService) RecordTerminalCommand(session *modelk8s.DiagnosticSession, commandLine string, output string, durationMs int, success bool) {
	_, riskLevel, disabled := s.ClassifyCommand(commandLine)
	execution := &modelk8s.DiagnosticExecution{
		SessionID: session.ID, ClusterID: session.ClusterID, AppName: session.AppName,
		AgentID: session.AgentID, Command: commandLine,
		RiskLevel: riskLevel, Channel: "terminal",
		ResultStatus: "success", ResultOutput: output, Duration: durationMs,
		UserID: session.UserID, Username: session.Username, Timestamp: time.Now(),
	}
	if disabled {
		execution.RiskLevel = "disabled"
	}
	if !success {
		execution.ResultStatus = "error"
	}
	if err := s.diagnosticRepo.CreateExecution(execution); err != nil {
		logger.Error("保存终端命令记录失败", zap.Error(err))
	}
}

// CloseSession 关闭会话（status: closed/timeout/terminated）
func (s *DiagnosticService) CloseSession(id uint, status string) error {
	session, err := s.diagnosticRepo.FindSessionByID(id)
	if err != nil {
		return err
	}
	if session.Status != "active" {
		return nil
	}
	now := time.Now()
	session.Status = status
	session.EndAt = &now
	return s.diagnosticRepo.UpdateSession(session)
}

// EnsureSessionActive 校验 agent 无活跃会话并创建（互斥；竞态由唯一键兜底）
func (s *DiagnosticService) EnsureSessionActive(clusterID, appName, agentID string, userID uint, username string) (*modelk8s.DiagnosticSession, error) {
	if existing, err := s.diagnosticRepo.FindSessionByKey(agentID); err == nil && existing != nil {
		return nil, fmt.Errorf("实例 %s 已有活跃会话（操作者 %s，开始于 %s），同一 JVM 仅允许一个诊断会话",
			agentID, existing.Username, existing.StartAt.Format("15:04:05"))
	}
	session := &modelk8s.DiagnosticSession{
		SessionKey: agentID, ClusterID: clusterID, AppName: appName, AgentID: agentID,
		Status: "active", UserID: userID, Username: username, StartAt: time.Now(),
	}
	if err := s.diagnosticRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("创建会话失败（可能已被占用）: %w", err)
	}
	return session, nil
}

// UpdateSessionIOLog 更新会话 I/O 录制
func (s *DiagnosticService) UpdateSessionIOLog(id uint, ioLog string) {
	session, err := s.diagnosticRepo.FindSessionByID(id)
	if err != nil || session == nil {
		return
	}
	session.IOLog = ioLog
	if err := s.diagnosticRepo.UpdateSession(session); err != nil {
		logger.Warn("更新会话 I/O 录制失败", zap.Uint64("sessionId", uint64(id)), zap.Error(err))
	}
}

// GetTunnelURL 获取平台访问 tunnel-server 的 HTTP 地址（探测 API 用）
func (s *DiagnosticService) GetTunnelURL(clusterID string) string {
	return s.getTunnelURL(clusterID)
}

// GetTunnelSessionURL 获取会话 WS 基地址（ws(s)://host:7777）
func (s *DiagnosticService) GetTunnelSessionURL(clusterID string) string {
	return s.getTunnelSessionURL(clusterID)
}

// DialTunnelSession 与 agent 建立诊断 WS 会话（导出供 WS controller 使用）
func (s *DiagnosticService) DialTunnelSession(sessionURL, agentID string) (*websocket.Conn, error) {
	return s.tunnelClient.DialAgentSession(sessionURL, agentID)
}

// CleanupExpiredExecutions 清理保留期外的执行记录（days 天，0 表示禁用）
func (s *DiagnosticService) CleanupExpiredExecutions(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	return s.diagnosticRepo.DeleteExecutionsBefore(time.Now().AddDate(0, 0, -days))
}

// ========== 接入与治理（管理员） ==========

// GetCommandOverrides 命令风险覆盖表
func (s *DiagnosticService) GetCommandOverrides() ([]modelk8s.DiagnosticCommandOverride, error) {
	return s.diagnosticRepo.ListCommandOverrides()
}

// CommandOverrideInput 覆盖表保存入参
type CommandOverrideInput struct {
	Command     string `json:"command" binding:"required"`
	RiskLevel   string `json:"riskLevel" binding:"required"` // L0-L5 或 disabled
	Description string `json:"description"`
	UpdatedBy   string `json:"updatedBy"`
}

// SaveCommandOverride 新增/更新命令风险覆盖（返回合并后的最新目录，便于前端即时生效）
func (s *DiagnosticService) SaveCommandOverride(input *CommandOverrideInput) error {
	level := strings.ToLower(strings.TrimSpace(input.RiskLevel))
	valid := map[string]bool{"l0": true, "l1": true, "l2": true, "l3": true, "l4": true, "l5": true, "disabled": true}
	if !valid[level] {
		return fmt.Errorf("无效的风险级: %s（允许 L0-L5 / disabled）", input.RiskLevel)
	}
	if strings.TrimSpace(input.Command) == "" {
		return fmt.Errorf("命令不能为空")
	}
	return s.diagnosticRepo.UpsertCommandOverride(&modelk8s.DiagnosticCommandOverride{
		Command:     strings.ToLower(strings.TrimSpace(input.Command)),
		RiskLevel:   level,
		Description: input.Description,
		UpdatedBy:   input.UpdatedBy,
	})
}

// DeleteCommandOverrideById 删除命令风险覆盖（恢复内置默认分级）
func (s *DiagnosticService) DeleteCommandOverrideById(id uint) error {
	return s.diagnosticRepo.DeleteCommandOverride(id)
}
