<template>
    <div>
        <h1
            class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
        >
            <span>Job Board</span>
        </h1>
        <section v-for="index in 10" :key="index">
            <JobApplicationComponent
                :data="(jobs[index % 3] || jobs[0]) as JobApplication"
                :is-selected="selectedIndex === index"
                @click="selectedIndex = index"
            />
        </section>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { mockJobData, type JobApplication } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

const jobs: JobApplication[] = mockJobData.jobs;
const selectedIndex = ref<number | null>(null);
</script>
