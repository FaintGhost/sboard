import { defineConfig, devices } from "@playwright/test";

const isCI = !!process.env.CI;
const baseURL = process.env.BASE_URL || "http://localhost:8080";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers: 1,
  timeout: 30_000,
  expect: { timeout: 10_000 },

  reporter: isCI
    ? [
        ["dot"],
        ["html", { open: "never", outputFolder: "playwright-report" }],
      ]
    : [
        ["line"],
        ["html", { open: "on-failure", outputFolder: "playwright-report" }],
      ],

  use: {
    baseURL,
    headless: true,
    trace: isCI ? "on-first-retry" : "off",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
    locale: "en-US",
  },

  outputDir: "test-results",

  projects: [
    {
      name: "smoke",
      testMatch: /.*\.smoke\.spec\.ts/,
      retries: 0,
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "e2e",
      testMatch: /tests\/e2e\/.*\.spec\.ts/,
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
