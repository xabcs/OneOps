// Package config 维护探针注入的运行时配置快照，
// 默认值内置，可通过组件命名空间的 ConfigMap / Secret 热更新。
package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
)

// ConfigMap / Secret 中的配置键。
const (
	KeyAgentVersion     = "agent.version"
	KeyAgentDownloadURL = "agent.download-url"
	KeyInitImage        = "init.image"
	KeyExporterEndpoint = "exporter.endpoint"
	KeyExporterProtocol = "exporter.protocol"
	KeyTracesExporter   = "traces.exporter"
	KeyMetricsExporter  = "metrics.exporter"
	KeyLogsExporter     = "logs.exporter"
	KeyPropagators      = "propagators"
	KeyResourceAttrs    = "resource.attributes"

	// AuthSecretKey 授权 Secret 中存放 OTEL_EXPORTER_OTLP_HEADERS 原始值的键。
	AuthSecretKey = "headers"
)

// Snapshot 注入动作依赖的配置快照（只读，整体原子替换）。
type Snapshot struct {
	// AgentVersion 默认探针版本，可被 Pod 标签 otel.pilot.agent.version 覆盖。
	AgentVersion string `json:"agentVersion"`
	// AgentDownloadURL 探针下载地址模板，{version} 为版本占位符。
	AgentDownloadURL string `json:"agentDownloadUrl"`
	// InitImage 拉取探针的 Init 容器镜像。
	InitImage string `json:"initImage"`
	// ExporterEndpoint OTLP 上报地址。
	ExporterEndpoint string `json:"exporterEndpoint"`
	// ExporterProtocol OTLP 传输协议：grpc 或 http/protobuf。
	ExporterProtocol string `json:"exporterProtocol"`
	// Traces/Metrics/LogsExporter 默认 none 表示关闭。
	TracesExporter  string `json:"tracesExporter"`
	MetricsExporter string `json:"metricsExporter"`
	LogsExporter    string `json:"logsExporter"`
	// Propagators 上下文传播器，默认 W3C：tracecontext,baggage。
	Propagators string `json:"propagators"`
	// ResourceAttributes 追加到 OTEL_RESOURCE_ATTRIBUTES 的静态资源属性。
	ResourceAttributes map[string]string `json:"resourceAttributes"`
	// AuthHeaders OTEL_EXPORTER_OTLP_HEADERS 原始值（如 Authorization=xxx），为空不注入。
	AuthHeaders string `json:"-"`
}

// DefaultSnapshot 内置默认配置。
func DefaultSnapshot() *Snapshot {
	return &Snapshot{
		AgentVersion:     "2.31.1",
		AgentDownloadURL: "https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v{version}/opentelemetry-javaagent.jar",
		InitImage:        "curlimages/curl:8.11.1",
		ExporterEndpoint: "http://otel-collector.observability.svc:4317",
		ExporterProtocol: "grpc",
		TracesExporter:   "otlp",
		MetricsExporter:  "none",
		LogsExporter:     "none",
		Propagators:      "tracecontext,baggage",
	}
}

// SnapshotFromConfigMap 基于默认值叠加 ConfigMap 覆盖项构建快照，cm 为 nil 时返回默认快照。
func SnapshotFromConfigMap(cm *corev1.ConfigMap, authHeaders string) *Snapshot {
	snap := DefaultSnapshot()
	snap.AuthHeaders = authHeaders
	if cm == nil {
		return snap
	}
	data := cm.Data
	override := func(key string, dst *string) {
		if v := strings.TrimSpace(data[key]); v != "" {
			*dst = v
		}
	}
	override(KeyAgentVersion, &snap.AgentVersion)
	override(KeyAgentDownloadURL, &snap.AgentDownloadURL)
	override(KeyInitImage, &snap.InitImage)
	override(KeyExporterEndpoint, &snap.ExporterEndpoint)
	override(KeyExporterProtocol, &snap.ExporterProtocol)
	override(KeyTracesExporter, &snap.TracesExporter)
	override(KeyMetricsExporter, &snap.MetricsExporter)
	override(KeyLogsExporter, &snap.LogsExporter)
	override(KeyPropagators, &snap.Propagators)
	if raw := strings.TrimSpace(data[KeyResourceAttrs]); raw != "" {
		attrs := map[string]string{}
		for _, pair := range strings.Split(raw, ",") {
			k, v, ok := strings.Cut(strings.TrimSpace(pair), "=")
			if ok && k != "" {
				attrs[k] = v
			}
		}
		snap.ResourceAttributes = attrs
	}
	return snap
}

// Hash 生成快照内容摘要，用于变更检测。
func (s *Snapshot) Hash() string {
	type plain struct {
		Snapshot
		ResourceAttributesSorted [][2]string
	}
	p := plain{Snapshot: *s}
	for k, v := range s.ResourceAttributes {
		p.ResourceAttributesSorted = append(p.ResourceAttributesSorted, [2]string{k, v})
	}
	sort.Slice(p.ResourceAttributesSorted, func(i, j int) bool {
		return p.ResourceAttributesSorted[i][0] < p.ResourceAttributesSorted[j][0]
	})
	b, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

// Store 配置快照的原子存储，Webhook 每次请求读取最新快照。
type Store struct {
	v atomic.Pointer[Snapshot]
}

// NewStore 基于默认配置创建存储。
func NewStore() *Store {
	s := &Store{}
	s.v.Store(DefaultSnapshot())
	return s
}

// Get 获取当前配置快照（只读，调用方不得修改）。
func (s *Store) Get() *Snapshot {
	return s.v.Load()
}

// Set 原子替换配置快照。
func (s *Store) Set(snap *Snapshot) {
	s.v.Store(snap)
}
