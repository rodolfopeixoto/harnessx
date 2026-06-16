import { test, expect } from "@playwright/test";

test("home page renders", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle(/HarnessX|Harness/i);
});

test("nav reachable", async ({ page }) => {
  await page.goto("/");
  const body = await page.locator("body");
  await expect(body).toBeVisible();
});
