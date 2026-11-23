# go-web-template

Gin Web 项目快速启动模板：**自研 Module 插拔内核 + 全家桶组件 + 分层示例 + 脚手架 CLI**。
新项目不需要从零搭链路——一条命令实例化，组件想留谁就留谁。

## 特性

- **Module 注册表内核**（~120 行，无第三方 DI）：每个组件实现 4 个方法，`main.go` 里加/删一行即插拔
- **全家桶真实可用**：gin / gorm(mysql) / go-redis v9 / JWT / Prometheus / CORS / zap+lumberjack / viper
- **配置开关**：每个组件带 `enabled` 开关，配合环境变量覆盖，无需改代码即可增减依赖
- **正确的生命周期**：信号处理先于服务启动注册；Init 顺序执行、Stop 逆序回滚；异步启动失败（端口被占等）会触发整体优雅退出
- **可观测性打通**：`X-Request-Id` 贯穿日志与响应体，单一 `/metrics` 管理端口（gin HTTP 指标 + gorm 连接池 + Go 运行时）
- **示例即文档**：分层 user CRUD（handler → service → repository → model）+ JWT 登录演示，配 fake repo 单元测试
- **一键脚手架**：复制模板并改写 module 名/应用名
- Swagger、Makefile、Dockerfile、docker-compose（mysql + redis + app）

## 快速开始

```bash
git clone <this-repo> my-app && cd my-app
go run ./cmd/api        # 零外部依赖即可启动（database/redis 默认 enabled: false）
```

```bash
curl http://localhost:8080/ping          # {"code":200,...,"data":{"message":"pong"}}
curl http://localhost:9091/metrics       # 全部 Prometheus 指标
open http://localhost:8080/swagger/index.html
```

跑通全家桶（含示例 CRUD + 登录）：

```bash
docker compose up -d mysql redis
APP_DATABASE_ENABLED=true APP_REDIS_ENABLED=true go run ./cmd/api
```

```bash
curl -X POST http://localhost:8080/api/v1/users -H 'Content-Type: application/json' \
  -d '{"name":"alice","email":"alice@example.com"}'
curl -X POST http://localhost:8080/api/v1/login -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"***"}'
curl http://localhost:8080/api/v1/me -H "Authorization: Bearer <token>"
```

## 新项目脚手架

```bash
go run ./cmd/scaffold -module github.com/acme/blog -out ../blog -name blog
cd ../blog && go mod tidy && go build ./... && go run ./cmd/api
```

scaffold 会复制全部模板文件（跳过 `.git`、`bin`、`logs`、`.env` 和 scaffold 自身），并把旧 module 路径、应用名、镜像名、指标 namespace 替换为新值。生成后建议跑一次 `make swagger` 重新生成文档。

## 架构：Module 插拔内核

```
cmd/api/main.go                     ← 组件清单就在这里，加/删一行
  └─ kernel.New(config, logger, database, redis, metrics, user, httpserver)
       │
       ├─ 顺序 Init（后者可 MustRequire 前者 Provide 的服务）
       ├─ 阻塞等待：SIGINT/SIGTERM 或任一服务器的 Fail 上报
       └─ 逆序 Stop（独立超时上下文，永不使用已取消的 ctx）
```

```go
// internal/kernel/module.go
type Module interface {
	Name() string
	Enabled(cfg *config.Config) bool           // false 则完全跳过 Init/Stop
	Init(ctx context.Context, app *App) error  // 发布服务；长任务起 goroutine，失败调 app.Fail
	Stop(ctx context.Context) error            // 清理资源
}
```

内核维护一个按类型索引的服务容器：

```go
kernel.Provide(a, db)                    // database 模块发布 *gorm.DB
db := kernel.MustRequire[*gorm.DB](a)    // user 模块消费；未注册时 panic 并提示原因
kernel.Provide[auth.TokenParser](a, p)   // 接口类型发布需显式实例化类型参数
```

暴露 HTTP 路由的模块再实现一个接口（`internal/route`，内核本身不依赖 gin）：

```go
type Routable interface { RegisterRoutes(rg *gin.RouterGroup) }
```

`httpserver` 启动时只对**真正 Init 成功**的模块（`app.Started()`）扫描 Routable 并挂到 `/api/v1` 下——关掉 `database.enabled` 后示例业务的路由会整体消失，不会产生指向已死依赖的悬挂路由。

### 增加一个自己的组件

1. 新建 `internal/module/payment/module.go`，实现 `Module` 4 方法（参照 `internal/module/user`）
2. 需要配置就在 `internal/config` 加结构体 + 在 `load.go setDefaults` 补默认键
3. `cmd/api/main.go` 清单里加一行 `paymentmod.New()`（在依赖的服务之后、httpserver 之前）

完成。删除组件 = 删那一行，或配置 `payment.enabled: false`。

## 目录结构

```
cmd/
  api/main.go            入口：信号 + 模块清单
  scaffold/main.go       模板实例化 CLI
  etc/config.{dev,prod,example}.yaml
internal/
  kernel/                模块生命周期内核（不依赖 gin）
  config/                配置 schema + viper 加载 + zap 构造
  route/                 Routable 接口（gin 感知层）
  middleware/            requestid / zap logger / recovery / cors / jwt
  auth/                  JWT 签发与解析（HS*/RS*）
  module/                各组件：config logger database redis metrics httpserver user
  api/v1/ service/ repository/ model/   分层示例（handler→service→repo→gorm）
pkg/response/            统一 JSON 信封（含 trace_id）
docs/                    swag 生成（提交入库，make swagger 重新生成）
Makefile / Dockerfile / docker-compose.yml
```

## 配置与覆盖

优先级：内置默认值 < `config.<APP_ENV>.yaml`（默认 dev）< `.env` < 真实环境变量。

环境键规则：`APP_<段>_<键>`，点和横线换成下划线，如：

| 配置键 | 环境变量 |
|---|---|
| `database.dsn` | `APP_DATABASE_DSN` |
| `redis.addr` | `APP_REDIS_ADDR` |
| `jwt.signing-key` | `APP_JWT_SIGNING_KEY` |
| `database.enabled` | `APP_DATABASE_ENABLED` |

> 实现说明：viper 的 `AutomaticEnv` 不参与 `Unmarshal`，模板在 `load.go` 里对全部键做了 `BindEnv`，并让 `setDefaults` 铺满每个叶子键，因此**任何**键都可以通过环境变量覆盖（甚至完全没有配置文件）。

`server.mode`（dev/test/prod）一个值同时决定：配置文件名、zap 编码（console/JSON+日志轮转）、gin 模式。

## 指标与文档

- `:9091/metrics` 是模板里**唯一**的指标端点（ginprom 请求指标带 namespace 前缀、gorm dbstats、Go 运行时）。应用端口从不暴露 `/metrics`
- Swagger 路由注解写在 `internal/api/v1/*.go`，改完执行 `make swagger` 重新生成 `docs/`；`swagger.enabled: false` 可在生产关闭

## Makefile

```
make run        go run ./cmd/api
make build      输出 bin/
make test       go test ./... -race
make lint       go vet + gofmt
make swagger    重新生成 docs/
make docker-up  起 mysql/redis（compose）
```

Windows 没有 make 时，按 Makefile 里的同名命令直接执行即可。

## 生产部署要点

- 密钥全部走环境变量：`APP_DATABASE_DSN`、`APP_REDIS_PASSWORD`、`APP_JWT_SIGNING_KEY`
- `APP_ENV=prod` 使用 `cmd/etc/config.prod.yaml`（JSON 日志 + 文件轮转 + gin release 模式）
- 示例 user 模块和 seed 数据在上线前删除：`internal/module/user`、`internal/{api/v1,service,repository,model}`
- 需要 schema 迁移时引入 golang-migrate，别依赖 dev 的 AutoMigrate
