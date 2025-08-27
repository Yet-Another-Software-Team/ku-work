<template>
    <section class="min-w-64">
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
                    <h1 class="text-2xl font-bold text-white mb-4">Icon here</h1>
                    <ThemeToggle />
                </header>
                <SidebarMenu :items="getSidebarItems(isViewer)" />
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
                <h1 class="text-2xl font-bold text-white mb-4">Icon here</h1>
                <ThemeToggle />
            </header>
            <SidebarMenu :items="getSidebarItems(isViewer)" />
            <div class="mt-auto">
                <LogoutButton />
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
withDefaults(
    defineProps<{
        isViewer?: boolean;
    }>(),
    {
        isViewer: false,
    }
);

function getSidebarItems(
    isViewer: boolean
): Array<{ label: string; icon: string; to: string; disable: boolean }> {
    const items = [
        {
            label: "Job Board",
            icon: "ic:baseline-work",
            to: "/jobboard",
            disable: false,
        },
    ];
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
