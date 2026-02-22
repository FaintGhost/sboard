# Task 009: Refactor router + integrate auth middleware

**depends-on**: task-004, task-005, task-006, task-007, task-008

## Description

Replace the hand-written router in `router.go` with oapi-codegen's generated `RegisterHandlers` function. Integrate JWT authentication middleware using oapi-codegen's security scheme mechanism, replacing the current Gin group-based auth middleware.

## Execution Context

**Task Number**: 009 of 013
**Phase**: Integration
**Prerequisites**: All StrictServerInterface methods are implemented (Tasks 004-008)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenario**: "未认证访问受保护端点"

## Files to Modify

- Modify: `panel/internal/api/router.go` — replace hand-written routes with `RegisterHandlersWithOptions`
- Modify: `panel/internal/api/auth.go` — adapt JWT middleware to work with oapi-codegen's middleware mechanism
- Possibly modify: `panel/internal/api/server.go` — ensure Server creation accepts all needed dependencies

## Steps

### Step 1: Refactor router.go

Replace the current manual route registration with oapi-codegen's generated handler registration:
- Create `Server` instance with store and config
- Create strict handler via `NewStrictHandler(server, middlewares)`
- Register all routes via `RegisterHandlersWithOptions(r, strictHandler, GinServerOptions{...})`
- Set `BaseURL: "/api"` in options
- Preserve existing Gin middleware: `RequestLogger`, `Recovery`, `CORSMiddleware`
- Preserve `ServeWebUI` for static file serving

### Step 2: Integrate authentication middleware

oapi-codegen supports security scheme middleware. Two approaches:

**Option A (Recommended)**: Use per-operation middleware via `GinServerOptions.Middlewares`
- The middleware checks the operation's security requirements
- Public endpoints (health, login, bootstrap, subscription) have `security: []` — middleware skips them
- Protected endpoints have `security: [BearerAuth]` — middleware validates JWT

**Option B**: Use strict middleware (`StrictMiddlewareFunc`)
- Intercept at the strict handler level
- Inspect the operation to determine if auth is needed

Choose the approach that best integrates with the existing `AuthMiddleware` JWT validation logic.

### Step 3: Handle timezone initialization

- The current router calls `InitSystemTimezone` — ensure this is preserved
- Either call it in the new `NewRouter` or in the server constructor

### Step 4: Handle singBoxToolsFactory dependency

- The current router calls `singBoxToolsFactory()` — ensure the SingBox tools are available to the Server
- Add as a field on the Server struct or pass via dependency injection

### Step 5: Verify the full router works

- The application should start and all endpoints should be reachable
- Auth-protected endpoints should reject requests without valid JWT
- Public endpoints should work without auth

## Verification Commands

```bash
# Compile the full application
cd panel && go build ./cmd/panel/...

# Run all API tests
cd panel && go test ./internal/api/ -v
```

## Success Criteria

- `router.go` uses `RegisterHandlersWithOptions` instead of manual route registration
- All ~25 paths are correctly registered
- JWT authentication works: protected endpoints require valid Bearer token
- Public endpoints (health, login, bootstrap, subscription) work without auth
- CORS, request logging, and recovery middleware preserved
- WebUI static file serving preserved
- Application compiles and starts successfully
