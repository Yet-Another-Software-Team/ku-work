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

export default defineNuxtPlugin(() => {
    const config = useRuntimeConfig();
    const toast = useToast();

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

    // Request interceptor to add auth token
    axiosInstance.interceptors.request.use(
        (config) => {
            const token = import.meta.client ? localStorage.getItem("jwt_token") : null;
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        },
        (error) => Promise.reject(error)
    );

    // Response interceptor to handle token refresh
    axiosInstance.interceptors.response.use(
        (response) => response,
        async (error: ApiError) => {
            const originalRequest = error.config as InternalAxiosRequestConfig & {
                _retry?: boolean;
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

                    if (import.meta.client) {
                        localStorage.setItem("jwt_token", newToken);
                    }

                    if (originalRequest) {
                        originalRequest.headers.Authorization = `Bearer ${newToken}`;
                    }

                    processQueue(null, newToken);
                    isRefreshing = false;

                    // Retry the original request with new token
                    return axiosInstance(originalRequest);
                } catch (refreshError) {
                    // Refresh failed, redirect to login
                    processQueue(refreshError, null);
                    isRefreshing = false;

                    if (import.meta.client) {
                        localStorage.removeItem("jwt_token");
                        localStorage.removeItem("username");
                        localStorage.removeItem("role");

                        toast.add({
                            title: "Session Expired",
                            description: "Please log in again",
                            color: "error",
                        });

                        await navigateTo("/");
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
});
