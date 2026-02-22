# Task 011: Configure frontend client + migrate all pages

**depends-on**: task-003

## Description

Configure the generated @hey-api/client-fetch with custom auth interceptors and error handling, then migrate all frontend pages to use the generated SDK functions instead of hand-written API calls.

## Execution Context

**Task Number**: 011 of 013
**Phase**: Core Features
**Prerequisites**: Generated TypeScript SDK exists (Task 003)

## BDD Scenario Reference

**Spec**: `../2026-02-22-openapi-unification-design/bdd-specs.md`
**Scenarios**: "前端调用列表用户", "前端 401 自动登出", "前端 Zod 校验"

## Files to Modify

- Modify: `panel/web/src/lib/api/client.ts` — rewrite to configure generated client
- Modify: All page components that import from `@/lib/api/*`
- Modify: All hooks/stores that call API functions

## Steps

### Step 1: Rewrite client.ts

Replace the current hand-written `apiRequest` function with configuration for the generated client:
- Import client from `./gen/client.gen`
- Configure `baseUrl` using `VITE_API_BASE_URL` or `window.location.origin`
- Add request interceptor: inject `Authorization: Bearer <token>` from `useAuthStore`
- Add request interceptor: set `Accept: application/json`
- Add response interceptor: on 401, call `useAuthStore.getState().clearToken()`
- Re-export the configured client

### Step 2: Identify all import sites

Search for all imports from the old API layer:
- `@/lib/api/types` — replace with `@/lib/api/gen/types.gen`
- `@/lib/api/users` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/nodes` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/groups` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/inbounds` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/sync-jobs` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/traffic` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/auth` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/system` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/singbox-tools` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/user-groups` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/group-users` — replace with `@/lib/api/gen/sdk.gen`
- `@/lib/api/client` (for `ApiError`) — replace with import from new client.ts or gen

### Step 3: Migrate API calls in page components

For each page, update the API call pattern:

**Old pattern:**
```typescript
const users = await listUsers({ limit: 50 });
```

**New pattern:**
```typescript
const { data } = await listUsers({ query: { limit: 50 } });
const users = data?.data ?? [];
```

Key differences:
- SDK functions take `{ query: {...}, body: {...}, path: {...} }` structured params
- SDK returns `{ data, error, request, response }` — need to unwrap `data`
- Response envelope `{ "data": T }` means the actual data is at `data.data`
- Type names may differ slightly (check generated types vs old hand-written types)

Pages to migrate (in order of complexity):
1. `login-page.tsx` — uses auth API
2. `dashboard-page.tsx` — uses traffic API
3. `users-page.tsx` — uses users, user-groups API
4. `groups-page.tsx` — uses groups, group-users API
5. `nodes-page.tsx` — uses nodes API
6. `inbounds-page.tsx` — uses inbounds, singbox-tools API
7. `sync-jobs-page.tsx` — uses sync-jobs API
8. `subscriptions-page.tsx` — uses users API
9. `settings-page.tsx` — uses system API

Also migrate:
- `store/auth.ts` — login API call
- `components/bootstrap-form.tsx` — bootstrap API call
- `pages/users/use-user-mutations.ts` — user mutation hooks
- Any other files importing old API modules

### Step 4: Update React Query integrations

Existing React Query hooks need to wrap the new SDK functions:
- `queryFn` must unwrap the response: `const { data } = await sdk.fn(); return data?.data;`
- `queryKey` should align with operationId for consistency
- Mutation functions follow the same unwrap pattern

### Step 5: Verify TypeScript compilation

- Run `bunx tsc -b` and fix any type errors
- Ensure all generated types are compatible with component props

## Verification Commands

```bash
# TypeScript compilation check
cd panel/web && bunx tsc -b

# Lint check
cd panel/web && bun run lint

# Format check
cd panel/web && bun run format:check
```

## Success Criteria

- All page components use generated SDK functions
- Custom client.ts correctly configures auth and error interceptors
- TypeScript compiles without errors
- Lint passes
- No imports from old hand-written API files remain
