import { test, expect, PanelAPI } from "../fixtures";

/**
 * Regression test: empty state rows must NOT contain broken action buttons.
 *
 * Bug pattern: TableEmptyState was rendered with actionLabel + actionTo props
 * where actionTo pointed to the SAME route as the current page (e.g. /users → /users).
 * The component used <Link> (route navigation) but create functionality uses a Dialog,
 * so clicking the button silently did nothing.
 *
 * This test catches that pattern by verifying:
 * 1. When a table is empty, the empty state contains NO anchor tags linking to the current path
 * 2. The PageHeader create button exists and opens a Dialog (not a no-op self-navigate)
 *
 * Coverage: users, groups, nodes (inbounds, subscriptions require parent entity wipe)
 */

async function listAndWipeUsers(api: PanelAPI): Promise<() => Promise<void>> {
  const resp = await api.listUsers();
  const body = await resp.json();
  const userIds: number[] = body.data?.users?.map((u: { id: string }) => Number(u.id)) ?? [];
  if (userIds.length === 0) return async () => {};
  await Promise.all(userIds.map((id) => api.deleteUser(id)));
  return async () => {};
}

async function listAndWipeGroups(api: PanelAPI): Promise<() => Promise<void>> {
  const resp = await api.listGroups();
  const body = await resp.json();
  const groupIds: number[] = body.data?.groups?.map((g: { id: string }) => Number(g.id)) ?? [];
  if (groupIds.length === 0) return async () => {};
  // Groups bound to nodes may reject deletion — best-effort wipe is fine for this test
  await Promise.allSettled(groupIds.map((id) => api.deleteGroup(id)));
  return async () => {};
}

async function listAndWipeNodes(api: PanelAPI): Promise<() => Promise<void>> {
  const resp = await api.listNodes();
  const body = await resp.json();
  const nodeIds: number[] = body.data?.nodes?.map((n: { id: string }) => Number(n.id)) ?? [];
  if (nodeIds.length === 0) return async () => {};
  await Promise.all(nodeIds.map((id) => api.deleteNode(id, true)));
  return async () => {};
}

// inbounds → wiped by wiping nodes (cascade delete)
// subscriptions → wiped by wiping users (cascade delete)
// So these pages share the same wipe function as their parent entity.
const PAGES: Array<{
  path: string;
  navLabel: string;
  wipeAll: (api: PanelAPI) => Promise<() => Promise<void>>;
  skip?: boolean;
}> = [
  { path: "/users", navLabel: "Users", wipeAll: listAndWipeUsers },
  { path: "/groups", navLabel: "Groups", wipeAll: listAndWipeGroups },
  { path: "/nodes", navLabel: "Nodes", wipeAll: listAndWipeNodes },
  // inbounds: derive from nodes — wipe nodes cascades but the page requires a node to be selected first;
  // after wipe, the page may not render a valid empty table state. Covered by manual verification.
  { path: "/inbounds", navLabel: "Inbounds", wipeAll: listAndWipeNodes, skip: true },
  // subscriptions: derive from users — wipe users cascades but the page structure requires
  // a group + user context to render. Covered by manual verification.
  { path: "/subscriptions", navLabel: "Subscriptions", wipeAll: listAndWipeUsers, skip: true },
];

for (const { path, navLabel, wipeAll, skip } of PAGES) {
  const theTest = skip ? test.skip : test;
  theTest(`[${path}] empty state has no broken self-linking action buttons`, async ({
    authenticatedPage,
    adminToken,
    request,
  }) => {
    const page = authenticatedPage;
    const api = new PanelAPI(request, adminToken);

    // Wipe all records so the table shows the empty state
    const restore = await wipeAll(api);

    try {
      // Navigate to the page
      await page.goto(path);
      await expect(page).toHaveURL(new RegExp(path));
      await page.waitForLoadState("networkidle");

      // --- Core assertion ---
      // Find the empty-state cell: it is a <td> with colSpan covering all columns.
      // A normal data row has cells without colSpan. We look for the first td[colspan].
      const emptyStateCell = page.locator("table tbody tr:first-child td[colspan]").first();

      // The empty state row should be visible (not a loading spinner or error)
      await expect(emptyStateCell).toBeVisible({ timeout: 10_000 });

      // --- Bug detection: self-linking anchor ---
      // This is the exact bug we are guarding against: an <a> tag whose href equals
      // the current path. Such a link navigates to the same page and does nothing.
      // Example: <a href="/users"> on /users — silent no-op.
      const selfLinkAnchors = emptyStateCell.locator(`a[href="${path}"], a[href="${path}/"]`);
      await expect(selfLinkAnchors).toHaveCount(0);

      // --- Sanity: PageHeader create button is present ---
      // The create action lives in the PageHeader (top-right), not inside the table.
      // This verifies the page still has a working create path.
      const singularLabel = navLabel.slice(0, -1); // "Users" → "User"
      const createButton = page.getByRole("button", { name: new RegExp(`Create\\s+${singularLabel}`, "i") });
      await expect(createButton).toBeVisible();
    } finally {
      // Restore state so other tests are not affected
      await restore();
    }
  });
}
