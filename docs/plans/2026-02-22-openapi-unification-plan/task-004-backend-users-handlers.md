# Task 004: Implement Server struct + Users/UserGroups handlers

**depends-on**: task-002

## Description

Create the `Server` struct that implements `StrictServerInterface`, and implement all Users and UserGroups-related methods. This is the first handler migration task that establishes the pattern for subsequent tasks.

## Execution Context

**Task Number**: 004 of 013
**Phase**: Core Features
**Prerequisites**: Generated `StrictServerInterface` exists (Task 002)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "列表用户 API", "创建用户 API - 请求校验", "分页参数"

## Files to Create/Modify

- Create: `panel/internal/api/server.go` — Server struct definition
- Create: `panel/internal/api/server_users.go` — Users handler implementations
- Existing reference: `panel/internal/api/users.go` (source of business logic to migrate)
- Existing reference: `panel/internal/api/user_groups.go` (source of business logic to migrate)

## Steps

### Step 1: Create Server struct

- Create `panel/internal/api/server.go` with `Server` struct containing `store *db.Store` and `cfg config.Config`
- Ensure the struct satisfies `StrictServerInterface` (add compile-time check: `var _ StrictServerInterface = (*Server)(nil)`)
- Add stub methods for ALL interface methods initially (return 501 Not Implemented) so the project compiles

### Step 2: Implement Users methods

Migrate business logic from existing `users.go` handler functions into `StrictServerInterface` methods on `Server`:

- `ListUsers` — migrate from `UsersList()`: pagination parsing, status filtering, batch group ID fetch, DTO construction
- `CreateUser` — migrate from `UsersCreate()`: username validation, store call, DTO response
- `GetUser` — migrate from `UsersGet()`: ID parsing, store fetch, DTO response
- `UpdateUser` — migrate from `UsersUpdate()`: partial update parsing, store call, sync trigger, DTO response
- `DeleteUser` — migrate from `UsersDelete()`: soft delete vs hard delete (based on `hard` query param), sync triggers

Key changes from old to new pattern:
- No more `*gin.Context` — use typed request objects
- No more `c.JSON()` — return typed response objects
- Error responses use generated response types (e.g., `ListUsers400JSONResponse`, `ListUsers500JSONResponse`)
- Preserve all existing business logic: `effectiveUserStatus`, `buildUserDTO`, `syncNodesForUser`, etc.

### Step 3: Implement UserGroups methods

Migrate from `user_groups.go`:
- `GetUserGroups` — migrate from `UserGroupsGet()`
- `UpdateUserGroups` — migrate from `UserGroupsPut()`

### Step 4: Verify compilation

- Ensure the project compiles with the new server methods
- Run relevant existing tests to verify behavior preservation

## Verification Commands

```bash
# Compile check
cd panel && go build ./...

# Run existing users tests
cd panel && go test ./internal/api/ -run TestUsers -v
```

## Success Criteria

- `Server` struct created with all stub methods (project compiles)
- Users and UserGroups methods fully implemented with business logic migrated
- `toUserDTO`, `buildUserDTO`, helper functions preserved and working
- Error handling follows the generated response type pattern
- No changes to external API behavior
