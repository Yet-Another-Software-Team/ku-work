<template>
    <div class="py-6">
        <JobRequestJobPostDetailCard :request-id="String(jobId)" />
    </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: "admin" });
const route = useRoute();
const jobId = computed(() => Number(route.params.id));
</script>
