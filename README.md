# foundation

`foundation` 是 `ez-api`（Control Plane）与 `balancer`（Data Plane）共享的基础库仓库：专门承载“无状态、与业务无关”的通用能力，让 DP/CP 在**协议约定与基础设施层**保持一致，避免两边重复实现导致漂移。

## 适用场景

- 你需要一个统一的 JSON 编解码入口（性能/行为一致）。
- 你希望业务日志统一走 `log/slog`，但输出后端仍使用 `zerolog`。
- 你需要把 provider type 的枚举、归一化与默认值集中管理（例如 Vertex 默认 `global`）。

## 包一览

- `github.com/ez-api/foundation/jsoncodec`：基于 Sonic 的 JSON 编解码统一入口。
- `github.com/ez-api/foundation/logging`：`log/slog` → `zerolog` handler bridge + 初始化入口。
- `github.com/ez-api/foundation/provider`：provider type 枚举/归一化/家族判断与默认值。

## 快速开始

### 依赖

在你的项目里添加依赖（建议使用已发布的 tag 版本）：

```bash
go get github.com/ez-api/foundation@v0.1.0
```

### JSON

```go
import "github.com/ez-api/foundation/jsoncodec"

payload, _ := jsoncodec.Marshal(map[string]any{"ok": true})
```

### 日志（slog + zerolog 后端）

```go
import "github.com/ez-api/foundation/logging"

logger, _ := logging.New(logging.Options{Service: "my-service"})
logger.Info("hello", "k", "v")
```

## 设计边界（与 DP/CP 分离不冲突）

DP/CP 分离强调的是运行时职责与依赖边界；`foundation` 只提供基础能力，不承载业务决策。

- 可以放：编码解码、日志适配、无状态工具、枚举与默认值、与业务无关的通用校验。
- 不应该放：路由/负载策略、禁用/熔断决策、Redis key 结构、控制面 DTO/DB model、任何 DP/CP 专属业务逻辑。

## 发布策略

- 使用语义化版本发布（git tag）：例如 `v0.1.0`、`v0.2.0`。
- `ez-api`/`balancer` 在 `go.mod` 中锁定到明确版本，保证行为可复现。

## 本地多仓联调（可选）

当你需要同时改 `foundation` 与 `ez-api`/`balancer` 时，推荐在本机创建临时 Go workspace：

```bash
go work init
go work use ./balancer ./ez-api ./foundation
```
