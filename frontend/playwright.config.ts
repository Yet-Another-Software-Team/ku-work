import { defineConfig } from "@playwright/test";

export default defineConfig({
    testDir: "./tests/e2e",
    use: {
        baseURL: "http://localhost:3000", // Nuxt dev server URL
        headless: false, // set to false to see the browser
        viewport: { width: 1280, height: 720 },
        screenshot: "only-on-failure",
        video: "retain-on-failure",
    },
    webServer: {
        command: "bun run dev", // automatically start Nuxt before tests
        port: 3000,
        reuseExistingServer: !process.env.CI, // speeds up local runs
        cwd: ".",
    },
});
