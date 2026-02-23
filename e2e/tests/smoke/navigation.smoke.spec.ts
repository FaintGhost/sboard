import { test, expect } from "../fixtures";

const routes = [
  { path: "/", name: "Dashboard" },
  { path: "/users", name: "Users" },
  { path: "/groups", name: "Groups" },
  { path: "/nodes", name: "Nodes" },
  { path: "/inbounds", name: "Inbounds" },
  { path: "/sync-jobs", name: "Sync Jobs" },
  { path: "/subscriptions", name: "Subscriptions" },
  { path: "/settings", name: "Settings" },
];

test.describe("核心页面导航", () => {
  for (const route of routes) {
    test(`页面加载: ${route.name} (${route.path})`, async ({ authenticatedPage }) => {
      const page = authenticatedPage;
      const errors: string[] = [];

      // Listen for uncaught JS errors
      page.on("pageerror", (err) => {
        errors.push(err.message);
      });

      await page.goto(process.env.BASE_URL + route.path);

      // Page should not be blank — check that body has meaningful content
      await expect(page.locator("body")).not.toBeEmpty();

      // Sidebar navigation should be visible (proves layout loaded)
      await expect(
        page.getByRole("link", { name: route.name }).first()
      ).toBeVisible({ timeout: 10_000 });

      // No JS errors should have occurred
      expect(errors).toEqual([]);
    });
  }
});
