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

const toast = useToast();

const logout = async () => {
    try {
        const api = useApi();
        await api.post(
            "/logout",
            {},
            {
                withCredentials: true,
            }
        );

        localStorage.removeItem("token");
        localStorage.removeItem("username");
        localStorage.removeItem("role");

        toast.add({
            title: "Logged Out",
            description: "You have been successfully logged out.",
            color: "neutral",
        });

        // Redirect to the home page after logout
        window.location.href = "/";
    } catch (error) {
        console.error("Error during logout:", error);
    }
};
</script>
