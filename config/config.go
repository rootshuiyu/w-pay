package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置，敏感项通过环境变量注入，禁止硬编码
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Order     OrderConfig     `yaml:"order"`
	Pay       PayConfig       `yaml:"pay"`
	Log       LogConfig       `yaml:"log"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Callback  CallbackConfig  `yaml:"callback"`
	Cache     CacheConfig     `yaml:"cache"`
	Security  SecurityConfig  `yaml:"security"`
}

type ServerConfig struct {
	Port       int    `yaml:"port"`
	Mode       string `yaml:"mode"`
	TLSEnabled bool   `yaml:"tls_enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
}

// DatabaseConfig PostgreSQL 连接配置
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	SSLMode      string `yaml:"ssl_mode"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type OrderConfig struct {
	TimeoutMinutes int `yaml:"timeout_minutes"`
}

// PayConfig 手机 H5/WAP 支付场景参数（微信 H5 必填 h5_info）
type PayConfig struct {
	H5AppName string `yaml:"h5_app_name"`
	H5AppURL  string `yaml:"h5_app_url"`
}

type LogConfig struct {
	Level                    string `yaml:"level"`
	SensitiveRetentionMonths int    `yaml:"sensitive_retention_months"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

type CallbackConfig struct {
	IdempotentExpireHours int `yaml:"idempotent_expire_hours"`
}

type CacheConfig struct {
	ChannelTTLHours       int `yaml:"channel_ttl_hours"`
	ChannelTTLJitterHours int `yaml:"channel_ttl_jitter_hours"`
}

// SecurityConfig 访问来源限制，逗号分隔的 IP 或 CIDR；留空表示不限制
type SecurityConfig struct {
	AdminIPWhitelist    string `yaml:"admin_ip_whitelist"`
	CallbackIPWhitelist string `yaml:"callback_ip_whitelist"`
	PayIPWhitelist      string `yaml:"pay_ip_whitelist"`
	// TrustedProxies 可信反向代理；留空表示不信任任何代理（ClientIP 取 TCP 源地址），
	// 防止客户端伪造 X-Forwarded-For 绕过 IP 白名单
	TrustedProxies string `yaml:"trusted_proxies"`
}

var Global *Config

var envVarPattern = regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

// Load 加载 yaml 配置并解析 ${ENV:default} 占位符
func Load(env string) (*Config, error) {
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = "dev"
	}

	filename, err := resolveConfigFile(env + ".yaml")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", filename, err)
	}

	expanded := expandEnvVars(string(data))
	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	Global = cfg
	return cfg, nil
}

// resolveConfigFile 定位 config/{env}.yaml：优先 WPAY_CONFIG_DIR，否则从当前目录逐级向上查找，
// 使 go test 在 tests/e2e 等子目录下运行时同样能读到配置
func resolveConfigFile(name string) (string, error) {
	if dir := os.Getenv("WPAY_CONFIG_DIR"); dir != "" {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("config file not found: %s: %w", path, err)
		}
		return path, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		path := filepath.Join(dir, "config", name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("config file config/%s not found from %s upwards", name, dir)
		}
		dir = parent
	}
}

func expandEnvVars(content string) string {
	return envVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := envVarPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		envKey := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}
		if val := os.Getenv(envKey); val != "" {
			return val
		}
		return defaultVal
	})
}

// DSN 生成 PostgreSQL 连接串
func (c *DatabaseConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.Host, c.Port, c.User, c.Password, c.Database, sslMode)
}

// IsProd 是否生产环境
func IsProd() bool {
	if Global == nil {
		return false
	}
	return strings.EqualFold(Global.Server.Mode, "release")
}
