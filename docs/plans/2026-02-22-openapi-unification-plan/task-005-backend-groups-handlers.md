# Task 005: Implement Groups/GroupUsers handlers

**depends-on**: task-002

## Description

Implement all Groups and GroupUsers-related `StrictServerInterface` methods, migrating business logic from existing handler functions.

## Execution Context

**Task Number**: 005 of 013
**Phase**: Core Features
**Prerequisites**: Generated `StrictServerInterface` exists (Task 002); `Server` struct exists (Task 004 creates it, but methods are independent)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "列表用户 API" (groups are referenced in user listing), "分页参数"

## Files to Create/Modify

- Create: `panel/internal/api/server_groups.go`
- Existing reference: `panel/internal/api/groups.go` (source logic)
- Existing reference: `panel/internal/api/group_users.go` (source logic)

## Steps

### Step 1: Implement Groups methods

Migrate from `groups.go`:
- `ListGroups` — pagination, store fetch, DTO conversion
- `CreateGroup` — name validation, store create, DTO response
- `GetGroup` — ID parsing, store fetch, DTO response
- `UpdateGroup` — partial update, conflict handling, DTO response
- `DeleteGroup` — store delete with conflict check (group in use)

### Step 2: Implement GroupUsers methods

Migrate from `group_users.go`:
- `ListGroupUsers` — list users belonging to a group
- `ReplaceGroupUsers` — replace group's user membership, trigger node syncs

### Step 3: Verify compilation and existing tests

## Verification Commands

```bash
cd panel && go build ./...
cd panel && go test ./internal/api/ -run TestGroups -v
```

## Success Criteria

- All Groups and GroupUsers methods implemented
- `toGroupDTO` helper preserved
- Conflict detection (group name uniqueness, group in use) preserved
- Node sync triggers on group user changes preserved
