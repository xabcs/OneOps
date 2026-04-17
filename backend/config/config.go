package config

import (
	"fmt"
	"log"
	"os"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	CORS     CORSConfig     `yaml:"cors"`
	App      AppConfig      `yaml:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `yaml:"port"`          // 监听端口
	Mode         string `yaml:"mode"`          // 运行模式: debug, release, test
	ReadTimeout  int    `yaml:"read_timeout"`  // 读超时(秒)
	WriteTimeout int    `yaml:"write_timeout"` // 写超时(秒)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return ":" + c.Port
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `yaml:"host"`           // 数据库主机
	Port         string `yaml:"port"`           // 数据库端口
	User         string `yaml:"user"`           // 数据库用户
	Password     string `yaml:"password"`       // 数据库密码
	DBName       string `yaml:"dbname"`         // 数据库名称
	MaxIdleConns int    `yaml:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `yaml:"max_open_conns"` // 最大打开连接数
	LogLevel     string `yaml:"log_level"`      // GORM日志级别
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`      // JWT密钥
	ExpireTime int    `yaml:"expire_time"` // 过期时间(小时)
	Issuer     string `yaml:"issuer"`      // 签发者
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error, fatal
	Filename   string `yaml:"filename"`    // 日志文件路径
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge     int    `yaml:"max_age"`     // 保留旧日志文件的最大天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`     // 允许的源地址
	AllowMethods     []string `yaml:"allow_methods"`     // 允许的HTTP方法
	AllowHeaders     []string `yaml:"allow_headers"`     // 允许的请求头
	ExposeHeaders    []string `yaml:"expose_headers"`    // 暴露的响应头
	MaxAge           int      `yaml:"max_age"`           // 预检请求缓存时间(秒)
	AllowCredentials bool     `yaml:"allow_credentials"` // 是否允许携带凭证
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `yaml:"name"`        // 应用名称
	Version     string `yaml:"version"`     // 应用版本
	Environment string `yaml:"environment"` // 运行环境
	Debug       bool   `yaml:"debug"`       // 是否开启调试
}

// GetConfig 获取配置（兼容旧代码，保持向后兼容）
// 已废弃：请使用 LoadConfig() 替代
func GetConfig() *Config {
	// 尝试加载配置文件，如果失败则返回默认配置
	cfg, err := LoadConfig()
	if err != nil {
		// 如果配置文件加载失败，记录警告日志并返回环境变量配置（向后兼容）
		log.Printf("警告: 配置文件加载失败，使用默认配置: %v", err)
		return &Config{
			Server: ServerConfig{
				Port:         getEnv("SERVER_PORT", "8082"),
				Mode:         getEnv("GIN_MODE", "debug"),
				ReadTimeout:  60,
				WriteTimeout: 60,
			},
			Database: DatabaseConfig{
				Host:         getEnv("DB_HOST", "localhost"),
				Port:         getEnv("DB_PORT", "3306"),
				User:         getEnv("DB_USER", "root"),
				Password:     getEnv("DB_PASSWORD", ""),
				DBName:       getEnv("DB_NAME", "oneops"),
				MaxIdleConns: 10,
				MaxOpenConns: 100,
				LogLevel:     "warn",
			},
			JWT: JWTConfig{
				Secret:     getEnv("JWT_SECRET", ""),
				ExpireTime: 24,
				Issuer:     "oneops",
			},
			Log: LogConfig{
				Level:      "info",
				Filename:   "logs/app.log",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
				Compress:   true,
			},
			// CORS配置 - 注意：当前CORS配置直接在middlewares/cors.go中硬编码
			// 这里的配置仅供参考，实际使用的是硬编码的AllowAllOrigins = true
			CORS: CORSConfig{
				AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
				AllowCredentials: true,
				MaxAge:           86400,
			},
			App: AppConfig{
				Name:        "OneOps",
				Version:     "1.0.0",
				Environment: "development",
				Debug:       true,
			},
		}
	}
	return cfg
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
