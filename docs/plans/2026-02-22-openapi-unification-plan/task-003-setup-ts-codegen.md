# Task 003: Setup TypeScript @hey-api/openapi-ts toolchain

**depends-on**: task-001

## Description

Install @hey-api/openapi-ts and related plugins, create configuration file, add npm script, and verify that generated TypeScript code compiles successfully.

## Execution Context

**Task Number**: 003 of 013
**Phase**: Setup
**Prerequisites**: `panel/openapi.yaml` exists and is valid (Task 001)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenario**: "从 OpenAPI spec 生成 TypeScript 前端代码"

## Files to Create/Modify

- Modify: `panel/web/package.json` (add dependencies and script)
- Create: `panel/web/openapi-ts.config.ts`
- Generated: `panel/web/src/lib/api/gen/types.gen.ts`
- Generated: `panel/web/src/lib/api/gen/sdk.gen.ts`
- Generated: `panel/web/src/lib/api/gen/zod.gen.ts`
- Generated: `panel/web/src/lib/api/gen/client.gen.ts`

## Steps

### Step 1: Install dependencies

- Add to `devDependencies`: `@hey-api/openapi-ts`
- Add to `dependencies`: `@hey-api/client-fetch`
- Run `bun install` in `panel/web/`

### Step 2: Create openapi-ts config

- Create `panel/web/openapi-ts.config.ts` using `defineConfig` from `@hey-api/openapi-ts`
- Input: `../openapi.yaml` (relative path to spec)
- Output: `src/lib/api/gen`
- Plugins:
  - `@hey-api/typescript` — TypeScript type generation
  - `@hey-api/sdk` — SDK client functions (function style, not class)
  - `@hey-api/client-fetch` — Fetch-based HTTP client
  - `zod` — Zod schema generation with `definitions: true`, `requests: true`, `responses: true`

### Step 3: Add generate script

- Add `"generate": "openapi-ts"` to `scripts` in `panel/web/package.json`

### Step 4: Run code generation

- Execute `cd panel/web && bun run generate`
- Verify `src/lib/api/gen/` directory is created with `.gen.ts` files

### Step 5: Verify TypeScript compilation

- Run `cd panel/web && bunx tsc -b` (should compile, possibly with errors from existing code — generated files themselves should be clean)
- Inspect generated types to verify they match spec schemas

## Verification Commands

```bash
# Run code generation
cd panel/web && bun run generate

# Check generated files exist
ls panel/web/src/lib/api/gen/

# Verify TypeScript compiles (generated files only)
cd panel/web && bunx tsc --noEmit src/lib/api/gen/types.gen.ts
```

## Success Criteria

- `panel/web/src/lib/api/gen/` contains `types.gen.ts`, `sdk.gen.ts`, `zod.gen.ts`, `client.gen.ts`
- Generated TypeScript types match the OpenAPI spec schemas
- Generated Zod schemas correspond to each entity type
- Generated SDK contains functions for all operations with correct names (from operationId)
- Generated code passes TypeScript compilation
