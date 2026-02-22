# Task 006: Implement Nodes handlers

**depends-on**: task-002

## Description

Implement all Nodes-related `StrictServerInterface` methods, including health check, sync trigger, and traffic listing. Node deletion with force-drain is a complex operation that must be carefully migrated.

## Execution Context

**Task Number**: 006 of 013
**Phase**: Core Features
**Prerequisites**: Generated `StrictServerInterface` exists (Task 002)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "分页参数"

## Files to Create/Modify

- Create: `panel/internal/api/server_nodes.go`
- Existing reference: `panel/internal/api/nodes.go` (source logic)
- Existing reference: `panel/internal/api/nodes_sync.go` (sync logic)
- Existing reference: `panel/internal/api/traffic.go` (traffic listing)

## Steps

### Step 1: Implement basic Nodes CRUD

Migrate from `nodes.go`:
- `ListNodes` — pagination, store fetch, DTO conversion
- `CreateNode` — validation (name, api_address, port, secret_key, public_address), store create
- `GetNode` — store fetch, DTO response
- `UpdateNode` — partial update with `optionalInt64` handling for nullable `group_id`
- `DeleteNode` — standard delete AND force delete with inbound drain

Key complexity: `DeleteNode` with `force=true` needs to:
1. Fetch node's inbounds
2. Send empty sync payload to drain the node
3. Delete all inbounds from DB
4. Delete the node
5. Return compound response `{"status": "ok", "force": true, "deleted_inbounds": N}`

### Step 2: Implement Node operations

- `GetNodeHealth` — proxy health check to node
- `TriggerNodeSync` — trigger sync operation
- `ListNodeTraffic` — list traffic samples for a node

### Step 3: Implement Traffic aggregate endpoints

Migrate from `traffic_aggregate.go`:
- `GetTrafficNodesSummary` — aggregate traffic per node
- `GetTrafficTotalSummary` — total traffic summary
- `GetTrafficTimeseries` — time-bucketed traffic data

### Step 4: Verify compilation and existing tests

## Verification Commands

```bash
cd panel && go build ./...
cd panel && go test ./internal/api/ -run TestNode -v
```

## Success Criteria

- All Nodes methods implemented including force delete
- `toNodeDTO` helper preserved
- Node sync helpers (`nodeLock`, `nodeClientFactory`) preserved and working
- Traffic aggregation logic preserved
- `optionalInt64` handling for nullable `group_id` works correctly
