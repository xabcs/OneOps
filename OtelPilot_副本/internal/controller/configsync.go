package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"otelpilot/internal/config"
)

// ConfigSyncer 周期同步组件命名空间下 ConfigMap / 授权 Secret 至内存快照。
// 不参与 Leader 选举：每个副本各自同步，保证各自 Webhook 使用最新配置。
type ConfigSyncer struct {
	Reader         client.Reader // 直连 API（绕过缓存），避免权限扩大
	Store          *config.Store
	Namespace      string
	ConfigMapName  string
	AuthSecretName string
	Interval       time.Duration
	Logger         logr.Logger
}

// NewConfigSyncer 创建配置同步器，默认 15s 同步一次。
func NewConfigSyncer(reader client.Reader, store *config.Store, namespace, configMapName, authSecretName string, logger logr.Logger) *ConfigSyncer {
	return &ConfigSyncer{
		Reader:         reader,
		Store:          store,
		Namespace:      namespace,
		ConfigMapName:  configMapName,
		AuthSecretName: authSecretName,
		Interval:       15 * time.Second,
		Logger:         logger,
	}
}

// Start 实现 manager.Runnable。
func (s *ConfigSyncer) Start(ctx context.Context) error {
	s.syncOnce(ctx)
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

// NeedLeaderElection 所有副本均需运行（各自 Webhook 依赖本地配置快照）。
func (s *ConfigSyncer) NeedLeaderElection() bool { return false }

func (s *ConfigSyncer) syncOnce(ctx context.Context) {
	cm := &corev1.ConfigMap{}
	cmFound := true
	if err := s.Reader.Get(ctx, client.ObjectKey{Namespace: s.Namespace, Name: s.ConfigMapName}, cm); err != nil {
		if !apierrors.IsNotFound(err) {
			s.Logger.Error(err, "failed to read configmap", "name", s.ConfigMapName)
			return
		}
		cmFound = false
	}

	headers := ""
	if s.AuthSecretName != "" {
		auth := &corev1.Secret{}
		if err := s.Reader.Get(ctx, client.ObjectKey{Namespace: s.Namespace, Name: s.AuthSecretName}, auth); err != nil {
			if !apierrors.IsNotFound(err) {
				s.Logger.Error(err, "failed to read auth secret", "name", s.AuthSecretName)
				return
			}
		} else {
			headers = string(auth.Data[config.AuthSecretKey])
		}
	}

	var cmOrNil *corev1.ConfigMap
	if cmFound {
		cmOrNil = cm
	}
	snap := config.SnapshotFromConfigMap(cmOrNil, headers)
	if snap.Hash() != s.Store.Get().Hash() {
		s.Store.Set(snap)
		s.Logger.Info("config synced",
			"agentVersion", snap.AgentVersion,
			"endpoint", snap.ExporterEndpoint,
			"authEnabled", snap.AuthHeaders != "",
		)
	}
}
