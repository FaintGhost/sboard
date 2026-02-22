# Task 012: Fix frontend tests + delete old API files

**depends-on**: task-011

## Description

Run the frontend test suite, fix any failures from the API migration, then delete all superseded hand-written API files.

## Execution Context

**Task Number**: 012 of 013
**Phase**: Testing
**Prerequisites**: All pages migrated to generated SDK (Task 011)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "前端 Zod 校验", "TypeScript 类型与 spec 一致"

## Files to Delete

- `panel/web/src/lib/api/types.ts`
- `panel/web/src/lib/api/users.ts`
- `panel/web/src/lib/api/nodes.ts`
- `panel/web/src/lib/api/groups.ts`
- `panel/web/src/lib/api/inbounds.ts`
- `panel/web/src/lib/api/sync-jobs.ts`
- `panel/web/src/lib/api/traffic.ts`
- `panel/web/src/lib/api/auth.ts`
- `panel/web/src/lib/api/system.ts`
- `panel/web/src/lib/api/singbox-tools.ts`
- `panel/web/src/lib/api/user-groups.ts`
- `panel/web/src/lib/api/group-users.ts`
- `panel/web/src/lib/api/pagination.ts`
- `panel/web/src/lib/api/client.test.ts` (if testing old client)
- `panel/web/src/lib/api/users.test.ts`
- `panel/web/src/lib/api/nodes.test.ts`

## Steps

### Step 1: Run frontend test suite

- Execute `cd panel/web && bun run test`
- Catalog all failures

### Step 2: Fix test failures

- Tests that import from old API files need to import from generated SDK
- Tests that mock `apiRequest` need to mock the generated client instead
- Update type assertions to use generated types
- Fix any test helper utilities that reference old API functions

### Step 3: Delete old API files

- Remove all files listed above
- Verify no remaining imports reference deleted files: search for `from "@/lib/api/types"`, `from "@/lib/api/users"`, etc.

### Step 4: Run all quality checks

- Run full test suite again
- Run TypeScript compilation
- Run lint
- Run format check

## Verification Commands

```bash
# Full test suite
cd panel/web && bun run test

# TypeScript compilation
cd panel/web && bunx tsc -b

# Lint
cd panel/web && bun run lint

# Format
cd panel/web && bun run format:check
```

## Success Criteria

- All frontend tests pass
- All old hand-written API files deleted
- No dangling imports to deleted files
- TypeScript compilation passes
- Lint passes
- Format passes
