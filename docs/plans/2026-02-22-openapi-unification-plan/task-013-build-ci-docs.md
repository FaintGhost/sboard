# Task 013: Build integration, CI, documentation

**depends-on**: task-010, task-012

## Description

Add Makefile targets for code generation and freshness checking, update project documentation to reflect the new OpenAPI-driven workflow.

## Execution Context

**Task Number**: 013 of 013
**Phase**: Documentation
**Prerequisites**: Both backend and frontend migrations are complete and all tests pass

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "生成代码新鲜度检查", "Spec 变更触发双端更新"

## Files to Create/Modify

- Create: `Makefile` (or modify if exists) — add generate/check targets
- Modify: `CLAUDE.md` — update project documentation
- Modify: `panel/web/.gitignore` — ensure gen/ files are NOT ignored

## Steps

### Step 1: Create Makefile targets

Add the following targets to the project's Makefile (create if it doesn't exist):
- `generate`: runs both Go and TS code generation
- `generate-go`: runs `cd panel && go generate ./internal/api/...`
- `generate-ts`: runs `cd panel/web && bun run generate`
- `check-generate`: runs `generate` then checks for uncommitted changes in `*.gen.go` and `*.gen.ts` files

The `check-generate` target should:
1. Run `make generate`
2. Run `git diff --exit-code -- '*.gen.go' '*.gen.ts'`
3. If diff exists, print error message and exit 1

### Step 2: Verify freshness check works

- Run `make check-generate` — should succeed (exit 0) since code is fresh
- Manually modify `panel/openapi.yaml` (add a comment), run `make check-generate` — should fail
- Revert the change

### Step 3: Update CLAUDE.md

Add a section to `CLAUDE.md` documenting the new OpenAPI workflow:
- Location of spec file: `panel/openapi.yaml`
- How to regenerate: `make generate`
- CI freshness check: `make check-generate`
- Workflow for API changes: modify spec → regenerate → update handlers/pages → test
- Tools: `oapi-codegen` (Go), `@hey-api/openapi-ts` (TS)

Update existing sections that reference the old API structure:
- Update "关键 API（后端）" section to mention OpenAPI spec
- Update "前端交付门禁" to include `make check-generate`

### Step 4: Ensure generated files are tracked in git

- Verify `panel/web/.gitignore` does NOT include `src/lib/api/gen/`
- Verify `panel/.gitignore` does NOT include `*.gen.go`
- Generated files should be committed so CI can verify freshness

### Step 5: Final end-to-end verification

- Run `make generate` from project root
- Run all backend tests: `cd panel && go test ./... -count=1`
- Run all frontend checks: `cd panel/web && bun run test && bunx tsc -b && bun run lint && bun run format:check`
- Run `make check-generate` — should pass

## Verification Commands

```bash
# Generate all code
make generate

# Check freshness
make check-generate

# Full backend tests
cd panel && go test ./... -count=1

# Full frontend checks
cd panel/web && bun run test && bunx tsc -b && bun run lint && bun run format:check
```

## Success Criteria

- `make generate` works correctly from project root
- `make check-generate` detects stale generated code
- `CLAUDE.md` documents the OpenAPI workflow
- All tests pass
- All generated files are tracked in git
- Project fully builds and passes all quality gates
