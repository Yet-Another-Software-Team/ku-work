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
        const { startRequest, endRequest, forceComplete } = useApiLoading();

        // Create axios instance
        const axiosInstance: AxiosInstance = axios.create({
            baseURL: config.public.apiBaseUrl,
            timeout: 10000,
            headers: {
                "Content-Type": "application/json",
            },
            withCredentials: true,
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
                if (error.config?.metadata?.requestId) {
                    endRequest(error.config.metadata.requestId);
                }
                return Promise.reject(error);
            }
        );

        // Response interceptor to handle token refresh and end loading
        axiosInstance.interceptors.response.use(
            (response) => {
                if (response.config?.metadata?.requestId) {
                    endRequest(response.config.metadata.requestId);
                }
                return response;
            },
            async (error: ApiError) => {
                const originalRequest = error.config as InternalAxiosRequestConfig & {
                    _retry?: boolean;
                };

                if (originalRequest?.metadata?.requestId) {
                    endRequest(originalRequest.metadata.requestId);
                }

                if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
                    if (isRefreshing) {
                        return new Promise((resolve, reject) => {
                            failedQueue.push({ resolve, reject });
                        })
                            .then((token) => {
                                originalRequest.headers.Authorization = `Bearer ${token}`;
                                return axiosInstance(originalRequest);
                            })
                            .catch((err) => {
                                return Promise.reject(err);
                            });
                    }

                    originalRequest._retry = true;
                    isRefreshing = true;

                    try {
                        // Start a new request for the refresh operation
                        const refreshRequestId = startRequest("/auth/refresh", "POST");
                        // Try to refresh the token
                        const response = await axios.post(
                            `${config.public.apiBaseUrl}/auth/refresh`,
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
                            if (response.data.userId) {
                                localStorage.setItem("userId", response.data.userId);
                            }
                            if (response.data.username) {
                                localStorage.setItem("username", response.data.username);
                            }
                            if (response.data.role) {
                                localStorage.setItem("role", response.data.role);
                            }
                        }

                        axiosInstance.defaults.headers.common["Authorization"] =
                            `Bearer ${newToken}`;
                        originalRequest.headers.Authorization = `Bearer ${newToken}`;

                        processQueue(null, newToken);

                        return axiosInstance(originalRequest);
                    } catch (refreshError: any) {
                        processQueue(refreshError, null);
                        forceComplete();

                        if (import.meta.client) {
                            localStorage.removeItem("token");
                            localStorage.removeItem("username");
                            localStorage.removeItem("role");
                        }
                        return Promise.reject(refreshError);
                    } finally {
                        isRefreshing = false;
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
