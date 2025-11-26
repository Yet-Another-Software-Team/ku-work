import crypto from "node:crypto";

export default defineEventHandler((event) => {
    // Generate a unique nonce for this request
    const nonce = crypto.randomUUID();

    // Make nonce available to Nuxt renderer
    event.context.nonce = nonce;

    // --- Content Security Policy ---
    // Note:
    // - No unsafe-inline
    // - Turnstile domain allowed
    // - Nuxt will automatically attach nonce to SSR inline scripts
    const csp = [
        `default-src 'self'`,
        `script-src 'self' https://challenges.cloudflare.com 'nonce-${nonce}'`,
        `style-src 'self'`, // no unsafe-inline
        `img-src 'self' data:`,
        `font-src 'self' data:`,
        `connect-src 'self' https://challenges.cloudflare.com`,
        `frame-src 'none'`,
        `frame-ancestors 'none'`,
        `form-action 'self'`,
        `base-uri 'self'`,
        `object-src 'none'`,
        `media-src 'self'`,
        `manifest-src 'self'`,
        `worker-src 'self'`,
        `upgrade-insecure-requests`,
    ].join("; ");

    setResponseHeader(event, "Content-Security-Policy", csp);

    // --- Additional security headers ---
    setResponseHeader(event, "X-Frame-Options", "DENY");
    setResponseHeader(event, "X-Content-Type-Options", "nosniff");
    setResponseHeader(event, "Referrer-Policy", "strict-origin-when-cross-origin");
    setResponseHeader(event, "X-XSS-Protection", "1; mode=block");

    // Permissions-Policy: disable optional APIs
    setResponseHeader(
        event,
        "Permissions-Policy",
        "geolocation=(), microphone=(), camera=(), payment=()"
    );

    // Cross-Origin security (recommended)
    setResponseHeader(event, "Cross-Origin-Opener-Policy", "same-origin");
    setResponseHeader(event, "Cross-Origin-Embedder-Policy", "require-corp");
    setResponseHeader(event, "Cross-Origin-Resource-Policy", "same-origin");
});
