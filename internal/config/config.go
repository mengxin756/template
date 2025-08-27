package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Environment 环境类型
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Address         string        `mapstructure:"address"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	EnableCORS      bool          `mapstructure:"enable_cors"`
	EnableMetrics   bool          `mapstructure:"enable_metrics"`
	EnableHealth    bool          `mapstructure:"enable_health"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level        string `mapstructure:"level"`
	Encoding     string `mapstructure:"encoding"`
	Development  bool   `mapstructure:"development"`
	Service      string `mapstructure:"service"`
	Output       string `mapstructure:"output"`
	MaxSize      int    `mapstructure:"max_size"`
	MaxAge       int    `mapstructure:"max_age"`
	MaxBackups   int    `mapstructure:"max_backups"`
	Compress     bool   `mapstructure:"compress"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver      string        `mapstructure:"driver"`
	DSN         string        `mapstructure:"dsn"`
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Username    string        `mapstructure:"username"`
	Password    string        `mapstructure:"password"`
	Database    string        `mapstructure:"database"`
	Charset     string        `mapstructure:"charset"`
	MaxOpen     int           `mapstructure:"max_open"`
	MaxIdle     int           `mapstructure:"max_idle"`
	MaxLifetime time.Duration `mapstructure:"max_lifetime"`
	AutoMigrate bool          `mapstructure:"auto_migrate"`
	LogLevel    string        `mapstructure:"log_level"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AsynqConfig Asynq 任务队列配置
type AsynqConfig struct {
	RedisAddr     string        `mapstructure:"redis_addr"`
	RedisPassword string        `mapstructure:"redis_password"`
	RedisDB       int           `mapstructure:"redis_db"`
	Concurrency   int           `mapstructure:"concurrency"`
	Queues        []string      `mapstructure:"queues"`
	StrictPriority bool         `mapstructure:"strict_priority"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers       []string      `mapstructure:"brokers"`
	GroupID       string        `mapstructure:"group_id"`
	Topic         string        `mapstructure:"topic"`
	Partitions    int           `mapstructure:"partitions"`
	Replicas      int           `mapstructure:"replicas"`
	Retention     time.Duration `mapstructure:"retention"`
	MaxMessageBytes int         `mapstructure:"max_message_bytes"`
}

// Config 应用配置
type Config struct {
	Environment Environment   `mapstructure:"environment"`
	Service     string       `mapstructure:"service"`
	Version     string       `mapstructure:"version"`
	HTTP        HTTPConfig   `mapstructure:"http"`
	Log         LogConfig    `mapstructure:"log"`
	DB          DBConfig     `mapstructure:"db"`
	Redis       RedisConfig  `mapstructure:"redis"`
	Asynq       AsynqConfig  `mapstructure:"asynq"`
	Kafka       KafkaConfig  `mapstructure:"kafka"`
}

// Load 加载配置
func Load() (*Config, error) {
	v := viper.New()
	
	// 设置配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")
	
	// 环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// 设置默认值
	setDefaults(v)
	
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}
	
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	
	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 基础配置
	v.SetDefault("environment", "development")
	v.SetDefault("service", "classic-api")
	v.SetDefault("version", "1.0.0")
	
	// HTTP 配置
	v.SetDefault("http.address", ":8080")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "20s")
	v.SetDefault("http.idle_timeout", "60s")
	v.SetDefault("http.max_header_bytes", 1048576)
	v.SetDefault("http.enable_cors", true)
	v.SetDefault("http.enable_metrics", true)
	v.SetDefault("http.enable_health", true)
	
	// 日志配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")
	v.SetDefault("log.development", false)
	v.SetDefault("log.service", "classic-api")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.compress", true)
	
	// 数据库配置
	v.SetDefault("db.driver", "sqlite")
	v.SetDefault("db.dsn", "file:ent?mode=memory&cache=shared&_fk=1")
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 3306)
	v.SetDefault("db.charset", "utf8mb4")
	v.SetDefault("db.max_open", 100)
	v.SetDefault("db.max_idle", 10)
	v.SetDefault("db.max_lifetime", "1h")
	v.SetDefault("db.auto_migrate", true)
	v.SetDefault("db.log_level", "warn")
	
	// Redis 配置
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")
	
	// Asynq 配置
	v.SetDefault("asynq.redis_addr", "localhost:6379")
	v.SetDefault("asynq.redis_db", 1)
	v.SetDefault("asynq.concurrency", 10)
	v.SetDefault("asynq.queues", []string{"default", "critical"})
	v.SetDefault("asynq.strict_priority", false)
	v.SetDefault("asynq.shutdown_timeout", "30s")
	
	// Kafka 配置
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.group_id", "classic-consumer")
	v.SetDefault("kafka.topic", "classic-events")
	v.SetDefault("kafka.partitions", 3)
	v.SetDefault("kafka.replicas", 1)
	v.SetDefault("kafka.retention", "168h")
	v.SetDefault("kafka.max_message_bytes", 1048576)
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证环境
	switch c.Environment {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}
	
	// 验证 HTTP 配置
	if c.HTTP.Address == "" {
		return fmt.Errorf("http address is required")
	}
	
	// 验证日志配置
	if c.Log.Level == "" {
		return fmt.Errorf("log level is required")
	}
	
	// 验证数据库配置
	if c.DB.Driver == "" {
		return fmt.Errorf("db driver is required")
	}
	
	return nil
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment
}

// IsStaging 是否为预发布环境
func (c *Config) IsStaging() bool {
	return c.Environment == EnvStaging
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction
}


