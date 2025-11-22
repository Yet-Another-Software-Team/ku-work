// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    compatibilityDate: "2025-07-15",
    devtools: { enabled: true },

    modules: ["@nuxt/image", "@nuxt/eslint", "@nuxt/test-utils", "@nuxt/scripts", "@nuxt/ui"],
    css: ["~/assets/css/main.css"],
    runtimeConfig: {
        public: {
            apiBaseUrl: process.env.API_BASE_URL,
            googleClientId: process.env.GOOGLE_CLIENT_ID,
            turnstileClientToken: process.env.TURNSTILE_CLIENT_TOKEN,
        },
    },
    nitro: {
        preset: "bun",
    },
    app: {
        head: {
            script: [
                {
                    src: "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit",
                    async: true,
                    defer: true,
                },
            ],
        },
    },
});
