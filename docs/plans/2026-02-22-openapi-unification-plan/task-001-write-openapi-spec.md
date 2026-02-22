# Task 001: Write complete OpenAPI 3.1 spec

**depends-on**: (none — foundational task)

## Description

Create the complete `panel/openapi.yaml` file that defines all ~37 API operations across ~25 paths. This is the Single Source of Truth for the entire migration. Every schema, parameter, and response must faithfully reflect the current API behavior observed in the Go handler code.

## Execution Context

**Task Number**: 001 of 013
**Phase**: Foundation
**Prerequisites**: Understanding of current API handlers in `panel/internal/api/*.go` and frontend types in `panel/web/src/lib/api/types.ts`

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: All — this is the foundation for every other scenario

## Files to Create

- Create: `panel/openapi.yaml`

## Steps

### Step 1: Define info, servers, securitySchemes

- OpenAPI version: `3.1.0`
- Server base URL: `/api`
- Security scheme: `BearerAuth` (HTTP Bearer with JWT)

### Step 2: Define reusable components

- `components/parameters/`: `LimitParam`, `OffsetParam`, `IdParam`
- `components/responses/`: `BadRequest`, `Unauthorized`, `NotFound`, `InternalError`, `Conflict`
- `components/schemas/`: `ErrorResponse`, `StatusResponse`, `MessageResponse`

### Step 3: Define entity schemas

Define all entity schemas in `components/schemas/` based on existing Go DTOs and TS types:

- **User-related**: `User`, `UserStatus` (enum), `CreateUserRequest`, `UpdateUserRequest`, `UserGroups`
- **Group-related**: `Group`, `CreateGroupRequest`, `UpdateGroupRequest`, `GroupUsersRequest`
- **Node-related**: `Node`, `CreateNodeRequest`, `UpdateNodeRequest`, `NodeHealthResponse`, `NodeDeleteResponse`
- **Inbound-related**: `Inbound`, `CreateInboundRequest`, `UpdateInboundRequest`, `SyncResult`, `InboundWithSyncResponse`
- **SyncJob-related**: `SyncJob`, `SyncJobStatus` (enum), `SyncAttempt`, `SyncJobDetail`
- **Traffic-related**: `TrafficNodeSummary`, `TrafficTotalSummary`, `TrafficTimeseriesPoint`, `NodeTrafficSample`
- **System-related**: `SystemInfo`, `SystemSettings`, `AdminProfile`, `UpdateAdminProfileRequest`
- **Auth-related**: `LoginRequest`, `LoginResponse`, `BootstrapStatus`, `BootstrapRequest`, `BootstrapResponse`
- **SingBox-related**: `SingBoxFormatRequest`, `SingBoxFormatResponse`, `SingBoxCheckRequest`, `SingBoxCheckResponse`, `SingBoxGenerateRequest`, `SingBoxGenerateResponse`

Key considerations:
- `settings`, `tls_settings`, `transport_settings` on Inbound use free-form JSON (`type: object` with `additionalProperties: true` or empty schema `{}`)
- `expire_at`, `last_seen_at`, `group_id` are nullable
- All `id` fields use `format: int64`
- Property names use `snake_case` to match existing JSON wire format

### Step 4: Define all paths and operations

For each path, define operations with:
- Unique `operationId` in camelCase (e.g., `listUsers`, `createUser`, `getUser`, `updateUser`, `deleteUser`)
- `security: []` for public endpoints; `security: [{ BearerAuth: [] }]` for protected
- Request body schemas for POST/PUT
- Query parameters for list/filter endpoints
- Path parameters for resource-specific endpoints
- All response codes with appropriate schemas

**Response envelope patterns** (must match current behavior):
- Most GET/POST/PUT: `{ "data": T }` — define inline response with `data` property
- Inbound POST/PUT: `{ "data": Inbound, "sync": SyncResult }` — compound response
- Most DELETE: `{ "status": "ok" }` — use `StatusResponse`
- User hard DELETE: `{ "message": "user deleted" }` — use `MessageResponse`
- Node force DELETE: `{ "status": "ok", "force": true, "deleted_inbounds": N }` — custom schema
- All errors: `{ "error": "..." }` — use `ErrorResponse`

**Public paths**: `/health`, `/admin/bootstrap` (GET+POST), `/admin/login`, `/sub/{user_uuid}`

**Authenticated paths**: `/admin/profile`, `/users`, `/users/{id}`, `/users/{id}/groups`, `/groups`, `/groups/{id}`, `/groups/{id}/users`, `/nodes`, `/nodes/{id}`, `/nodes/{id}/health`, `/nodes/{id}/sync`, `/nodes/{id}/traffic`, `/traffic/nodes/summary`, `/traffic/total/summary`, `/traffic/timeseries`, `/system/info`, `/system/settings`, `/sync-jobs`, `/sync-jobs/{id}`, `/sync-jobs/{id}/retry`, `/inbounds`, `/inbounds/{id}`, `/sing-box/format`, `/sing-box/check`, `/sing-box/generate`

### Step 5: Validate the spec

- Use an OpenAPI validator (e.g., `swagger-cli validate` or online tool) to ensure the spec is valid
- Cross-check every operationId against the router in `panel/internal/api/router.go`
- Cross-check every schema field against the Go DTO structs and TypeScript types

## Verification Commands

```bash
# Validate spec syntax (install if needed: go install github.com/pb33f/libopenapi-validator/cmd/openapi-validator@latest)
# Or use: npx @redocly/cli lint panel/openapi.yaml
npx @redocly/cli lint panel/openapi.yaml
```

## Success Criteria

- `panel/openapi.yaml` exists and passes OpenAPI 3.1 validation
- All ~37 operations from `router.go` are represented
- All entity schemas match existing Go DTOs and TS types
- Response envelope patterns match current API behavior
