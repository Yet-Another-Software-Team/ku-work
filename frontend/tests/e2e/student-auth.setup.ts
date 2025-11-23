import { test as setup, expect } from "@playwright/test";

const authFile = "./playwright/.auth/student.json";

setup("authenticate as student", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    // Click login with Google
    const [popup] = await Promise.all([
        page.waitForEvent("popup"),
        page.getByRole("button", { name: /continue with google/i }).click(),
    ]);

    // Handle Google OAuth popup
    await popup.waitForLoadState("domcontentloaded");

    // Fill in Google credentials
    const emailInput = popup.getByRole("textbox", { name: /email or phone/i });
    await emailInput.waitFor({ state: "visible" });
    await emailInput.fill(process.env.TEST_STUDENT_EMAIL);
    await emailInput.press("Enter");

    // Wait for password field and fill
    const passwordInput = popup.getByRole("textbox", { name: /enter your password/i });
    await passwordInput.waitFor({ state: "visible", timeout: 10000 });
    await passwordInput.fill(process.env.TEST_STUDENT_PASSWORD);
    await passwordInput.press("Enter");

    // Wait for the popup to close (auth complete)
    await popup.waitForEvent("close", { timeout: 15000 }).catch(() => {
        // If it doesn't close automatically, that's okay
    });

    // Wait for successful login and redirect
    await page.waitForURL(/dashboard|home/, { timeout: 15000 });

    // Verify we're logged in
    await expect(page.getByText(/dashboard|profile|logout/i).first()).toBeVisible({
        timeout: 10000,
    });

    // Save authentication state
    await page.context().storageState({ path: authFile });
});
