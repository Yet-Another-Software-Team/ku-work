// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    compatibilityDate: "2025-07-15",
    devtools: { enabled: true },

    modules: ["@nuxt/image", "@nuxt/eslint", "@nuxt/test-utils", "@nuxt/scripts", "@nuxt/ui"],
    css: ["~/assets/css/main.css"],
    runtimeConfig: {
        public: {
            apiBaseUrl: process.env.API_BASE_URL || "http://localhost:8000",
            googleClientId: process.env.GOOGLE_CLIENT_ID,
            testCompanyEmail: process.env.TEST_COMPANY_EMAIL,
            testStudentEmail: process.env.TEST_STUDENT_EMAIL,
        },
    },
    nitro: {
        preset: "bun",
    },
});
