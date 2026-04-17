package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从配置文件加载配置
// 支持环境变量覆盖配置文件中的值，格式：${ENV_VAR}
func LoadConfig(configPath ...string) (*Config, error) {
	// 1. 确定配置文件路径
	cfgFile := getConfigFilePath(configPath...)

	// 2. 读取配置文件
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败 [%s]: %w", cfgFile, err)
	}

	// 3. 环境变量替换
	data = []byte(os.ExpandEnv(string(data)))

	// 4. 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 5. 设置默认值
	setDefaults(&config)

	// 6. 验证配置
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// getConfigFilePath 获取配置文件路径
// 优先级：参数指定 > 环境变量 > 当前目录 > /etc/oneops
func getConfigFilePath(customPaths ...string) string {
	// 1. 检查参数指定的路径
	for _, path := range customPaths {
		if path != "" && fileExists(path) {
			return path
		}
	}

	// 2. 检查环境变量指定的路径
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		if fileExists(envPath) {
			return envPath
		}
	}

	// 3. 检查当前目录的config.yaml
	currentDir := "."
	if fileExists(filepath.Join(currentDir, "config.yaml")) {
		return filepath.Join(currentDir, "config.yaml")
	}

	// 4. 检查config目录下的config.yaml
	configDir := "config"
	if fileExists(filepath.Join(configDir, "config.yaml")) {
		return filepath.Join(configDir, "config.yaml")
	}

	// 5. 检查/etc/oneops/config.yaml
	etcPath := "/etc/oneops/config.yaml"
	if fileExists(etcPath) {
		return etcPath
	}

	// 6. 默认返回config.yaml（会在后续读取时报错）
	return filepath.Join(configDir, "config.yaml")
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// setDefaults 设置配置默认值
func setDefaults(cfg *Config) {
	// 服务器默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8082"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 60
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 60
	}

	// 数据库默认值
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == "" {
		cfg.Database.Port = "3306"
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "root"
	}
	if cfg.Database.DBName == "" {
		cfg.Database.DBName = "oneops"
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.LogLevel == "" {
		cfg.Database.LogLevel = "warn"
	}

	// JWT默认值
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 24
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "oneops"
	}

	// 日志默认值
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Filename == "" {
		cfg.Log.Filename = "logs/app.log"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 100
	}
	if cfg.Log.MaxBackups == 0 {
		cfg.Log.MaxBackups = 3
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 28
	}

	// CORS默认值
	if len(cfg.CORS.AllowOrigins) == 0 {
		cfg.CORS.AllowOrigins = []string{"http://localhost:5173"}
	}
	if len(cfg.CORS.AllowMethods) == 0 {
		cfg.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.CORS.AllowHeaders) == 0 {
		cfg.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	}
	if cfg.CORS.MaxAge == 0 {
		cfg.CORS.MaxAge = 86400
	}

	// 应用默认值
	if cfg.App.Name == "" {
		cfg.App.Name = "OneOps"
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "1.0.0"
	}
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}
}

// validate 验证配置
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}

	// 验证数据库配置
	if cfg.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("数据库名称不能为空")
	}

	// 验证JWT配置
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "${JWT_SECRET}" {
		return fmt.Errorf("JWT密钥不能为空，请设置JWT_SECRET环境变量")
	}
	if strings.Contains(cfg.JWT.Secret, "change-in-production") {
		return fmt.Errorf("JWT密钥使用了默认值，生产环境请修改")
	}

	// 验证日志配置
	if cfg.Log.Filename == "" {
		return fmt.Errorf("日志文件路径不能为空")
	}

	return nil
}
