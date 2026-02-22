# Task 007: Implement Inbounds handlers

**depends-on**: task-002

## Description

Implement all Inbounds-related `StrictServerInterface` methods. Inbound operations are notable for their compound response format (`{"data": inbound, "sync": syncResult}`) and protocol-specific validation.

## Execution Context

**Task Number**: 007 of 013
**Phase**: Core Features
**Prerequisites**: Generated `StrictServerInterface` exists (Task 002)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenario**: "Inbound 创建的复合响应"

## Files to Create/Modify

- Create: `panel/internal/api/server_inbounds.go`
- Existing reference: `panel/internal/api/inbounds.go` (source logic)

## Steps

### Step 1: Implement Inbounds CRUD

Migrate from `inbounds.go`:
- `ListInbounds` — pagination + optional `node_id` filter
- `GetInbound` — store fetch, DTO response
- `CreateInbound` — extensive validation (tag, protocol, port, JSON settings, protocol-specific validation via `inbval.ValidateSettings`), store create, auto-trigger node sync, return compound response
- `UpdateInbound` — partial update with merged validation (current + updated fields), auto-trigger node sync, return compound response
- `DeleteInbound` — store delete, auto-trigger node sync, return compound response with sync status

### Step 2: Handle compound response pattern

The create/update/delete responses include both the resource data AND a sync result:
- `{"data": inboundDTO, "sync": {"status": "success"|"error", "error": "..."}}`

This must be defined as a specific response schema in the OpenAPI spec and the strict server response type must include both fields.

### Step 3: Preserve protocol-specific validation

- `inbval.ValidateSettings(proto, settingsMap)` must still be called
- JSON validity checks for `settings`, `tls_settings`, `transport_settings` must be preserved
- `json.RawMessage` handling for free-form JSON fields

### Step 4: Verify compilation and existing tests

## Verification Commands

```bash
cd panel && go build ./...
cd panel && go test ./internal/api/ -run TestInbound -v
```

## Success Criteria

- All Inbound methods implemented
- Compound response format preserved (data + sync)
- Protocol-specific validation (`inbval.ValidateSettings`) preserved
- Auto node sync on inbound changes preserved
- `toInboundDTO` helper preserved
- `json.RawMessage` handling for free-form JSON fields works correctly
