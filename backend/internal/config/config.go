package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	MySQL   MySQLConfig   `yaml:"mysql"`
	Redis   RedisConfig   `yaml:"redis"`
	Storage StorageConfig `yaml:"storage"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type StorageConfig struct {
	RootPath   string `yaml:"rootPath"`
	QuotaBytes int64  `yaml:"quotaBytes"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(&cfg)
	if err := overrideFromEnv(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	if cfg.MySQL.Port == 0 {
		cfg.MySQL.Port = 3306
	}

	if cfg.Redis.DB == 0 {
		cfg.Redis.DB = 0
	}

	if strings.TrimSpace(cfg.Storage.RootPath) == "" {
		cfg.Storage.RootPath = filepath.Join("storage", "course")
	}
	if cfg.Storage.QuotaBytes == 0 {
		cfg.Storage.QuotaBytes = 1 << 30
	}
}

func overrideFromEnv(cfg *Config) error {
	if v := os.Getenv("APP_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("APP_SERVER_PORT"); v != "" {
		value, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("invalid APP_SERVER_PORT: %w", err)
		}
		cfg.Server.Port = value
	}
	if v := os.Getenv("APP_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}

	if v := os.Getenv("APP_MYSQL_HOST"); v != "" {
		cfg.MySQL.Host = v
	}
	if v := os.Getenv("APP_MYSQL_PORT"); v != "" {
		value, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("invalid APP_MYSQL_PORT: %w", err)
		}
		cfg.MySQL.Port = value
	}
	if v := os.Getenv("APP_MYSQL_USERNAME"); v != "" {
		cfg.MySQL.Username = v
	}
	if v := os.Getenv("APP_MYSQL_PASSWORD"); v != "" {
		cfg.MySQL.Password = v
	}
	if v := os.Getenv("APP_MYSQL_DATABASE"); v != "" {
		cfg.MySQL.Database = v
	}

	if v := os.Getenv("APP_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("APP_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("APP_REDIS_DB"); v != "" {
		value, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("invalid APP_REDIS_DB: %w", err)
		}
		cfg.Redis.DB = value
	}
	if v := os.Getenv("APP_STORAGE_ROOT_PATH"); v != "" {
		cfg.Storage.RootPath = v
	}
	if v := os.Getenv("APP_STORAGE_QUOTA_BYTES"); v != "" {
		value, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid APP_STORAGE_QUOTA_BYTES: %w", err)
		}
		cfg.Storage.QuotaBytes = value
	}

	return nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Server.Host) == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if strings.TrimSpace(c.MySQL.Host) == "" {
		return fmt.Errorf("mysql.host is required")
	}
	if c.MySQL.Port <= 0 {
		return fmt.Errorf("mysql.port must be positive")
	}
	if strings.TrimSpace(c.MySQL.Username) == "" {
		return fmt.Errorf("mysql.username is required")
	}
	if strings.TrimSpace(c.MySQL.Database) == "" {
		return fmt.Errorf("mysql.database is required")
	}
	if strings.TrimSpace(c.Redis.Addr) == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if strings.TrimSpace(c.Storage.RootPath) == "" {
		return fmt.Errorf("storage.rootPath is required")
	}
	if c.Storage.QuotaBytes < 0 {
		return fmt.Errorf("storage.quotaBytes must be non-negative")
	}
	return nil
}

func DefaultMySQLMaxOpenConns() int { return 20 }
func DefaultMySQLMaxIdleConns() int { return 10 }
func DefaultMySQLConnMaxLifetime() time.Duration { return 5 * time.Minute }
func DefaultMySQLCharset() string { return "utf8mb4" }
func DefaultMySQLLoc() string { return "Local" }
func DefaultMySQLParseTime() bool { return true }
func DefaultRedisPoolSize() int { return 10 }
func DefaultRedisMinIdleConns() int { return 2 }
func DefaultRedisDialTimeout() time.Duration { return 5 * time.Second }
func DefaultRedisReadTimeout() time.Duration { return 3 * time.Second }
func DefaultRedisWriteTimeout() time.Duration { return 3 * time.Second }
func DefaultRedisConnMaxIdleTime() time.Duration { return 5 * time.Minute }
