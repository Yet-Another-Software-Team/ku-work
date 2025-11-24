<template>
    <footer class="w-full bg-white dark:bg-neutral-800 pt-6 pb-2 mt-12" aria-label="Site footer">
        <div class="px-4">
            <div class="flex justify-between gap-6">
                <!-- Brand / Mission -->
                <div class="space-y-1 text-center md:text-left">
                    <p class="text-sm font-semibold text-gray-700 dark:text-gray-200">KU-Work</p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                        Kasetsart University's Talent-Company Bridge
                    </p>
                </div>

                <!-- Navigation -->
                <nav
                    class="flex flex-wrap justify-center md:justify-end gap-x-6 gap-y-2 text-sm"
                    aria-label="Footer navigation"
                >
                    <NuxtLink
                        to="/agreement/privacy_policy"
                        class="text-gray-600 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-400 transition-colors"
                    >
                        Privacy Policy
                    </NuxtLink>
                    <NuxtLink
                        to="/agreement/terms_of_service"
                        class="text-gray-600 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-400 transition-colors"
                    >
                        Terms of Service
                    </NuxtLink>
                    <NuxtLink
                        :to="getHomePath()"
                        class="text-gray-600 hover:text-primary-600 dark:text-gray-300 dark:hover:text-primary-400 transition-colors"
                    >
                        Home
                    </NuxtLink>
                </nav>
            </div>

            <!-- Meta / Legal -->
            <div class="mt-8 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <p class="text-xs text-gray-500 dark:text-gray-400 text-center md:text-left">
                    &copy; {{ year }} KU-Work. All rights reserved.
                </p>
                <p class="text-xs text-gray-400 dark:text-gray-500 text-center md:text-right">
                    For inquiries:
                    <a
                        href="mailto:contact@ku-work.local"
                        class="underline hover:text-primary-600 dark:hover:text-primary-400"
                        >contact@ku-work.local</a
                    >
                </p>
            </div>
        </div>
    </footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear();
const authStore = useAuthStore();

const getHomePath = () => {
    // Get Home path according the user role.

    if (!authStore.isAuthenticated) {
        return "/";
    }

    if (authStore.isAdmin) {
        return "/admin/dashboard";
    }

    if (authStore.isStudent || authStore.isViewer) {
        return "/jobs";
    }

    if (authStore.isCompany) {
        return "/dashboard";
    }

    return "/";
};
</script>

<style scoped>
footer {
    /* Ensures footer does not shrink oddly in flex layouts */
    flex-shrink: 0;
}
</style>
