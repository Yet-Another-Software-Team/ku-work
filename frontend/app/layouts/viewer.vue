<template>
    <div class="block lg:flex light:bg-[#F7F8F4]">
        <SidebarComponent :is-viewer="isViewer" />
        <div class="m-10 lg:w-full h-full min-h-dvh">
            <slot />
        </div>
    </div>
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
</script>
