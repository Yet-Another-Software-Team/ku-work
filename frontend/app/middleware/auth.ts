export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const token = localStorage.getItem("token");

    if (!token) {
        return navigateTo("/", { replace: true });
    }
});
