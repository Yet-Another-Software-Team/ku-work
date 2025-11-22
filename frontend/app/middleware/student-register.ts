export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const authStore = useAuthStore();
    const token = authStore.token;
    const role = authStore.role;
    const isRegistered = authStore.isRegistered;

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
