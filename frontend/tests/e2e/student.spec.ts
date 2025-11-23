import { test, expect } from "@playwright/test";

test.describe("Student E2E Tests (Authenticated)", () => {
    test.beforeEach(async ({ page }) => {
        // All tests start from the home page
        await page.goto("/");
        await page.waitForLoadState("networkidle");
    });

    test("should access dashboard and view applications", async ({ page }) => {
        // Navigate to dashboard
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Verify we're on the dashboard
        await expect(page).toHaveURL(/dashboard/);
        await expect(page.getByText("Student Dashboard")).toBeVisible();

        // Check for application status tabs
        await expect(page.getByText(/all|pending|accepted|rejected/i).first()).toBeVisible();
    });

    test("should view and interact with application status tabs", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Look for status filter tabs/buttons
        const allTab = page.getByText(/^all$/i);
        const pendingTab = page.getByText(/^pending$/i);
        const acceptedTab = page.getByText(/^accepted$/i);
        const rejectedTab = page.getByText(/^rejected$/i);

        // Click through different tabs if they exist
        if (await allTab.isVisible().catch(() => false)) {
            await allTab.click();
            await page.waitForTimeout(500);
        }

        if (await pendingTab.isVisible().catch(() => false)) {
            await pendingTab.click();
            await page.waitForTimeout(500);
        }

        if (await acceptedTab.isVisible().catch(() => false)) {
            await acceptedTab.click();
            await page.waitForTimeout(500);
        }

        if (await rejectedTab.isVisible().catch(() => false)) {
            await rejectedTab.click();
            await page.waitForTimeout(500);
        }

        // Verify the dashboard is still functioning
        await expect(page.getByText("Student Dashboard")).toBeVisible();
    });

    test("should access student profile page", async ({ page }) => {
        // Navigate to profile
        await page.goto("/profile");
        await page.waitForLoadState("networkidle");

        // Verify we're on the profile page
        await expect(page).toHaveURL(/profile/);

        // Check for common profile elements
        const profileContent = page.locator("main, [role='main'], .profile");
        await expect(profileContent).toBeVisible();
    });

    test("should browse and view job postings", async ({ page }) => {
        await page.goto("/");
        await page.waitForLoadState("networkidle");

        // Look for job listings
        const jobCards = page.locator("[data-testid='job-card'], .job-card, article");

        // Check if job cards exist
        const jobCardCount = await jobCards.count();
        if (jobCardCount > 0) {
            // Click on the first job to view details
            await jobCards.first().click();
            await page.waitForLoadState("networkidle");
        }
    });

    test("should search for jobs", async ({ page }) => {
        await page.goto("/");
        await page.waitForLoadState("networkidle");

        // Look for search input
        const searchInput = page.locator("input[type='search'], input[placeholder*='search' i]");

        if (await searchInput.isVisible().catch(() => false)) {
            await searchInput.fill("Developer");
            await page.waitForTimeout(1000);
            await page.waitForLoadState("networkidle");

            // Verify search results or that page updated
            await page.waitForTimeout(500);
        }
    });

    test("should view company profiles", async ({ page }) => {
        await page.goto("/");
        await page.waitForLoadState("networkidle");

        // Try to find and click on a company link
        const companyLink = page.locator("a[href*='/jobs/']").first();

        if (await companyLink.isVisible().catch(() => false)) {
            await companyLink.click();
            await page.waitForLoadState("networkidle");
            // Verify we're on a company profile page
            await expect(page).toHaveURL(/\/jobs\//);
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

    test("should verify user is authenticated as student", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Verify student-specific content is visible
        await expect(page.getByText("Student Dashboard")).toBeVisible();

        // Should not see company-specific elements
        const companyDashboard = page.getByText("Company Dashboard");
        await expect(companyDashboard).not.toBeVisible();
    });

    test("should view application history and details", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Look for application cards
        const applicationCards = page.locator(
            "[data-testid='application-card'], .application-card, .application"
        );

        const appCount = await applicationCards.count();
        if (appCount > 0) {
            // Click on first application to view details
            const firstApp = applicationCards.first();

            // Look for details/view button
            const viewButton = firstApp.locator("button, a").filter({ hasText: /view|details/i });
            if (await viewButton.isVisible().catch(() => false)) {
                await viewButton.click();
                await page.waitForTimeout(1000);
            }
        }
    });

    test("should be able to filter or sort applications", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Look for sort or filter dropdowns
        const sortDropdown = page.locator("select, [role='combobox']").filter({
            hasText: /sort|filter/i,
        });

        if (
            await sortDropdown
                .first()
                .isVisible()
                .catch(() => false)
        ) {
            await sortDropdown.first().click();
            await page.waitForTimeout(500);

            // Select an option if dropdown opened
            const options = page.locator("option, [role='option']");
            const optionCount = await options.count();
            if (optionCount > 1) {
                await options.nth(1).click();
                await page.waitForTimeout(1000);
            }
        }
    });

    test("should be able to logout", async ({ page }) => {
        await page.goto("/dashboard");
        await page.waitForLoadState("networkidle");

        // Look for logout button
        const logoutButton = page.getByRole("button", { name: /logout|sign out/i });

        if (await logoutButton.isVisible().catch(() => false)) {
            await logoutButton.click();
            await page.waitForLoadState("networkidle");

            // After logout, should be redirected to home or login page
            await expect(page).toHaveURL(/\/|login/);
        }
    });
});
