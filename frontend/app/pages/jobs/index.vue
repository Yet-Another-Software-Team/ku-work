<template>
    <div class="flex">
        <span class="w-full">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <span>Job Board</span>
            </h1>
            <div class="my-5">
                <JobSearchComponents />
            </div>
            <section v-for="index in totalJob" :key="index">
                <JobApplicationComponent
                    :is-selected="selectedIndex === index"
                    :data="jobs[index % jobs.length] || jobs[0]!"
                    @click="selectedIndex = index"
                />
            </section>
        </span>
        <span v-if="selectedIndex" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <JobApplicationExpanded
                    :is-viewer="false"
                    :is-selected="true"
                    :data="jobs[selectedIndex % jobs.length] || jobs[0]!"
                />
            </section>
        </span>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { mockJobData, type JobApplication } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

const totalJob = 10;
const jobs: JobApplication[] = mockJobData.jobs;
const selectedIndex = ref<number | null>(null);
</script>
