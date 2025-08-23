<template>
    <UButton label="Logout" color="primary" @click="logout" />
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
    } catch (error) {
        console.error("Error during logout:", error);
    }
};
</script>
