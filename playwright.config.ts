import { defineConfig, devices } from "@playwright/test";

const desktop = process.env.BUZZ_DESKTOP_ROOT;
if (!desktop) {
  throw new Error("BUZZ_DESKTOP_ROOT must point to the Buzz desktop checkout");
}

export default defineConfig({
  testDir: "./tests/ui",
  timeout: 120_000,
  workers: 1,
  fullyParallel: false,
  reporter: [["list"], ["html", { open: "never", outputFolder: "playwright-report" }]],
  use: {
    baseURL: "http://127.0.0.1:4173",
    ...devices["Desktop Chrome"],
    viewport: { width: 1280, height: 800 },
    // Always keep video for demos
    video: {
      mode: "on",
      size: { width: 1280, height: 800 },
    },
    screenshot: "on",
    trace: "retain-on-failure",
  },
  outputDir: "docs/demos/playwright-output",
  webServer: {
    command: "python3 -m http.server 4173 -d dist",
    cwd: desktop,
    url: "http://127.0.0.1:4173",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
