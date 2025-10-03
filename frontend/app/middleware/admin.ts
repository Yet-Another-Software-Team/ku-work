export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const token = localStorage.getItem("token");
    const role = localStorage.getItem("role");

    if (!token) {
        return navigateTo("/", { replace: true });
    }

    if (role !== "admin") {
        throw createError({
            statusCode: 403,
            statusMessage: "Admin access required",
        });
    }
});
