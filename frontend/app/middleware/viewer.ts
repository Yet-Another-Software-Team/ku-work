export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const token = localStorage.getItem("token");
    const role = localStorage.getItem("role");

    if (!token) {
        return navigateTo("/", { replace: true });
    }

    // Allow viewer, student, or company (any authenticated user)
    if (!role || !["viewer", "student", "company"].includes(role)) {
        throw createError({
            statusCode: 403,
            statusMessage: "Access denied",
        });
    }

    if (role !== "viewer") {
        if (role === "company" || role === "student") {
            return navigateTo("/dashboard", { replace: true });
        } else if (role === "admin") {
            return navigateTo("/admin/dashboard", { replace: true });
        } else {
            return navigateTo("/", { replace: true });
        }
    }
});
