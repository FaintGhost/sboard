# Task 001: Proto 定义与代码生成

**depends-on:** (none)

## Summary

在 Panel proto 中新增 `NodeRegistrationService` 服务和 Heartbeat/Approve/Reject 消息类型，然后运行 codegen 生成 Go 和 TypeScript stubs。

## BDD Scenario

```gherkin
Scenario: Proto definitions generate valid code
  Given panel.proto contains NodeRegistrationService with Heartbeat RPC
  And NodeService contains ApproveNode and RejectNode RPCs
  When `moon run panel:generate` is executed
  Then Go connect stubs are generated in panel/internal/rpc/gen/
  And TypeScript connect-query types are generated in web/src/lib/rpc/gen/
  And `moon run panel:check-generate` passes (no diff)
```

## Files

- **Modify:** `panel/proto/sboard/panel/v1/panel.proto`
  - 新增 `NodeRegistrationService` 服务，含 `Heartbeat` RPC
  - 在 `NodeService` 中新增 `ApproveNode` 和 `RejectNode` RPCs
  - 新增消息类型: `NodeHeartbeatRequest`, `NodeHeartbeatResponse`, `ApproveNodeRequest`, `ApproveNodeResponse`, `RejectNodeRequest`, `RejectNodeResponse`
  - `NodeHeartbeatResponse` 包含 `NodeHeartbeatStatus` 枚举: `UNSPECIFIED`, `RECOGNIZED`, `PENDING`, `REJECTED`

- **Auto-generated:** `panel/internal/rpc/gen/`, `web/src/lib/rpc/gen/` (由 codegen 自动更新)

## Steps

1. 编辑 `panel.proto`，在文件末尾添加新服务和消息定义
2. `NodeHeartbeatRequest` 字段: `uuid` (string), `secret_key` (string), `version` (string), `api_addr` (string)
3. `ApproveNodeRequest` 字段: `id` (int64), `name` (string), `group_id` (optional int64), `public_address` (optional string)
4. `RejectNodeRequest` 字段: `id` (int64)
5. 运行 `moon run panel:generate` 生成代码

## Verify

```bash
cd panel && moon run panel:check-generate
```
