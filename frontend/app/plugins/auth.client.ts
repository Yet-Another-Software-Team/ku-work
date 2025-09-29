import { jwtDecode } from "jwt-decode";

export default defineNuxtPlugin((nuxtApp) => {
    const { $axios } = nuxtApp;
    let refreshTimeoutId: NodeJS.Timeout | null = null;

    const scheduleTokenRefresh = () => {
        if (process.server) {
            return;
        }

        const token = localStorage.getItem("token");
        if (!token) {
            return;
        }

        try {
            const decodedToken = jwtDecode<{ exp: number }>(token);
            const expirationTime = decodedToken.exp * 1000;
            const now = Date.now();
            const timeUntilExpiration = expirationTime - now;

            // Schedule refresh 1 minute before expiration, or immediately if it's closer than that.
            const refreshTime = Math.max(timeUntilExpiration - 60 * 1000, 1);

            if (refreshTimeoutId) {
                clearTimeout(refreshTimeoutId);
            }

            refreshTimeoutId = setTimeout(async () => {
                try {
                    const response = await $axios.post("/api/auth/refresh");
                    const newToken = response.data.token;
                    if (newToken) {
                        localStorage.setItem("token", newToken);
                        scheduleTokenRefresh(); // Schedule the next refresh
                    } else {
                        logout();
                    }
                } catch (error) {
                    console.error("Failed to refresh token:", error);
                    logout();
                }
            }, refreshTime);
        } catch (error) {
            console.error("Invalid token:", error);
            logout();
        }
    };

    const logout = () => {
        if (process.server) return;
        localStorage.removeItem("token");
        localStorage.removeItem("username");
        localStorage.removeItem("role");

        if (refreshTimeoutId) {
            clearTimeout(refreshTimeoutId);
        }

        const toast = useToast();
        toast.add({
            title: "Session Expired",
            description: "Please log in again",
            color: "error",
        });

        // Redirect to login page
        navigateTo("/");
    };

    // Initial call to schedule refresh when the app loads
    scheduleTokenRefresh();

    // Provide the logout function to be accessible globally
    nuxtApp.provide("logout", logout);
});
