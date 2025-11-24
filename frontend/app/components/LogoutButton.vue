<template>
    <UButton
        label="Logout"
        variant="ghost"
        size="xl"
        icon="ic:round-logout"
        :ui="{ base: 'justify-start text-left text-white hover:bg-white/10 cursor-pointer' }"
        @click="logout"
    />
</template>

<script setup lang="ts">
import { useApi } from "~/composables/useApi";
import { useAuthStore } from "~/stores/auth";

const toast = useToast();

const logout = async () => {
    try {
        const api = useApi();
        await api.post(
            "/auth/logout",
            {},
            {
                withCredentials: true,
            }
        );

        const authStore = useAuthStore();
        authStore.logout();

        // Redirect to the home page after logout
        window.location.href = "/";
    } catch (error) {
        console.error("Error during logout:", error);
    }
};
</script>
