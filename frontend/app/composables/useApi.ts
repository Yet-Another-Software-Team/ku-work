/* eslint-disable @typescript-eslint/no-explicit-any */
import type { AxiosRequestConfig, AxiosResponse } from "axios";

export interface ApiResponse<T = any> {
    data: T;
    status: number;
    message?: string;
}

export interface ApiError {
    status: number;
    message: string;
    data?: any;
}

// Extend AxiosRequestConfig to include our custom properties
interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
    skipGlobalLoading?: boolean;
    metadata?: {
        requestId?: string;
    };
}

export const useApi = () => {
    const { $axios } = useNuxtApp();
    const toast = useToast();
    const { startRequest, endRequest } = useApiLoading();

    const handleError = (error: any): ApiError => {
        let apiError: ApiError = {
            status: 500,
            message: "An unexpected error occurred",
        };

        if (error.response) {
            apiError = {
                status: error.response.status,
                message:
                    error.response.data?.error || error.response.data?.message || error.message,
                data: error.response.data,
            };
        } else if (error.request) {
            apiError = {
                status: 0,
                message: "Network error - please check your connection",
            };
        } else {
            apiError = {
                status: 500,
                message: error.message || "An unexpected error occurred",
            };
        }

        return apiError;
    };

    const get = async <T = any>(
        url: string,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.get(url, config);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    const post = async <T = any>(
        url: string,
        data?: any,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.post(url, data, config);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    const put = async <T = any>(
        url: string,
        data?: any,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.put(url, data, config);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    const patch = async <T = any>(
        url: string,
        data?: any,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.patch(url, data, config);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    const del = async <T = any>(
        url: string,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.delete(url, config);
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    // Helper for form data uploads
    const postFormData = async <T = any>(
        url: string,
        formData: FormData,
        config?: ExtendedAxiosRequestConfig
    ): Promise<ApiResponse<T>> => {
        try {
            const response: AxiosResponse<T> = await $axios.post(url, formData, {
                ...config,
                headers: {
                    ...config?.headers,
                    "Content-Type": "multipart/form-data",
                },
            });
            return {
                data: response.data,
                status: response.status,
            };
        } catch (error) {
            throw handleError(error);
        }
    };

    // Helper method to show error toast
    const showErrorToast = (error: ApiError, title?: string) => {
        toast.add({
            title: title || "Error",
            description: error.message,
            color: "error",
        });
    };

    // Helper method to show success toast
    const showSuccessToast = (message: string, title?: string) => {
        toast.add({
            title: title || "Success",
            description: message,
            color: "success",
        });
    };

    return {
        get,
        post,
        put,
        patch,
        delete: del,
        postFormData,
        handleError,
        showErrorToast,
        showSuccessToast,
    };
};

// Type definitions for common API responses
export interface LoginResponse {
    token: string;
    username: string;
    isCompany: boolean;
    isStudent: boolean;
}

export interface AuthResponse {
    token?: string;
    username?: string;
    isCompany?: boolean;
    isStudent?: boolean;
    message?: string;
}
