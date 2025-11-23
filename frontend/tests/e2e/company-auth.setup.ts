import { test as setup, expect } from "@playwright/test";

const authFile = "./playwright/.auth/company.json";

function generateRandomString(length: number): string {
    const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";

    for (let i = 0; i < length; i++) {
        const randomIndex = Math.floor(Math.random() * characters.length);
        result += characters.charAt(randomIndex);
    }

    return result;
}
const uniqueEmail = `testcompany+${generateRandomString(8)}@kuwork.com`;

setup("authenticate as company", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    // Navigate to company login
    await page.getByText("Company").click();
    await page.waitForTimeout(500);

    // Click sign up
    page.getByRole("link", { name: /Sign Up/i }).click();

    // Fill in Account Signup
    const emailInput = page.getByRole("textbox", { name: /Username/i });
    await emailInput.waitFor({ state: "visible" });
    await emailInput.fill(uniqueEmail);
    await emailInput.press("Enter");

    // Wait for password field and fill
    const passwordInput = page.getByRole("textbox", { name: /Password/i });
    await passwordInput.waitFor({ state: "visible", timeout: 10000 });
    await passwordInput.fill("TestPassword123");
    await passwordInput.press("Enter");

    // Wait for successful login and redirect
    await page.waitForURL(/dashboard|home/, { timeout: 15000 });

    // Verify we're logged in
    await expect(page.getByText(/dashboard|profile|logout/i).first()).toBeVisible({
        timeout: 10000,
    });

    // Save authentication state
    await page.context().storageState({ path: authFile });
});
