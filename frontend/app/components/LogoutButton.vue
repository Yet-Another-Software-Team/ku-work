<template>
    <UButton
        label="Logout"
        @click="logout"
        variant="ghost"
        size="xl"
        icon="ic:round-logout"
        :ui="{ base: 'justify-start text-left text-white hover:bg-white/10' }"
    />
</template>

<script setup lang="ts">
const config = useRuntimeConfig();
const toast = useToast();

const logout = async () => {
    try {
        const _ = await $fetch("/logout", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            credentials: "include",
        });
        localStorage.removeItem("jwt_token");

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
