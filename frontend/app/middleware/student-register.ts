export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const token = localStorage.getItem("token");
    const role = localStorage.getItem("role");
    const isRegistered = localStorage.getItem("isRegistered") === "true";

    if (!token) {
        return navigateTo("/", { replace: true });
    }

    // Only viewers can register as students
    if (role !== "viewer") {
        return navigateTo("/dashboard", { replace: true });
    }

    // If already registered, redirect to jobs page
    if (isRegistered) {
        throw createError({
            statusCode: 403,
            statusMessage: "Already registered",
        });
    }
});
