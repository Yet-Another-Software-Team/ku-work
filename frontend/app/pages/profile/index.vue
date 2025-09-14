<template>
    <div class="relative">
        <!-- Loading screen -->
        <div v-if="loading" class="fixed inset-0 flex items-center justify-center bg-white z-50">
            <div
                class="h-10 w-10 animate-spin rounded-full border-4 border-gray-300 border-t-primary"
            ></div>
        </div>

        <!-- Content -->
        <ProfileCard v-if="!isCompany" />
        <CompanyProfileCard v-if="isCompany" />
    </div>
</template>

<script setup lang="ts">
const isCompany = ref(false);
const isViewer = ref(true);
const loading = ref(true);

onMounted(() => {
    const role = localStorage.getItem("role");

    if (role === "company") {
        isCompany.value = true;
        isViewer.value = false;
    }
    if (role === "student") {
        isCompany.value = false;
        isViewer.value = false;
    }
    loading.value = false;
});

definePageMeta({
    layout: "viewer",
});
</script>
