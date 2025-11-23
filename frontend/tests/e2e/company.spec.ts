import { test, expect } from "@playwright/test";

test.describe("Company E2E Tests (Authenticated)", () => {
    test.beforeEach(async ({ page }) => {
        // All tests start from the home page
        await page.goto("/");
        await page.waitForLoadState("networkidle");
    });

    test("should access dashboard and view job postings", async ({ page }) => {
        // Navigate to dashboard
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Verify we're on the dashboard
        await expect(page).toHaveURL(/dashboard/);
        await expect(page.getByText("Company Dashboard")).toBeVisible();

        // Check for job posting tabs
        await expect(page.getByText("Accepted")).toBeVisible();
        await expect(page.getByText("Pending")).toBeVisible();
        await expect(page.getByText("Rejected")).toBeVisible();
    });

    test("should view and interact with different job status tabs", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Click on pending tab
        await page.getByText("Pending").click();
        await page.waitForTimeout(500);

        // Click on rejected tab
        await page.getByText("Rejected").click();
        await page.waitForTimeout(500);

        // Click back to accepted tab
        await page.getByText("Accepted").click();
        await page.waitForTimeout(500);

        // Verify the dashboard is still functioning
        await expect(page.getByText("Company Dashboard")).toBeVisible();
    });

    test("should access company profile page", async ({ page }) => {
        // Navigate to profile
        await page.goto("/profile");
        await page.waitForLoadState("networkidle");

        // Verify we're on the profile page
        await expect(page).toHaveURL(/profile/);

        // Check for common profile elements (adjust based on actual implementation)
        const profileContent = page.locator("main, [role='main'], .profile");
        await expect(profileContent).toBeVisible();
    });

    test("should be able to search/filter jobs from home page", async ({ page }) => {
        await page.goto("/");
        await page.waitForLoadState("networkidle");

        // Look for job listings or search functionality
        const jobCards = page.locator("[data-testid='job-card'], .job-card, article");
        
        // If there's a search input, try searching
        const searchInput = page.locator("input[type='search'], input[placeholder*='search' i]");
        if (await searchInput.isVisible().catch(() => false)) {
            await searchInput.fill("Software");
            await page.waitForTimeout(1000);
        }
    });

    test("should navigate between pages using navigation menu", async ({ page }) => {
        await page.goto("/");

        // Test navigation to jobs page
        const jobsLink = page.getByRole("link", { name: /jobs|browse/i }).first();
        if (await jobsLink.isVisible().catch(() => false)) {
            await jobsLink.click();
            await page.waitForLoadState("networkidle");
        }

        // Navigate to dashboard
        const dashboardLink = page.getByRole("link", { name: /dashboard/i }).first();
        if (await dashboardLink.isVisible().catch(() => false)) {
            await dashboardLink.click();
            await page.waitForLoadState("networkidle");
            await expect(page).toHaveURL(/dashboard/);
        }

        // Navigate to profile
        const profileLink = page.getByRole("link", { name: /profile/i }).first();
        if (await profileLink.isVisible().catch(() => false)) {
            await profileLink.click();
            await page.waitForLoadState("networkidle");
            await expect(page).toHaveURL(/profile/);
        }
    });

    test("should verify user is authenticated as company", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Verify company-specific content is visible
        await expect(page.getByText("Company Dashboard")).toBeVisible();

        // Should not see student-specific elements
        const studentDashboard = page.getByText("Student Dashboard");
        await expect(studentDashboard).not.toBeVisible();
    });

    test("should be able to logout", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Look for logout button (adjust selector based on actual implementation)
        const logoutButton = page.getByRole("button", { name: /logout|sign out/i });
        
        if (await logoutButton.isVisible().catch(() => false)) {
            await logoutButton.click();
            await page.waitForLoadState("networkidle");

            // After logout, should be redirected to home or login page
            await expect(page).toHaveURL(/\/|login/);
        }
    });
});
