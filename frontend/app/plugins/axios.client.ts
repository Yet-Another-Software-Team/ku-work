/* eslint-disable @typescript-eslint/no-explicit-any */
import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from "axios";

interface ApiError extends AxiosError {
    status?: number;
}

declare module "#app" {
    interface NuxtApp {
        $axios: AxiosInstance;
    }
}

declare module "vue" {
    interface ComponentCustomProperties {
        $axios: AxiosInstance;
    }
}

// Extend axios config to include metadata
declare module "axios" {
    export interface InternalAxiosRequestConfig {
        metadata?: {
            requestId: string;
        };
    }
}

export default defineNuxtPlugin({
    name: "axios-client",
    setup: () => {
        const config = useRuntimeConfig();
        const toast = useToast();
        const { startRequest, endRequest, forceComplete } = useApiLoading();

        // Create axios instance
        const axiosInstance: AxiosInstance = axios.create({
            baseURL: config.public.apiBaseUrl,
            timeout: 10000,
            headers: {
                "Content-Type": "application/json",
            },
        });

        // Flag to prevent multiple refresh attempts
        let isRefreshing = false;
        let failedQueue: Array<{
            resolve: (token: string) => void;
            reject: (error: any) => void;
        }> = [];

        const processQueue = (error: any, token: string | null = null) => {
            failedQueue.forEach(({ resolve, reject }) => {
                if (error) {
                    reject(error);
                } else {
                    resolve(token!);
                }
            });

            failedQueue = [];
        };

        // Request interceptor to add auth token and start loading
        axiosInstance.interceptors.request.use(
            (config) => {
                const token = import.meta.client ? localStorage.getItem("token") : null;
                if (token) {
                    config.headers.Authorization = `Bearer ${token}`;
                }

                // Start tracking this request for loading state
                const requestId = startRequest(config.url || "", config.method || "GET");
                config.metadata = { requestId };

                return config;
            },
            (error) => {
                // End loading on request error
                if (error.config?.metadata?.requestId) {
                    endRequest(error.config.metadata.requestId);
                }
                return Promise.reject(error);
            }
        );

        // Response interceptor to handle token refresh and end loading
        axiosInstance.interceptors.response.use(
            (response) => {
                // End loading on successful response
                if (response.config?.metadata?.requestId) {
                    endRequest(response.config.metadata.requestId);
                }
                return response;
            },
            async (error: ApiError) => {
                // End loading on error (will be restarted if request is retried)
                if (error.config?.metadata?.requestId) {
                    endRequest(error.config.metadata.requestId);
                }
                const originalRequest = error.config as InternalAxiosRequestConfig & {
                    _retry?: boolean;
                    metadata?: { requestId: string };
                };

                if (
                    (error.response?.status === 401 || error.response?.status === 403) &&
                    originalRequest &&
                    !originalRequest._retry
                ) {
                    if (isRefreshing) {
                        // If already refreshing, queue this request
                        return new Promise((resolve, reject) => {
                            failedQueue.push({ resolve, reject });
                        })
                            .then((token) => {
                                if (originalRequest) {
                                    originalRequest.headers.Authorization = `Bearer ${token}`;
                                    return axiosInstance(originalRequest);
                                }
                            })
                            .catch((err) => Promise.reject(err));
                    }

                    originalRequest._retry = true;
                    isRefreshing = true;

                    try {
                        // Start a new request for the refresh operation
                        const refreshRequestId = startRequest("/refresh", "POST");
                        // Try to refresh the token
                        const response = await axios.post(
                            `${config.public.apiBaseUrl}/refresh`,
                            {},
                            {
                                withCredentials: true,
                                headers: {
                                    "Content-Type": "application/json",
                                },
                            }
                        );

                        const newToken = response.data.token;

                        // End the refresh request
                        endRequest(refreshRequestId);

                        if (import.meta.client) {
                            localStorage.setItem("token", newToken);
                        }

                        // Update the authorization header for the original request
                        if (originalRequest) {
                            originalRequest.headers.Authorization = `Bearer ${newToken}`;
                            // Create new request ID for retry
                            const retryRequestId = startRequest(
                                originalRequest.url || "",
                                originalRequest.method || "GET"
                            );
                            originalRequest.metadata = { requestId: retryRequestId };
                        }

                        processQueue(null, newToken);
                        isRefreshing = false;

                        // Retry the original request with new token
                        return axiosInstance(originalRequest);
                    } catch (refreshError) {
                        // Refresh failed, redirect to login
                        processQueue(refreshError, null);
                        isRefreshing = false;

                        // Force complete all loading states
                        forceComplete();

                        if (import.meta.client) {
                            localStorage.removeItem("token");
                            localStorage.removeItem("username");
                            localStorage.removeItem("role");

                            toast.add({
                                title: "Session Expired",
                                description: "Please log in again",
                                color: "error",
                            });

                            await navigateTo("/", { replace: true });
                        }

                        return Promise.reject(refreshError);
                    }
                }

                return Promise.reject(error);
            }
        );

        return {
            provide: {
                axios: axiosInstance,
            },
        };
    },
});
