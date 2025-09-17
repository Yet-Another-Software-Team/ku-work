<template>
    <div class="flex">
        <section class="w-full">
            <!-- Title -->
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <span>Job Board</span>
            </h1>
            <!-- Search component -->
            <div class="my-5">
                <JobSearchComponents
                    :locations="jobs.map((job) => job.location)"
                    @update:search="search = $event"
                    @update:location="location = $event"
                />

                <!-- More options -->
                <div>
                    <SearchMoreButton
                        @update:salary-range="salaryRange = $event"
                        @update:job-type="jobType = $event"
                        @update:exp-type="expType = $event"
                    />
                </div>
            </div>
            <!-- Job applications -->
            <section v-for="(job, index) in filteredJobs" :key="index">
                <JobApplicationComponent
                    :is-selected="selectedIndex === index"
                    :data="job"
                    @click="selectedIndex = index"
                />
            </section>
        </section>
        <!-- Expanded application -->
        <section v-if="selectedIndex !== null" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <ExpandedJobApplication
                    v-if="filteredJobs.length > 0"
                    :is-viewer="false"
                    :is-selected="true"
                    :data="filteredJobs[selectedIndex % filteredJobs.length] || filteredJobs[0]!"
                />
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import ExpandedJobApplication from "~/components/ExpandedJobApplication.vue";
import { mockJobData, type JobApplication } from "~/data/mockData";
import type { CheckboxGroupValue } from "@nuxt/ui";

definePageMeta({
    layout: "viewer",
});

// Jobs
const jobs: JobApplication[] = mockJobData.jobs;
const selectedIndex = ref<number | null>(null);

// Search and Location
const search = ref("");
const location = ref<string | null>(null);
// More filters
const jobType = ref<CheckboxGroupValue[] | null>(null);
const expType = ref<CheckboxGroupValue[] | null>(null);
const salaryRange = ref<number[] | null>(null);

const filteredJobs = computed(() => {
    return jobs.filter((job) => {
        const matchesSearch =
            job.position.toLowerCase().includes(search.value.toLowerCase()) ||
            job.name.toLowerCase().includes(search.value.toLowerCase());

        const matchesLocation =
            !location.value || job.location.toLowerCase().includes(location.value.toLowerCase());

        const matchesSalary =
            !salaryRange.value ||
            (job.minSalary >= (salaryRange.value[0] ?? 0) &&
                job.maxSalary <= (salaryRange.value[1] ?? Infinity));

        const matchesJobType =
            !jobType.value || jobType.value.length === 0 || jobType.value.includes(job.jobType);

        const matchesExpType =
            !expType.value ||
            expType.value.length === 0 ||
            expType.value.includes(job.experienceType);

        return (
            matchesSearch && matchesLocation && matchesSalary && matchesJobType && matchesExpType
        );
    });
});
</script>
