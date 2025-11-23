import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
    testDir: "./tests/e2e",
    fullyParallel: true,
    reporter: "html",
    use: {
        baseURL: "http://localhost:3000", // Nuxt dev server URL
        headless: false,
        viewport: { width: 1280, height: 720 },
        screenshot: "only-on-failure",
        video: "retain-on-failure",
        trace: "on-first-retry",
    },
    projects: [
        // Setup projects
        { name: "setup-company", testMatch: /.*company-auth\.setup\.ts/ },

        // Company tests using authenticated state
        {
            name: "company",
            testMatch: /.*company\.spec\.ts/,
            use: {
                ...devices["Desktop Chrome"],
                storageState: "./playwright/.auth/company.json",
            },
        },

        // Other tests without authentication
        {
            name: "other",
            testIgnore: [/.*company\.spec\.ts/, /.*student\.spec\.ts/],
            use: { ...devices["Desktop Chrome"] },
        },
    ],
    // Run test server
    webServer: {
        command: "bun run dev",
        port: 3000,
        cwd: ".",
        reuseExistingServer: true,
    },
});
