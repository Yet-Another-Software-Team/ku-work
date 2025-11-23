import { jwtDecode } from "jwt-decode";

export default defineNuxtPlugin({
    setup: (nuxtApp) => {
        const { $axios } = nuxtApp;
        let refreshTimeoutId: NodeJS.Timeout | null = null;

        const scheduleTokenRefresh = () => {
            if (import.meta.server) {
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
                        const response = await (
                            $axios as {
                                post: (
                                    url: string,
                                    data: unknown | undefined,
                                    options: { withCredentials: boolean }
                                ) => Promise<{ data: { token?: string } }>;
                            }
                        ).post(`${useRuntimeConfig().public.apiBaseUrl}/auth/refresh`, undefined, {
                            withCredentials: true,
                        });
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
            const role = localStorage.getItem("role");
            if (import.meta.server) return;
            localStorage.removeItem("token");
            localStorage.removeItem("username");
            localStorage.removeItem("role");

            if (refreshTimeoutId) {
                clearTimeout(refreshTimeoutId);
            }

            // Redirect to login page
            if (role === "admin") {
                navigateTo("/admin", { replace: true });
            } else {
                navigateTo("/", { replace: true });
            }
        };

        // Initial call to schedule refresh when the app loads
        scheduleTokenRefresh();

        return {
            provide: {
                // Provide the logout function to be accessible globally
                logout: logout,
            },
        };
    },
    // Depends on axios client, prevents can't find $axios error
    dependsOn: ["axios-client"],
});
