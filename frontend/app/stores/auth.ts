import { defineStore } from "pinia";
import { ref, computed } from "vue";

export const useAuthStore = defineStore("auth", () => {
    const token = ref<string | null>(null);
    const username = ref<string | null>(null);
    const role = ref<string | null>(null);
    const userId = ref<string | null>(null);
    const isRegistered = ref<boolean>(false);

    // Computed properties for easy access
    const isAuthenticated = computed(() => !!token.value);
    const isAdmin = computed(() => role.value === "admin");
    const isCompany = computed(() => role.value === "company");
    const isStudent = computed(() => role.value === "student");
    const isViewer = computed(() => role.value === "viewer");

    // Action to set auth data
    const setAuthData = (data: {
        token: string;
        username: string;
        role: string;
        userId?: string;
        isRegistered?: boolean;
    }) => {
        token.value = data.token;
        username.value = data.username;
        role.value = data.role;
        if (data.userId) {
            userId.value = data.userId;
        }
        if (data.isRegistered !== undefined) {
            isRegistered.value = data.isRegistered;
        }
    };

    // Action to update token (for refresh)
    const updateToken = (newToken: string) => {
        token.value = newToken;
    };

    // Action to set isRegistered
    const setIsRegistered = (value: boolean) => {
        isRegistered.value = value;
    };

    // Action to logout
    const logout = () => {
        token.value = null;
        username.value = null;
        role.value = null;
        userId.value = null;
        isRegistered.value = false;
    };

    // Action to clear all data
    const clearAuthData = () => {
        logout();
    };

    return {
        // State
        token,
        username,
        role,
        userId,
        isRegistered,
        // Computed
        isAuthenticated,
        isAdmin,
        isCompany,
        isStudent,
        isViewer,
        // Actions
        setAuthData,
        updateToken,
        setIsRegistered,
        logout,
        clearAuthData,
    };
});
