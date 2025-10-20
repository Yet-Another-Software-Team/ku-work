<template>
    <div class="py-6">
        <RequestedJobPost :request-id="String(jobId)" />
    </div>
</template>

<script setup lang="ts">
import RequestedJobPost from "~/components/job/request/RequestedJobPost.vue";
definePageMeta({ layout: "admin" });
const route = useRoute();
const jobId = computed(() => Number(route.params.id));
</script>
