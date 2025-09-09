<template>
    <div v-if="isSelected">
        <p>
            {{ JSON.stringify(data, null, 2) }}
        </p>
    </div>
</template>

<script setup lang="ts">
import type { JobApplication } from "~/data/mockData";

defineProps<{
    data: JobApplication;
    isSelected: boolean;
}>();
</script>
