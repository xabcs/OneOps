package config

import (
	"fmt"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	JWT        JWTConfig        `yaml:"jwt"`
	Encryption EncryptionConfig `yaml:"encryption"`
	Log        LogConfig        `yaml:"log"`
	CORS       CORSConfig       `yaml:"cors"`
	App        AppConfig        `yaml:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `yaml:"port"`          // 监听端口
	Mode         string `yaml:"mode"`          // 运行模式: debug, release, test
	ReadTimeout  int    `yaml:"read_timeout"`  // 读超时(秒)
	WriteTimeout int    `yaml:"write_timeout"` // 写超时(秒)
	ExternalURL  string `yaml:"external_url"`  // 外部访问地址（Agent心跳上报）
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return ":" + c.Port
}

// GetExternalURL 获取外部访问地址（用于Agent心跳）
func (c *ServerConfig) GetExternalURL() string {
	if c.ExternalURL != "" {
		return c.ExternalURL
	}
	// 默认使用 localhost:port
	return "http://localhost:" + c.Port
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
	// interpolateParams=true：带参数的查询由驱动在客户端转义拼接后走单次
	// COM_QUERY（1 个网络往返）；默认模式下走 COM_STMT_PREPARE + COM_STMT_EXECUTE
	// 两个往返，远程数据库下每条带参查询多付一次 RTT。charset=utf8mb4 已保证
	// 多字节字符转义正确，由驱动完成的转义无注入风险
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `yaml:"host"`           // Redis主机
	Port         string `yaml:"port"`           // Redis端口
	Password     string `yaml:"password"`       // Redis密码
	DB           int    `yaml:"db"`             // Redis数据库编号
	PoolSize     int    `yaml:"pool_size"`      // 连接池大小
	MinIdleConns int    `yaml:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  int    `yaml:"dial_timeout"`   // 连接超时时间(秒)
	ReadTimeout  int    `yaml:"read_timeout"`   // 读取超时时间(秒)
	WriteTimeout int    `yaml:"write_timeout"`  // 写入超时时间(秒)
	Enabled      bool   `yaml:"enabled"`        // 是否启用Redis
}

// GetAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return c.Host + ":" + c.Port
}

// IsEnabled 检查Redis是否启用
func (c *RedisConfig) IsEnabled() bool {
	return c.Enabled && c.Host != ""
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`      // JWT密钥
	ExpireTime int    `yaml:"expire_time"` // 过期时间(小时)
	Issuer     string `yaml:"issuer"`      // 签发者
}

// EncryptionConfig 数据加密配置
type EncryptionConfig struct {
	Key string `yaml:"key"` // 数据加密密钥（Base64编码的32字节密钥）
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
