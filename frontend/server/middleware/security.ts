export default defineEventHandler((event) => {
    // Anti-clickjacking headers
    // X-Frame-Options: Prevents your site from being embedded in iframes
    setResponseHeader(event, "X-Frame-Options", "DENY");

    // Content-Security-Policy: Comprehensive policy for security
    // Note: Using nonces would be ideal, but for Nuxt SSR compatibility, we need unsafe-inline
    // Consider implementing nonces in production for better security
    const cspDirectives = [
        "default-src 'self'", // Default fallback for unspecified directives
        // Allow self, inline scripts with hash, and Turnstile scripts
        "script-src 'self' 'unsafe-inline' https://challenges.cloudflare.com",
        // Allow styles from self and inline styles for Nuxt
        "style-src 'self' 'unsafe-inline'",
        // Restrict images to self, data URIs, and specific CDN domains (avoid wildcard https:)
        "img-src 'self' data:",
        "font-src 'self' data:",
        // Allow API connections to self and Turnstile
        "connect-src 'self' https://challenges.cloudflare.com",
        // No iframes allowed
        "frame-src 'none'",
        "frame-ancestors 'none'", // Prevent embedding (anti-clickjacking)
        "form-action 'self'", // Forms can only submit to same origin
        "base-uri 'self'", // Restrict base tag URLs
        "object-src 'none'", // No plugins (Flash, Java, etc.)
        "media-src 'self'", // Media from self only
        "manifest-src 'self'", // Web manifest from self only
        "worker-src 'self'", // Web workers from self only
        "upgrade-insecure-requests", // Upgrade HTTP to HTTPS
    ];
    setResponseHeader(event, "Content-Security-Policy", cspDirectives.join("; "));

    // Additional security headers
    // X-Content-Type-Options: Prevents MIME type sniffing
    setResponseHeader(event, "X-Content-Type-Options", "nosniff");

    // Referrer-Policy: Controls how much referrer information is shared
    setResponseHeader(event, "Referrer-Policy", "strict-origin-when-cross-origin");

    // X-XSS-Protection: Legacy XSS protection (mostly deprecated but still useful for older browsers)
    setResponseHeader(event, "X-XSS-Protection", "1; mode=block");

    // Permissions-Policy: Controls browser features and APIs
    setResponseHeader(
        event,
        "Permissions-Policy",
        "geolocation=(), microphone=(), camera=(), payment=()"
    );
});
