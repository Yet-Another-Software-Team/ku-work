export default defineNuxtRouteMiddleware((_to, _from) => {
    // Skip middleware on server-side rendering
    if (import.meta.server) return;

    const token = localStorage.getItem("token");
    const role = localStorage.getItem("role");

    if (!token) {
        return navigateTo("/", { replace: true });
    }

    if (role !== "student") {
        if (role === "company") {
            return navigateTo("/dashboard", { replace: true });
        } else if (role === "admin") {
            return navigateTo("/admin/dashboard", { replace: true });
        } else if (role === "viewer") {
            return navigateTo("/jobs", { replace: true });
        } else {
            return navigateTo("/", { replace: true });
        }
    }
});
