# 微服务与单体双模式架构改进方案

## 概述

本文档描述如何将现有的单体架构项目改造为同时支持**单体模式**和**微服务模式**的通用模板，支持服务注册中心、配置中心等微服务特性。

---

## 核心设计理念

```
┌─────────────────────────────────────────────────────────────┐
│                    同一套代码库                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   ┌──────────────┐          ┌──────────────┐               │
│   │   Monolithic │          │ Microservice │               │
│   │    Mode      │          │    Mode      │               │
│   └──────┬───────┘          └──────┬───────┘               │
│          │                         │                        │
│          ▼                         ▼                        │
│   ┌──────────────┐          ┌──────────────┐               │
│   │  Local DB    │          │  Consul SD   │               │
│   │  File Config │          │ Consul Config│               │
│   └──────┬───────┘          └──────┬───────┘               │
│          │                         │                        │
│          └──────────┬──────────────┘                        │
│                     ▼                                       │
│         ┌─────────────────────┐                            │
│         │  Clean Architecture │                            │
│         │   (不变的业务逻辑)   │                            │
│         └─────────────────────┘                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| 注册中心 | **Consul** | 成熟、功能齐全，支持服务发现、配置中心、健康检查 |
| 配置中心 | **Consul KV** | 使用 Consul 的 KV 存储作为配置中心 |
| 服务通信 | **HTTP + gRPC** | 同时支持两者，灵活切换 |
| 熔断器 | **Sentinel** | 调用容错保护 |
| 限流 | **Token Bucket / Leaky Bucket** | 防止服务过载 |
| 分布式追踪 | **Jaeger** | 链路追踪与分析 |

---

## 项目结构调整

```
template/
├── internal/
│   ├── config/                    # 配置管理（扩展）
│   │   ├── config.go             # 现有配置
│   │   ├── deployment.go         # 新增：部署模式配置
│   │   ├── remote.go             # 新增：远程配置接口
│   │   └── consul.go             # 新增：Consul 客户端封装
│   ├── discovery/                 # 新增：服务发现层
│   │   ├── registry.go           # 注册中心接口
│   │   ├── consul_registry.go    # Consul 实现（微服务模式）
│   │   └── local_registry.go     # 本地实现（单体模式）
│   ├── client/                    # 新增：服务调用客户端
│   │   ├── client.go             # 客户端接口
│   │   ├── http_client.go        # HTTP 客户端实现
│   │   └── grpc_client.go        # gRPC 客户端实现
│   ├── middleware/                # 新增：微服务中间件
│   │   ├── tracing.go            # 分布式追踪
│   │   ├── circuit_breaker.go    # 熔断器
│   │   ├── rate_limiter.go       # 限流
│   │   └── retry.go              # 重试机制
│   ├── domain/                    # 领域层（保持不变）
│   ├── handler/                   # Handler 层（保持不变）
│   ├── service/                   # Service 层（保持不变）
│   ├── repository/                # Repository 层（保持不变）
│   ├── data/                      # 数据层（保持不变）
│   └── [其他现有层...]
├── api/
│   ├── http/                      # 现有 HTTP API
│   └── grpc/                      # 新增：gRPC 定义
│       └── user.proto
├── config/
│   ├── config.yaml                # 单体模式配置
│   └── config.micro.yaml          # 微服务模式配置
├── deployments/
│   ├── docker-compose.yml         # 单体模式编排
│   ├── docker-compose.consul.yml # 微服务模式编排
│   └── kubernetes/                # K8s 部署配置
└── docs/
    ├── architecture.md            # 本文档
    └── migration-guide.md         # 迁移指南
```

---

## 一、配置系统设计

### 1. 部署模式配置

在 `internal/config/deployment.go` 中定义：

```go
package config

// DeploymentMode 部署模式
type DeploymentMode string

const (
    ModeMonolithic  DeploymentMode = "monolithic"  // 单体模式
    ModeMicroservice DeploymentMode = "microservice" // 微服务模式
)

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
    Enabled    bool   `mapstructure:"enabled"`     // 是否启用服务发现
    Backend    string `mapstructure:"backend"`     // 注册中心类型
    Address    string `mapstructure:"address"`     // 注册中心地址
    Token      string `mapstructure:"token"`       // 认证 Token
    HealthPath string `mapstructure:"health_path"` // 健康检查路径
    ServiceID  string `mapstructure:"service_id"`  // 服务 ID
}

// RemoteConfigConfig 远程配置中心配置
type RemoteConfigConfig struct {
    Enabled       bool   `mapstructure:"enabled"`          // 是否启用远程配置
    Backend       string `mapstructure:"backend"`          // 配置中心类型
    Namespace     string `mapstructure:"namespace"`        // 配置命名空间
    ConfigKey     string `mapstructure:"config_key"`      // 配置键
    WatchInterval int    `mapstructure:"watch_interval"`  // 监听间隔（秒）
}

// GRPCConfig gRPC 配置
type GRPCConfig struct {
    Enabled    bool          `mapstructure:"enabled"`       // 是否启用 gRPC
    Address    string        `mapstructure:"address"`       // gRPC 监听地址
    Timeout    time.Duration `mapstructure:"timeout"`       // 默认超时
    MaxRecvMsg int           `mapstructure:"max_recv_msg"`  // 最大接收消息大小
    MaxSendMsg int           `mapstructure:"max_send_msg"`  // 最大发送消息大小
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
    Enabled            bool          `mapstructure:"enabled"`
    MaxConcurrentCalls int           `mapstructure:"max_concurrent_calls"`
    Interval           time.Duration `mapstructure:"interval"`
    Timeout            time.Duration `mapstructure:"timeout"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
    Enabled  bool          `mapstructure:"enabled"`
    Requests int           `mapstructure:"requests"`
    Window   time.Duration `mapstructure:"window"`
}

// Config 扩展添加新配置项
type Config struct {
    // ... 现有配置
    Deployment     DeploymentMode         `mapstructure:"deployment"`
    Discovery      DiscoveryConfig        `mapstructure:"discovery"`
    RemoteConfig   RemoteConfigConfig     `mapstructure:"remote_config"`
    GRPC           GRPCConfig             `mapstructure:"grpc"`
    CircuitBreaker CircuitBreakerConfig   `mapstructure:"circuit_breaker"`
    RateLimit      RateLimitConfig        `mapstructure:"rate_limit"`
}

// IsMonolithic 是否为单体模式
func (c *Config) IsMonolithic() bool {
    return c.Deployment == ModeMonolithic
}

// IsMicroservice 是否为微服务模式
func (c *Config) IsMicroservice() bool {
    return c.Deployment == ModeMicroservice
}
```