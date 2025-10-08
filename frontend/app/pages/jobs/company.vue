<template>
    <div>
        <!-- Role-based rendering -->
        <div v-if="['student', 'viewer'].includes(userRole)" class="pt-5 pb-2">
            <CompanyProfileCard :is-owner="false" :company-id="companyId" />
        </div>

        <!-- Access denied for other roles -->
        <div v-else class="text-center py-8">
            <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">Access Denied</h2>
            <p class="text-gray-500 mt-2">You don't have permission to view this page.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const userRole = ref<string>("viewer");
const route = useRoute();
const companyId = computed(() => route.query.id as string);

definePageMeta({
    layout: "viewer",
    middleware: "student-or-viewer",
});

onMounted(() => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
});
</script>
