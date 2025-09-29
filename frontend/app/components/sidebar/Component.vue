<template>
    <section class="min-w-64">
        <div
            v-if="loading"
            class="fixed inset-0 flex items-center justify-center bg-white z-50"
        ></div>
        <!-- Sidebar toggle with button (mobile, ipad, small screens) -->
        <USlideover
            side="left"
            :ui="{
                content:
                    'bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A] w-64 flex flex-col h-full p-4 lg:hidden',
            }"
        >
            <UButton
                icon="ic:twotone-format-list-bulleted"
                color="neutral"
                variant="subtle"
                class="m-5"
            />
            <template #content>
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
                    <LogoutButton />
                </div>
            </template>
        </USlideover>
        <!-- Sidebar expanded (desktop) -->
        <div
            class="fixed top-0 left-0 w-64 hidden lg:flex flex-col h-full p-4 bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A]/90 shadow-md space-y-2"
        >
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
                <img v-else src="~/assets/images/base.png" alt="KU-Work Logo" class="h-12 mr-4" />
                <ThemeToggle />
            </header>
            <SidebarMenu :items="getSidebarItems(isViewer, isAdmin, isCompany)" />
            <div class="mt-auto">
                <LogoutButton />
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
const isViewer = ref(true);
const isCompany = ref(false);
const isAdmin = ref(false);
const loading = ref(true);

onMounted(() => {
    const role = localStorage.getItem("role");

    if (role === "company") {
        isViewer.value = false;
        isCompany.value = true;
        isAdmin.value = false;
    }
    if (role === "student") {
        isViewer.value = false;
        isCompany.value = false;
        isAdmin.value = false;
    }
    loading.value = false;
});

function getSidebarItems(
    isViewer: boolean,
    isAdmin: boolean,
    isCompany: boolean
): Array<{ label: string; icon: string; to: string; disable: boolean }> {
    if (isAdmin) {
        return [
            {
                label: "Admin",
                icon: "ic:baseline-person",
                to: "/admin/dashboard",
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
            label: "Profile",
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
            label: "Profile",
            icon: "ic:baseline-person",
            to: "/profile",
            disable: true,
        });
    }
    return items;
}
</script>
