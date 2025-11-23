import { defineConfig } from "@playwright/test";

export default defineConfig({
    testDir: "./tests/e2e",
    use: {
        baseURL: "http://localhost:3000", // Nuxt dev server URL
        headless: false,
        viewport: { width: 1280, height: 720 },
        screenshot: "only-on-failure",
        video: "retain-on-failure",
    },
    // Run test server
    webServer: {
        command: "bun run dev",
        port: 3000,
        cwd: ".",
    },
});
