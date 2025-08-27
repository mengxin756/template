## Classic Go Project (Gin + Wire + Ent)

### 快速开始

1) 生成 Ent 代码
```bash
go run entgo.io/ent/cmd/ent generate ./internal/data/ent/schema
```

2) 运行 API 服务
```bash
go run ./cmd/api
```

3) 健康检查与示例
- 健康检查: GET http://localhost:8080/healthz
- Ping: GET http://localhost:8080/api/v1/ping

### 配置
- 默认读取 `./config/config.yaml`，环境变量覆盖（`.` 替换为 `_`）。

### 日志
- Zap JSON 结构化，支持 `trace_id`；访问日志输出 method/path/status/latency/ip/bytes。
- Recovery 捕获 panic 并返回统一错误体。

### 目录结构
```
/cmd/api              # 入口
/internal/config      # 配置加载
/internal/logger      # 日志封装
/internal/server/http # Gin 引擎与路由
/internal/...         # 业务模块
/internal/data/ent    # Ent schema 与生成代码
/config               # 配置文件
```

### 后续
- 集成 Wire：在 `cmd/api` 组装 ProviderSet，生成 `wire_gen.go`。
- 按领域拆分 handler/service/repo，从 `user` 模块起步。


