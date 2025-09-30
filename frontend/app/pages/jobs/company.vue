<template>
    <div>
        <!-- Role-based rendering -->
        <div v-if="userRole === 'student' || userRole === 'viewer'">
            <CompanyProfileCard :is-owner="false" :is-viewer="userRole === 'viewer'" />
        </div>

        <!-- Access denied for other roles -->
        <div v-else class="text-center py-8">
            <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">Access Denied</h2>
            <p class="text-gray-500 mt-2">You don't have permission to view this page.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
const userRole = ref<string>("viewer");

definePageMeta({
    layout: "viewer",
    middleware: "viewer",
});

onMounted(() => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
});
</script>
