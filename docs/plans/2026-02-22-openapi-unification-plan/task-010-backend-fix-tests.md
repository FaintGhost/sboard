# Task 010: Fix and verify all backend tests

**depends-on**: task-009

## Description

Run the complete backend test suite and fix all failures caused by the migration. Existing tests may need adjustments to work with the new strict server handler signatures, but the HTTP-level behavior must remain unchanged.

## Execution Context

**Task Number**: 010 of 013
**Phase**: Testing
**Prerequisites**: Router is refactored and all handlers are wired up (Task 009)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "列表用户 API", "创建用户 API - 请求校验", "未认证访问受保护端点", "Inbound 创建的复合响应", "分页参数", "Go 类型与 spec 一致"

## Files to Modify

- Modify: Various `panel/internal/api/*_test.go` files as needed
- Delete: Old handler files that are fully replaced (e.g., `users.go`, `nodes.go`, `groups.go`, `inbounds.go`, etc.)

## Steps

### Step 1: Run full test suite and catalog failures

- Execute `cd panel && go test ./internal/api/ -v` and capture all failures
- Categorize failures:
  - **Import errors**: Old handler functions no longer exist
  - **Type mismatches**: DTOs replaced by generated types
  - **Compilation errors**: Function signatures changed
  - **Behavioral changes**: Response format differences

### Step 2: Fix compilation-level test failures

- Update test imports to use generated types instead of old DTOs
- Update test helpers that create handler functions (they now go through the strict server)
- Ensure test setup creates `Server` instances correctly

### Step 3: Fix behavioral test failures

- If any response format changed (it shouldn't if the spec is correct), fix the spec or handler
- Verify all HTTP status codes match expectations
- Verify all JSON response shapes match expectations

### Step 4: Delete old handler files

Once all tests pass with the new implementation, delete the superseded files:
- `panel/internal/api/users.go` (replaced by `server_users.go`)
- `panel/internal/api/nodes.go` (replaced by `server_nodes.go`)
- `panel/internal/api/groups.go` (replaced by `server_groups.go`)
- `panel/internal/api/inbounds.go` (replaced by `server_inbounds.go`)
- `panel/internal/api/group_users.go` (logic in `server_groups.go`)
- `panel/internal/api/user_groups.go` (logic in `server_users.go`)
- `panel/internal/api/sync_jobs.go` (replaced by `server_sync_jobs.go`)
- `panel/internal/api/traffic_aggregate.go` (replaced by `server_nodes.go` or `server_system.go`)
- `panel/internal/api/request_params.go` (pagination now handled by generated code)
- Other files that are fully superseded

Keep files that contain shared utilities: `node_sync_helpers.go`, `auth.go` (if middleware is still separate), `cors.go`, `request_logger.go`, `web.go`

### Step 5: Run full test suite again

- All tests must pass
- Run `go vet ./internal/api/...` to check for issues

## Verification Commands

```bash
# Full test suite
cd panel && go test ./internal/api/ -v -count=1

# Also run tests for other packages that might be affected
cd panel && go test ./... -count=1

# Vet check
cd panel && go vet ./...
```

## Success Criteria

- ALL existing tests pass (with necessary adjustments for new handler signatures)
- No behavioral changes to the HTTP API (same request/response format)
- Old handler files deleted
- `go vet` passes cleanly
- `go build ./...` succeeds
