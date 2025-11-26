export default defineEventHandler((event) => {
    // Anti-clickjacking headers
    // X-Frame-Options: Prevents your site from being embedded in iframes
    setResponseHeader(event, "X-Frame-Options", "DENY");

    // Content-Security-Policy frame-ancestors: Modern replacement for X-Frame-Options
    // 'none' prevents the page from being embedded in any frame/iframe
    setResponseHeader(event, "Content-Security-Policy", "frame-ancestors 'none'");

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
