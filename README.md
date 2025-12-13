# foundation

`foundation` 是 `ez-api`（Control Plane）与 `balancer`（Data Plane）共享的基础库仓库，用于沉淀“无状态、与业务无关”的通用能力，避免两边重复实现导致漂移。

## 包列表

- `github.com/ez-api/foundation/logging`：`log/slog` → `zerolog` handler bridge + 初始化入口。
- `github.com/ez-api/foundation/jsoncodec`：Sonic JSON 编解码统一入口。
- `github.com/ez-api/foundation/provider`：provider type 枚举/归一化/家族判断与默认值（例如 Vertex 默认 `global`）。

## 边界（与 DP/CP 分离不冲突）

- 可以放：编码解码、日志适配、无状态工具、枚举与默认值、与业务无关的通用校验。
- 不应该放：路由/负载策略、禁用/熔断决策、Redis key 结构、控制面 DTO/DB model、任何 DP/CP 专属业务逻辑。

## 发布与依赖

- 本仓库建议以语义化版本发布（tag）：`vX.Y.Z`。
- `ez-api`/`balancer` 在 `go.mod` 中依赖 `github.com/ez-api/foundation vX.Y.Z`，通过版本固化行为，避免“隐式漂移”。

## 本地多仓联调

当你需要同时改 `foundation` 与 `ez-api`/`balancer` 时，推荐在本机创建临时 Go workspace：

```bash
mkdir -p /workspace && cd /workspace
git clone <balancer_repo> balancer
git clone <ez_api_repo> ez-api
git clone <foundation_repo> foundation

go work init
go work use ./balancer ./ez-api ./foundation
```

然后在各自仓库下正常执行 `go test ./...` 即可。

