# Classic Go Project

一个基于 Clean Architecture 的经典 Go 项目，使用现代化的技术栈和最佳实践。

## 🏗️ 项目架构

### 目录结构
```
template/
├── cmd/                    # 主程序入口
│   └── api/              # API 服务入口
├── internal/              # 核心业务逻辑（不可被外部依赖）
│   ├── config/           # 配置管理
│   ├── domain/           # 领域对象、实体、接口
│   ├── service/          # 业务用例实现
│   ├── repository/       # 数据访问层
│   ├── handler/          # HTTP 处理器
│   ├── data/             # 数据层
│   │   └── ent/         # Ent ORM
│   └── server/           # 服务器配置
├── pkg/                   # 公共工具库
│   ├── errors/           # 统一错误处理
│   ├── logger/           # 结构化日志
│   └── response/         # HTTP 响应格式
├── config/                # 配置文件
├── api/                   # API 定义
└── test/                  # 测试文件
```

### 技术栈
- **Web 框架**: Gin
- **ORM**: Ent
- **依赖注入**: Google Wire
- **日志**: Zerolog
- **配置管理**: Viper
- **数据库**: SQLite (支持 MySQL/PostgreSQL)
- **任务队列**: Asynq (计划中)
- **消息队列**: Kafka (计划中)

## 🚀 快速开始

### 环境要求
- Go 1.21+
- SQLite (开发环境)

### 安装依赖
```bash
go mod tidy
```

### 运行项目
```bash
go run ./cmd/api
```

### 构建项目
```bash
go build ./cmd/api
```

## 📋 功能特性

### 用户管理
- ✅ 用户注册
- ✅ 用户查询
- ✅ 用户更新
- ✅ 用户删除
- ✅ 用户状态管理
- ✅ 分页查询

### 技术特性
- ✅ Clean Architecture 分层设计
- ✅ 依赖注入 (Wire)
- ✅ 结构化日志 (Zerolog)
- ✅ 统一错误处理
- ✅ 统一响应格式
- ✅ 请求追踪 (Trace ID)
- ✅ 中间件支持
- ✅ 配置管理
- ✅ 单元测试

## 🔧 配置说明

### 环境变量
项目支持通过环境变量覆盖配置，格式为 `SECTION_KEY`，例如：
- `HTTP_ADDRESS` 对应 `http.address`
- `DB_DRIVER` 对应 `db.driver`

### 配置文件
主要配置文件位于 `config/config.yaml`，包含：
- HTTP 服务配置
- 日志配置
- 数据库配置
- Redis 配置
- Asynq 配置
- Kafka 配置

## 🧪 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service

# 运行测试并显示覆盖率
go test -cover ./...
```

### 测试覆盖率
目标测试覆盖率 ≥ 60%

## 📚 API 文档

### 用户管理 API

#### 用户注册
```http
POST /api/v1/users
Content-Type: application/json

{
  "name": "用户名",
  "email": "user@example.com",
  "password": "password123"
}
```

#### 获取用户列表
```http
GET /api/v1/users?page=1&page_size=20&status=active
```

#### 获取用户详情
```http
GET /api/v1/users/{id}
```

#### 更新用户
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "name": "新用户名",
  "status": "inactive"
}
```

#### 删除用户
```http
DELETE /api/v1/users/{id}
```

#### 改变用户状态
```http
PATCH /api/v1/users/{id}/status
Content-Type: application/json

{
  "status": "banned"
}
```

## 🔍 当前状态

### ✅ 已完成
- Clean Architecture 架构设计
- 用户领域模型和接口定义
- 用户仓储层实现
- 用户服务层实现
- HTTP 处理器实现
- 统一错误处理
- 统一响应格式
- 配置管理
- 中间件实现
- 单元测试框架

### ⚠️ 已知问题
- 日志包 (pkg/logger) 存在 zerolog API 使用问题
- Ent 代码生成需要重新执行
- 部分依赖注入配置需要完善

### 🚧 进行中
- 项目重构和架构优化
- 代码质量改进

### 📋 计划中
- Redis 集成
- Asynq 任务队列
- Kafka 消息队列
- 监控和指标
- 集成测试
- Docker 支持
- CI/CD 配置

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发送邮件
- 参与讨论

---

**注意**: 这是一个重构中的项目，部分功能可能不稳定。建议在开发环境中使用。


