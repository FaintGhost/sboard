# Task 008: Implement System/Auth/Bootstrap/Traffic/SyncJobs/SingBox handlers

**depends-on**: task-002

## Description

Implement all remaining `StrictServerInterface` methods that haven't been covered by Tasks 004-007. This includes system info, settings, admin auth/profile, bootstrap, sync jobs, and sing-box tools.

## Execution Context

**Task Number**: 008 of 013
**Phase**: Core Features
**Prerequisites**: Generated `StrictServerInterface` exists (Task 002)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenario**: "未认证访问受保护端点"

## Files to Create/Modify

- Create: `panel/internal/api/server_system.go`
- Create: `panel/internal/api/server_sync_jobs.go`
- Existing reference: `panel/internal/api/auth.go`, `panel/internal/api/bootstrap.go`, `panel/internal/api/admin_profile.go`
- Existing reference: `panel/internal/api/system_settings.go`, `panel/internal/api/singbox_tools.go`
- Existing reference: `panel/internal/api/sync_jobs.go`, `panel/internal/api/subscription.go`

## Steps

### Step 1: Implement Auth/Bootstrap methods

- `AdminLogin` — credential validation, JWT generation
- `GetBootstrapStatus` — check if initial setup is needed
- `CreateBootstrap` — initial admin setup with setup token
- `GetAdminProfile` — return current admin profile
- `UpdateAdminProfile` — update admin credentials

### Step 2: Implement System methods

- `GetSystemInfo` — return panel version, commit, sing-box version
- `GetSystemSettings` — return subscription base URL, timezone
- `UpdateSystemSettings` — update settings

### Step 3: Implement SyncJobs methods

- `ListSyncJobs` — pagination + filters (node_id, status, trigger_source, date range)
- `GetSyncJob` — return job detail with attempts
- `RetrySyncJob` — retry a failed job

### Step 4: Implement SingBox tools methods

- `SingBoxFormat` — format sing-box config JSON
- `SingBoxCheck` — validate sing-box config
- `SingBoxGenerate` — generate sing-box artifacts (uuid, keys, etc.)

Note: SingBox tools use a `singBoxToolsFactory()` pattern. This factory must be available to the Server struct or passed as a dependency.

### Step 5: Implement Subscription endpoint

- `GetSubscription` — return subscription data based on user UUID, format parameter, and User-Agent header
- This endpoint returns different content types (JSON for sing-box, plain text for v2ray)
- Must handle the multi-format response pattern in the spec

### Step 6: Implement Health endpoint

- `GetHealth` — simple health check returning status

### Step 7: Verify compilation

## Verification Commands

```bash
cd panel && go build ./...
cd panel && go test ./internal/api/ -run "TestAuth|TestBootstrap|TestSystem|TestSync|TestSingBox|TestSubscription" -v
```

## Success Criteria

- All remaining StrictServerInterface methods implemented (no more stubs)
- JWT token generation/validation logic preserved
- SingBox tools factory pattern integrated
- Subscription multi-format response handled
- Sync jobs filtering and retry logic preserved
