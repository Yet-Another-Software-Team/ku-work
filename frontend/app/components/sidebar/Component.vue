<template>
    <section class="min-w-64">
        <!-- Mobile Header -->
        <div class="lg:hidden fixed top-5 right-5 z-50 flex items-center justify-end">
            <USlideover
                side="right"
                :ui="{
                    content:
                        'bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A] w-64 flex flex-col h-full p-4',
                }"
            >
                <UButton
                    icon="ic:twotone-format-list-bulleted"
                    variant="ghost"
                    color="neutral"
                    size="xl"
                />
                <template #content>
                    <template v-if="loading">
                        <header class="flex items-center justify-between mb-4 border-none">
                            <img
                                src="~/assets/images/base.png"
                                alt="KU-Work Logo"
                                class="h-12 mr-4"
                            />
                            <USkeleton class="h-8 w-8" :ui="{ background: 'bg-white/10' }" />
                        </header>
                        <div class="space-y-4">
                            <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                            <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                            <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                        </div>
                        <div class="mt-auto">
                            <USkeleton class="h-12 w-full" :ui="{ background: 'bg-white/10' }" />
                        </div>
                    </template>
                    <template v-else>
                        <!-- top -->
                        <header class="flex items-center justify-between mb-4 border-none">
                            <img
                                v-if="isViewer && !isCompany && !isAdmin"
                                src="~/assets/images/viewer.png"
                                alt="KU-Work Viewer Logo"
                                class="h-12 mr-4"
                            />
                            <img
                                v-else-if="isCompany && !isAdmin"
                                src="~/assets/images/company.png"
                                alt="KU-Work Company Logo"
                                class="h-12 mr-4"
                            />
                            <img
                                v-else-if="isAdmin"
                                src="~/assets/images/admin.png"
                                alt="KU-Work Admin Logo"
                                class="h-12 mr-4"
                            />
                            <img
                                v-else
                                src="~/assets/images/base.png"
                                alt="KU-Work Logo"
                                class="h-12 mr-4"
                            />
                            <ThemeToggle />
                        </header>
                        <SidebarMenu :items="getSidebarItems(isViewer, isAdmin, isCompany)" />
                        <!-- bottom -->
                        <div class="mt-auto">
                            <!-- Register -->
                            <UButton
                                v-if="!isRegistered && isViewer && !isAdmin && !isCompany"
                                label="Register"
                                variant="ghost"
                                size="xl"
                                icon="ic:baseline-add-circle-outline"
                                :ui="{
                                    base: 'justify-start text-left text-white hover:bg-white/10',
                                }"
                                @click="navigateStudentRegister"
                            />
                            <!-- Deactivate -->
                            <UButton
                                v-if="isViewer && !isCompany && !isAdmin && !isDeactivated"
                                label="Deactivate Profile"
                                variant="ghost"
                                size="xl"
                                icon="ic:baseline-delete-forever"
                                :ui="{
                                    base: 'justify-start text-left text-white hover:bg-white/10',
                                }"
                                class="mt-2"
                                @click="openDeactivateModal = true"
                            />
                            <DeactivateModal
                                v-if="isViewer && !isCompany && !isAdmin && !isDeactivated"
                                v-model:open="openDeactivateModal"
                                class="mt-2"
                                @update:close="(value) => (openDeactivateModal = value)"
                            />
                            <!-- Reactivate -->
                            <!-- Reactivate -->
                            <UButton
                                v-if="isViewer && !isCompany && !isAdmin && isDeactivated"
                                label="Reactivate Profile"
                                variant="ghost"
                                size="xl"
                                icon="ic:baseline-refresh"
                                :ui="{
                                    base: 'justify-start text-left text-white hover:bg-white/10',
                                }"
                                class="mt-2"
                                @click="openReactivateModal = true"
                            />
                            <ReactivateModal
                                v-if="isViewer && !isCompany && !isAdmin && isDeactivated"
                                v-model:open="openReactivateModal"
                                class="mt-2"
                                @update:close="(value) => (openReactivateModal = value)"
                            />
                            <!-- Logout -->
                            <LogoutButton />
                        </div>
                    </template>
                </template>
            </USlideover>
        </div>
        <!-- Sidebar expanded (desktop) -->
        <div
            class="fixed top-0 left-0 w-64 hidden lg:flex flex-col h-full p-4 bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A]/90 shadow-md space-y-2"
        >
            <template v-if="loading">
                <header class="flex items-center justify-between mb-4">
                    <img src="~/assets/images/base.png" alt="KU-Work Logo" class="h-12 mr-4" />
                    <USkeleton class="h-8 w-8" :ui="{ background: 'bg-white/10' }" />
                </header>
                <div class="space-y-4">
                    <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                    <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                    <USkeleton class="h-10 w-full" :ui="{ background: 'bg-white/10' }" />
                </div>
                <div class="mt-auto">
                    <USkeleton class="h-12 w-full" :ui="{ background: 'bg-white/10' }" />
                </div>
            </template>
            <template v-else>
                <header class="flex items-center justify-between mb-4">
                    <img
                        v-if="isViewer && !isCompany && !isAdmin"
                        src="~/assets/images/viewer.png"
                        alt="KU-Work Viewer Logo"
                        class="h-12 mr-4"
                    />
                    <img
                        v-else-if="isCompany && !isAdmin"
                        src="~/assets/images/company.png"
                        alt="KU-Work Company Logo"
                        class="h-12 mr-4"
                    />
                    <img
                        v-else-if="isAdmin"
                        src="~/assets/images/admin.png"
                        alt="KU-Work Admin Logo"
                        class="h-12 mr-4"
                    />
                    <img
                        v-else
                        src="~/assets/images/base.png"
                        alt="KU-Work Logo"
                        class="h-12 mr-4"
                    />
                    <ThemeToggle />
                </header>
                <!-- Menu -->
                <SidebarMenu :items="getSidebarItems(isViewer, isAdmin, isCompany)" />
                <!-- Footer -->
                <div class="mt-auto">
                    <!-- Register -->
                    <UButton
                        v-if="!isRegistered && isViewer && !isAdmin && !isCompany"
                        label="Register"
                        variant="ghost"
                        size="xl"
                        icon="ic:baseline-add-circle-outline"
                        :ui="{ base: 'justify-start text-left text-white hover:bg-white/10' }"
                        @click="navigateRegister"
                    />
                    <!-- Deactivate -->
                    <UButton
                        v-if="isViewer && !isCompany && !isAdmin && !isDeactivated"
                        label="Deactivate Profile"
                        variant="ghost"
                        size="xl"
                        icon="ic:baseline-delete-forever"
                        :ui="{ base: 'justify-start text-left text-white hover:bg-white/10' }"
                        class="mt-2"
                        @click="openDeactivateModal = true"
                    />
                    <DeactivateModal
                        v-if="isViewer && !isCompany && !isAdmin && !isDeactivated"
                        v-model:open="openDeactivateModal"
                        class="mt-2"
                        @update:close="(value) => (openDeactivateModal = value)"
                    />
                    <!-- Reactivate -->
                    <UButton
                        v-if="isViewer && !isCompany && !isAdmin && isDeactivated"
                        label="Reactivate Profile"
                        variant="ghost"
                        size="xl"
                        icon="ic:baseline-refresh"
                        :ui="{ base: 'justify-start text-left text-white hover:bg-white/10' }"
                        class="mt-2"
                        @click="openReactivateModal = true"
                    />
                    <ReactivateModal
                        v-if="isViewer && !isCompany && !isAdmin && isDeactivated"
                        v-model:open="openReactivateModal"
                        class="mt-2"
                        @update:close="(value) => (openReactivateModal = value)"
                    />
                    <!-- Logout -->
                    <LogoutButton />
                </div>
            </template>
        </div>
    </section>
</template>

<script setup lang="ts">
import { useNuxtApp } from "#app";

const logout = useNuxtApp().$logout as () => void;

const username = ref<string | null>(null);

const openDeactivateModal = ref(false);
const openReactivateModal = ref(false);
const isDeactivated = ref(false);

const isViewer = ref(true);
const isCompany = ref(false);
const isAdmin = ref(false);
const isRegistered = ref(false);
const loading = ref(true);

onMounted(async () => {
    const role = localStorage.getItem("role");
    username.value = localStorage.getItem("username");
    isRegistered.value = localStorage.getItem("isRegistered") === "true";
    if (!role || !username.value) {
        // Data is wrong --> logout
        logout();
    }

    if (role === "company") {
        isViewer.value = false;
        isCompany.value = true;
        isAdmin.value = false;
    } else if (role === "student") {
        isViewer.value = false;
        isCompany.value = false;
        isAdmin.value = false;
    } else if (role === "viewer") {
        isViewer.value = true;
        isCompany.value = false;
        isAdmin.value = false;
    } else if (role === "admin") {
        isViewer.value = false;
        isCompany.value = false;
        isAdmin.value = true;
    } else {
        // Data is wrong --> logout
        logout();
    }

    // Determine if the account is deactivated by calling the /me endpoint.
    // Backend returns 403 with { message: "account deactivated" } when deactivated.
    try {
        const api = useApi();
        await api.get("/me", {
            headers: {
                Authorization: `Bearer ${localStorage.getItem("token")}`,
            },
        });
        isDeactivated.value = false;
    } catch (err) {
        // Normalize unknown error to an axios-like shape without using `any` in the catch param
        type Response = {
            response?: { status?: number; data?: { message?: string } };
            status?: number;
            message?: string;
        };
        const e = err as Response;
        const status = e?.response?.status || e?.status;
        const message = e?.response?.data?.message || e?.message || null;
        if (status === 403 && message === "account deactivated") {
            isDeactivated.value = true;
            console.log("Account is deactivated.");
        }
    }

    loading.value = false;
});

function navigateRegister() {
    navigateTo("/register/student", { replace: true });
}

function navigateStudentRegister() {
    navigateTo("/register/student", { replace: true });
}

function getSidebarItems(
    isViewer: boolean,
    isAdmin: boolean,
    isCompany: boolean
): Array<{ label: string; icon: string; to: string; disable: boolean }> {
    if (isAdmin) {
        return [
            {
                label: "dashboard",
                icon: "ic:baseline-dashboard",
                to: "/admin/dashboard",
                disable: false,
            },
            {
                label: "accounts",
                icon: "ic:baseline-people",
                to: "/admin/accounts",
                disable: false,
            },
            {
                label: "posts",
                icon: "ic:baseline-work",
                to: "/admin/posts",
                disable: false,
            },
            {
                label: "log",
                icon: "ic:baseline-history",
                to: "/admin/log",
                disable: false,
            },
        ];
    }

    const items = [];
    if (!isCompany) {
        items.push({
            label: "Job Board",
            icon: "ic:baseline-work",
            to: "/jobs",
            disable: false,
        });
    }
    if (!isViewer) {
        items.unshift({
            label: username.value || "Profile",
            icon: "ic:baseline-person",
            to: "/profile",
            disable: false,
        });
        items.push({
            label: "Dashboard",
            icon: "ic:baseline-dashboard",
            to: "/dashboard",
            disable: false,
        });
    } else {
        items.unshift({
            label: username.value || "Profile",
            icon: "ic:baseline-person",
            to: "/profile",
            disable: true,
        });
    }
    return items;
}
</script>
