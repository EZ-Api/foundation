# Responses Ingress Contract Generation

`foundation/contract/responses_ingress.go` 是 **生成文件**，用于给 DP/CP 运行时代码提供稳定的 Go 常量与校验函数。

## 为什么 `ez-contract` + `foundation/contract` 同时存在

- `ez-contract`：
  - 对外契约源（OpenAPI / schema / examples），偏文档与接口标准。
  - 是“协议真相源（source of truth）”。
- `foundation/contract`：
  - 运行时可直接 import 的 Go 契约包（常量、校验、golden）。
  - 是“工程消费层（runtime consumption layer）”。

两者职责不同，避免让业务仓库直接解析 YAML；通过生成流程保持一致，避免手写漂移。

## 生成输入（来自 ez-contract）

源文件：`../../ez-contract/schemas/responses/responses.yaml`

机读字段：

- `ResponsesRequest.x-ez-ingress-field-policy`
  - `pass_through`
  - `reject`
- `ResponsesResponse.x-ez-ingress-stream-events`
  - `output_text_delta`
  - `output_tool_call_delta`
  - `completed`
  - `error`

## 生成输出

- `responses_ingress.go`
  - SSE 事件常量（`ResponsesEvent*`）
  - 字段决策枚举与函数（`ResponsesRequestFieldDecision`）
  - 字段校验函数（`ValidateResponsesRequestFields`）
  - 可观测/文档辅助函数（`ResponsesPassThroughRequestFields` / `ResponsesRejectedRequestFields`）

## 如何更新

1. 修改 `ez-contract/schemas/responses/responses.yaml` 中的 metadata。
2. 在 `foundation` 仓执行：
   - `go generate ./contract`
3. 运行测试：
   - `go test ./...`
4. 在 `balancer` 仓执行回归：
   - `go test ./internal/protocol/openai ./internal/proxy`

## 生成器实现

- 入口：`contract/generate.go`
- 命令：`contract/cmd/genresponsesingress/main.go`
- 默认 schema 路径：`../../ez-contract/schemas/responses/responses.yaml`
- 可覆盖：环境变量 `EZ_CONTRACT_RESPONSES_SCHEMA` 或命令参数 `-schema`
