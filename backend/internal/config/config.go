// Package config 负责配置结构定义与加载。
// 使用 Viper 读取 configs/config.yaml，并支持环境变量覆盖（教学简化：通过 viper.AutomaticEnv）。
package config

import (
	"strconv"

	"github.com/spf13/viper"
)

// Config 应用全局配置，字段与 configs/config.yaml 一一对应（mapstructure tag）。
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口
	Mode string `mapstructure:"mode"` // gin 运行模式：debug / release / test
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 空闲连接池上限
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 最大打开连接数
	LogLevel     string `mapstructure:"log_level"`     // GORM 日志级别
}

// DSN 拼接 GORM MySQL DSN（parseTime 让 DATETIME 自动映射 time.Time，
// loc=Local 保证时区正确）
func (m *MySQLConfig) DSN() string {
	return m.Username + ":" + m.Password +
		"@tcp(" + m.Host + ":" + itoa(m.Port) + ")/" + m.Database +
		"?charset=" + m.Charset + "&parseTime=True&loc=Local"
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回 "host:port" 形式地址
func (r *RedisConfig) Addr() string {
	return r.Host + ":" + itoa(r.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`       // 签名密钥（生产必须更换并通过环境变量注入）
	ExpireHours int    `mapstructure:"expire_hours"` // Token 有效期（小时）
	Issuer      string `mapstructure:"issuer"`       // 签发者
}

// OpenAIConfig 大模型配置（Eino ChatModel 使用，兼容 OpenAI 协议的任意服务）
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// RateLimitConfig 接口限流配置
type RateLimitConfig struct {
	Enabled        bool `mapstructure:"enabled"`          // 是否开启限流
	LimitPerMinute int  `mapstructure:"limit_per_minute"` // 每 IP 每分钟最大请求数
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`    // debug / info / warn / error
	Encoding string `mapstructure:"encoding"` // json / console
	Output   string `mapstructure:"output"`   // stdout / file
}

// Load 从指定路径加载配置文件并反序列化为 Config。
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	// 支持环境变量覆盖（优先级：环境变量 > 配置文件）
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// itoa 简单封装 strconv.Itoa，供 DSN / Addr 拼接端口使用
func itoa(n int) string {
	return strconv.Itoa(n)
}
