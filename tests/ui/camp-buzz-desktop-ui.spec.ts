/**
 * Live UI proof: post via camp-buzz (real buzz CLI → local relay), then open
 * Buzz Desktop e2e build with installRelayBridge and assert the message appears.
 *
 * Requires:
 *   - buzz relay on http://localhost:3000
 *   - BUZZ_BIN / real buzz on PATH
 *   - desktop e2e dist at $BUZZ_DESKTOP/dist
 *
 * Run: just demo-ui  (from camp-buzz)
 */
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { expect, test } from "@playwright/test";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const campBuzzRoot = path.resolve(__dirname, "../..");

// Buzz desktop checkout (prefer hermit/worktree used in this environment)
const BUZZ_DESKTOP =
  process.env.BUZZ_DESKTOP ??
  "/Users/lancerogers/Dev/AI/obey_example_repos/buzz/desktop";

// Tyler's test identity from Buzz desktop e2e helpers (seeded by setup-desktop-test-data).
const TYLER_SK =
  "3dbaebadb5dfd777ff25149ee230d907a15a9e1294b40b830661e65bb42f6c03";

async function loadInstallRelayBridge() {
  // Dynamic import from the Buzz desktop tree so we reuse their real bridge.
  const bridgePath = path.join(BUZZ_DESKTOP, "tests/helpers/bridge.ts");
  // Playwright ts config may not resolve outside root — use compiled approach via path
  const mod = await import(bridgePath);
  return mod as {
    installRelayBridge: (
      page: import("@playwright/test").Page,
      user?: string,
    ) => Promise<void>;
  };
}

function run(
  bin: string,
  args: string[],
  env: Record<string, string>,
): string {
  return execFileSync(bin, args, {
    encoding: "utf8",
    env: { ...process.env, ...env },
  }).trim();
}

test.describe("camp-buzz message visible in Buzz Desktop UI", () => {
  test("posts via camp-buzz and appears in channel timeline", async ({
    page,
  }) => {
    const buzzBin =
      process.env.BUZZ_BIN ??
      "/Users/lancerogers/Dev/AI/obey_example_repos/buzz/target/release/buzz";
    const campBuzzBin = path.join(campBuzzRoot, "bin/camp-buzz");
    const relay = process.env.BUZZ_RELAY_URL ?? "http://localhost:3000";

    const marker = `camp-buzz UI proof ${Date.now()}`;

    // Create a disposable channel as tyler
    const created = JSON.parse(
      run(
        buzzBin,
        [
          "channels",
          "create",
          "--name",
          `camp-buzz-ui-${Date.now()}`,
          "--type",
          "stream",
          "--visibility",
          "open",
        ],
        {
          BUZZ_PRIVATE_KEY: TYLER_SK,
          BUZZ_RELAY_URL: relay,
          BUZZ_AUTH_TAG: "",
        },
      ),
    ) as { channel_id: string };
    const channelId = created.channel_id;

    // Post through camp-buzz (real buzz CLI under the hood)
    run(
      campBuzzBin,
      [
        "post",
        "-m",
        marker,
        "--channel",
        channelId,
        "--relay",
        relay,
        "--festival",
        "UI0001",
        "--task",
        "FEST-ui",
        "--gate",
        "pass",
      ],
      {
        BUZZ_PRIVATE_KEY: TYLER_SK,
        BUZZ_RELAY_URL: relay,
        PATH: `${path.dirname(buzzBin)}:${process.env.PATH ?? ""}`,
      },
    );

    // Open Buzz Desktop e2e build against the live relay
    const { installRelayBridge } = await loadInstallRelayBridge();
    await installRelayBridge(page, "tyler");
    await page.goto("/");

    // Wait for app shell
    await expect(page.getByTestId("app-sidebar")).toBeVisible({
      timeout: 30_000,
    });

    // Open channel browser and join / click our channel if not in list
    // Prefer stream list click by name pattern
    const streamList = page.getByTestId("stream-list");
    await expect(streamList).toBeVisible({ timeout: 30_000 });

    // Channel may appear in sidebar after join; open browser if needed
    const nameMatch = page.getByText(/camp-buzz-ui-/).first();
    if (!(await nameMatch.isVisible().catch(() => false))) {
      // keyboard open channel browser (mod+shift+o)
      const isMac = process.platform === "darwin";
      await page.keyboard.press(
        isMac ? "Meta+Shift+KeyO" : "Control+Shift+KeyO",
      );
      await expect(page.getByTestId("channel-browser-dialog")).toBeVisible({
        timeout: 15_000,
      });
      // click join on matching channel if shown
      const browse = page.locator('[data-testid^="browse-channel-"]').filter({
        hasText: "camp-buzz-ui-",
      });
      if (await browse.count()) {
        await browse
          .first()
          .getByRole("button", { name: /Join|Open/i })
          .click()
          .catch(async () => {
            await browse.first().click();
          });
      }
    }

    // Click channel in stream list
    await page
      .locator('[data-testid^="channel-"]')
      .filter({ hasText: /camp-buzz-ui-/ })
      .first()
      .click({ timeout: 20_000 })
      .catch(async () => {
        // fallback: any channel containing our marker after search
        await page.getByText(/camp-buzz-ui-/).first().click();
      });

    // Assert message body + footer fields in the timeline
    const timeline = page.getByTestId("message-timeline");
    await expect(timeline).toBeVisible({ timeout: 30_000 });
    await expect(timeline).toContainText(marker, { timeout: 30_000 });
    await expect(timeline).toContainText("festival: UI0001");
    await expect(timeline).toContainText("task: FEST-ui");
    await expect(timeline).toContainText("gate: pass");

    // Hold a beat for the video recording
    await page.waitForTimeout(2500);
  });
});
