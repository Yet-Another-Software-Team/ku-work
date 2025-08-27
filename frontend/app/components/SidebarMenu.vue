<template>
    <div class="flex flex-col space-y-2 border-none">
        <UButton
            v-for="item in items"
            :key="item.label"
            :icon="item.icon"
            :to="item.to"
            :label="item.label"
            :disabled="item.disable"
            :ui="{ base: 'justify-start text-left text-white hover:bg-white/10' }"
            variant="ghost"
            size="xl"
            block
        />
    </div>
</template>

<script setup lang="ts">
defineProps<{
    items: Array<{ label: string; icon: string; to: string; disable: boolean }>;
}>();
</script>
