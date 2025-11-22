import type { Axios } from "axios";
import { jwtDecode } from "jwt-decode";
import { useAuthStore } from "~/stores/auth";

export default defineNuxtPlugin({
    setup: (nuxtApp) => {
        const { $axios } = nuxtApp;
        let refreshTimeoutId: NodeJS.Timeout | null = null;

        const scheduleTokenRefresh = () => {
            if (import.meta.server) {
                return;
            }

            const authStore = useAuthStore();
            const token = authStore.token;

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
                            authStore.updateToken(newToken);
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
            const authStore = useAuthStore();
            if (import.meta.server) return;

            const role = authStore.role;
            authStore.logout();

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

        const authenticateWithToken = async () => {
            const authStore = useAuthStore();
            try {
                const response = await ($axios as Axios).get(
                    `${useRuntimeConfig().public.apiBaseUrl}/me`,
                    {
                        withCredentials: true,
                    }
                );

                if (response.data) {
                    authStore.setAuthData({
                        token: authStore.token as string,
                        username: response.data.username,
                        role: response.data.role,
                        userId: response.data.userId,
                        isRegistered: response.data.isRegistered,
                    });
                    scheduleTokenRefresh();

                    // Successfully authenticated, redirect to dashboard
                    if (authStore.isAdmin) {
                        navigateTo("/admin", { replace: true });
                    } else if (authStore.isStudent || authStore.isViewer) {
                        navigateTo("/jobs", { replace: true });
                    } else if (authStore.isCompany) {
                        navigateTo("/dashboard", { replace: true });
                    }
                }
                // eslint-disable-next-line @typescript-eslint/no-unused-vars
            } catch (error) {
                // To handle error if the token is invalid or expired
                // so it can clear the store and redirect to login page
                logout();
            }
        };

        // Initial call to authenticate when the app loads
        authenticateWithToken();

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
