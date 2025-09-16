<template>
    <div class="flex">
        <section class="w-full">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <span>Job Board</span>
            </h1>
            <div class="my-5">
                <JobSearchComponents />
            </div>
            <section v-for="(job, index) in jobs" :key="index">
                <JobApplicationComponent
                    :is-selected="selectedIndex === index"
                    :data="job"
                    @click="selectedIndex = index"
                />
            </section>
        </section>
        <section v-if="selectedIndex !== null" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <ExpandedJobApplication
                    :is-viewer="false"
                    :is-selected="true"
                    :data="jobs[selectedIndex % jobs.length] || jobs[0]!"
                />
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import ExpandedJobApplication from "~/components/ExpandedJobApplication.vue";
import { mockJobData, type JobApplication } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

const jobs: JobApplication[] = mockJobData.jobs;
const selectedIndex = ref<number | null>(null);
</script>
