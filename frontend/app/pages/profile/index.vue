<template>
    <div class="relative pt-5 pb-2">
        <!-- Loading screen -->
        <div
            v-if="loading"
            class="fixed inset-0 flex items-center bg-[#F7F8F4] dark:bg-neutral-700 justify-center z-50"
        >
            <div
                class="h-10 w-10 animate-spin rounded-full border-4 border-gray-300 border-t-primary"
            ></div>
        </div>

        <!-- Content -->
        <ProfileCard v-if="userRole === 'student'" />
        <CompanyProfileCard v-if="userRole === 'company'" />
        <div v-if="userRole === 'viewer'" class="text-center items-center py-8">
            <h1 class="text-3xl font-semibold text-gray-600 dark:text-gray-400 mb-4">
                Profile Not Available
            </h1>
            <p class="text-gray-500">
                You are a viewer and don't have a profile page. Register as a student to create your
                profile.
            </p>
            <UButton v-if="!isRegistered" class="mt-4" @click="navigateToRegisterStudent">
                Register as Student
            </UButton>
        </div>
        <div v-if="!['student', 'company', 'viewer'].includes(userRole)" class="text-center py-8">
            <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">Access Denied</h2>
            <p class="text-gray-500 mt-2">You don't have permission to view this page.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
const userRole = ref<string>("viewer");
const isRegistered = ref(false);
const loading = ref(true);

onMounted(() => {
    if (import.meta.client) {
        const role = localStorage.getItem("role") || "viewer";
        const registered = localStorage.getItem("isRegistered") === "true";
        if (role === "viewer") {
            const toast = useToast();
            toast.add({
                title: "Insufficient Permissions",
                description: "Register as a student to create your profile.",
                color: "error",
            });
            navigateTo("/jobs", { replace: true });
        }
        if (!["student", "company", "viewer"].includes(role)) {
            const toast = useToast();
            toast.add({
                title: "Unrecognized Role",
                description: "Please relogin to access your profile.",
                color: "error",
            });
            navigateTo("/", { replace: true });
        }
        userRole.value = role;
        isRegistered.value = registered;
        loading.value = false;
    }
});

function navigateToRegisterStudent() {
    navigateTo("/register/student", { replace: true });
}

definePageMeta({
    layout: "viewer",
    middleware: "auth",
});
</script>
