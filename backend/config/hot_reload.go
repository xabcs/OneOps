package config

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// HotReloadConfig 热加载配置
type HotReloadConfig struct {
	Enabled       bool          // 是否启用热加载
	CheckInterval time.Duration // 检查间隔（默认：30秒）
	OnReload      func(*Config) // 配置重新加载时的回调函数
}

// ConfigWatcher 配置监控器
type ConfigWatcher struct {
	configPath   string
	config       *Config
	hotConfig    HotReloadConfig
	lastModTime  time.Time
	mu           sync.RWMutex
	stopChan     chan struct{}
}

// NewConfigWatcher 创建配置监控器
func NewConfigWatcher(configPath string, hotConfig HotReloadConfig) *ConfigWatcher {
	return &ConfigWatcher{
		configPath: configPath,
		hotConfig:  hotConfig,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动配置监控
func (cw *ConfigWatcher) Start() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// 加载初始配置
	cfg, err := LoadConfig(cw.configPath)
	if err != nil {
		return fmt.Errorf("加载初始配置失败: %w", err)
	}
	cw.config = cfg

	// 获取文件修改时间
	info, err := os.Stat(cw.configPath)
	if err != nil {
		return fmt.Errorf("获取配置文件信息失败: %w", err)
	}
	cw.lastModTime = info.ModTime()

	// 启动监控协程
	if cw.hotConfig.Enabled {
		go cw.watchConfig()
		log.Printf("配置热加载已启动: config_path=%s, check_interval=%s",
			cw.configPath, cw.hotConfig.CheckInterval)
	}

	return nil
}

// watchConfig 监控配置文件变化
func (cw *ConfigWatcher) watchConfig() {
	ticker := time.NewTicker(cw.hotConfig.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cw.checkAndUpdate()
		case <-cw.stopChan:
			return
		}
	}
}

// checkAndUpdate 检查配置文件并更新
func (cw *ConfigWatcher) checkAndUpdate() {
	// 获取文件修改时间
	info, err := os.Stat(cw.configPath)
	if err != nil {
		log.Printf("获取配置文件信息失败: %v", err)
		return
	}

	// 检查是否有修改
	if info.ModTime().After(cw.lastModTime) {
		cw.mu.Lock()
		defer cw.mu.Unlock()

		// 重新加载配置
		newCfg, err := LoadConfig(cw.configPath)
		if err != nil {
			log.Printf("重新加载配置失败: %v", err)
			return
		}

		// 验证新配置
		if err := validate(newCfg); err != nil {
			log.Printf("新配置验证失败，不更新: %v", err)
			return
		}

		// 更新配置
		cw.config = newCfg
		cw.lastModTime = info.ModTime()

		log.Printf("配置已热加载: modified_time=%s", info.ModTime().Format(time.RFC3339))

		// 调用回调函数
		if cw.hotConfig.OnReload != nil {
			cw.hotConfig.OnReload(newCfg)
		}
	}
}

// GetConfig 获取当前配置
func (cw *ConfigWatcher) GetConfig() *Config {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

// Stop 停止监控
func (cw *ConfigWatcher) Stop() {
	close(cw.stopChan)
}

// GetConfigSummary 获取配置摘要（用于调试）
func GetConfigSummary(cfg *Config) map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		},
		"database": map[string]interface{}{
			"host":   cfg.Database.Host,
			"port":   cfg.Database.Port,
			"dbname": cfg.Database.DBName,
		},
		"redis": map[string]interface{}{
			"enabled": cfg.Redis.Enabled,
			"host":    cfg.Redis.Host,
			"port":    cfg.Redis.Port,
		},
		"app": map[string]interface{}{
			"name":        cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
		},
	}
}

// WatchConfig 监控配置文件变化的便捷函数
func WatchConfig(configPath string, checkInterval time.Duration, onReload func(*Config)) (*ConfigWatcher, error) {
	hotConfig := HotReloadConfig{
		Enabled:       true,
		CheckInterval: checkInterval,
		OnReload:      onReload,
	}

	watcher := NewConfigWatcher(configPath, hotConfig)
	if err := watcher.Start(); err != nil {
		return nil, err
	}

	return watcher, nil
}
