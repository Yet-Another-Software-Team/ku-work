<template>
    <div class="block lg:flex light:bg-[#F7F8F4]">
        <SidebarComponent :is-admin="isAdmin" />
        <div class="m-10 lg:w-full h-full min-h-dvh">
            <slot />
        </div>
    </div>
</template>

<script setup lang="ts">
withDefaults(
    defineProps<{
        isAdmin?: boolean;
    }>(),
    {
        isAdmin: true,
    }
);
</script>
