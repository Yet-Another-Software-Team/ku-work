<template>
    <div class="relative">
        <ProfileCard v-if="!isCompany" />
        <CompanyProfileCard v-if="isCompany" />
    </div>
</template>

<script setup lang="ts">
definePageMeta({
    layout: "viewer",
});

const isCompany = ref(false);
const isViewer = ref(true);

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
    console.log("Role:", role);
});
</script>
