# Task 002: Setup Go oapi-codegen toolchain

**depends-on**: task-001

## Description

Install oapi-codegen as a Go tool dependency, create configuration files for types and server generation, add `go generate` directives, and verify that generated Go code compiles successfully.

## Execution Context

**Task Number**: 002 of 013
**Phase**: Setup
**Prerequisites**: `panel/openapi.yaml` exists and is valid (Task 001)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenario**: "从 OpenAPI spec 生成 Go 后端代码"

## Files to Create/Modify

- Modify: `panel/go.mod` (add oapi-codegen dependency)
- Create: `panel/internal/api/oapi_types.cfg.yaml`
- Create: `panel/internal/api/oapi_server.cfg.yaml`
- Create: `panel/internal/api/generate.go`
- Generated: `panel/internal/api/oapi_types.gen.go`
- Generated: `panel/internal/api/oapi_server.gen.go`

## Steps

### Step 1: Add oapi-codegen tool dependency

- Add `github.com/oapi-codegen/oapi-codegen/v2` to `panel/go.mod` as a tool dependency
- Also add `github.com/oapi-codegen/runtime` for runtime support types
- Run `go mod tidy` in `panel/`

### Step 2: Create types generation config

- Create `panel/internal/api/oapi_types.cfg.yaml` with:
  - `package: api`
  - `generate: { models: true }`
  - `output: oapi_types.gen.go`
  - `output-options: { nullable-type: true }`

### Step 3: Create server generation config

- Create `panel/internal/api/oapi_server.cfg.yaml` with:
  - `package: api`
  - `generate: { gin-server: true, strict-server: true }`
  - `output: oapi_server.gen.go`
  - `output-options: { nullable-type: true }`

### Step 4: Create go generate file

- Create `panel/internal/api/generate.go` with `//go:generate` directives for both configs
- The directives should reference `../../openapi.yaml` (relative to the config file location)

### Step 5: Run code generation

- Execute `cd panel && go generate ./internal/api/...`
- Verify both `.gen.go` files are created
- Verify the generated code compiles: `cd panel && go build ./...`

### Step 6: Inspect generated interface

- Verify the generated `StrictServerInterface` contains methods for all operations defined in the spec
- Verify generated model types (User, Group, Node, Inbound, etc.) match the spec schemas

## Verification Commands

```bash
# Run code generation
cd panel && go generate ./internal/api/...

# Verify compilation (will fail because StrictServerInterface is not yet implemented — that's expected)
# Instead, just verify the generated files exist and have correct package declaration
head -5 panel/internal/api/oapi_types.gen.go
head -5 panel/internal/api/oapi_server.gen.go

# Count generated operations (grep for method signatures)
grep -c 'func.*StrictServerInterface' panel/internal/api/oapi_server.gen.go
```

## Success Criteria

- `oapi_types.gen.go` and `oapi_server.gen.go` are generated without errors
- Generated types package is `api`
- `StrictServerInterface` is defined with methods for all spec operations
- Model types are generated for all component schemas
- `go vet ./internal/api/...` passes on the generated files (ignoring unimplemented interface)
