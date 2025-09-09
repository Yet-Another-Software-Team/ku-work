<template>
    <div v-if="isSelected" class="w-[20em]">
        <UButton
            label="Close"
            color="neutral"
            variant="outline"
            class="mb-5"
            @click="$emit('close')"
        />
        <p>
            {{ JSON.stringify(data, null, 2) }}
        </p>
    </div>
</template>

<script setup lang="ts">
import { UButton } from "#components";
import type { JobApplication } from "~/data/mockData";

defineProps<{
    data: JobApplication;
    isSelected: boolean;
}>();

defineEmits<{
    (e: "close"): void;
}>();
</script>
