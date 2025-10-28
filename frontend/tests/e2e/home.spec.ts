import { test, expect } from "@playwright/test";

test("homepage", async ({ page }) => {
    await page.goto("/"); // goes to http://localhost:3000/
    await page.waitForLoadState("networkidle");

    // Check the quote on the homepage
    await expect(page.locator("h2#quote")).toContainText("Great job awaits for you");
});
