<template>
    <div class="flex">
        <section class="w-full">
            <!-- Title -->
            <section class="sticky top-0 z-20 bg-[#F7F8F4] dark:bg-neutral-900 pt-5 pb-2">
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
            </section>
            <!-- Job Post -->
            <section ref="jobListElement">
                <!-- Initial Loading -->
                <div v-if="isInitialLoad" class="flex justify-center py-12">
                    <div class="flex items-center gap-3 text-gray-600 dark:text-gray-400">
                        <Icon name="svg-spinners:ring-resize" class="w-8 h-8" />
                        <span class="text-lg">Loading jobs...</span>
                    </div>
                </div>

                <!-- Job Posts -->
                <div v-for="(job, index) in jobs" :key="job.id">
                    <JobPostComponent
                        :is-selected="selectedIndex === index"
                        :data="job"
                        @click="setSelectedIndex(index)"
                    />
                </div>

                <!-- Loading Indicator -->
                <div v-if="isLoadingMore" class="flex justify-center py-8">
                    <div class="flex items-center gap-3 text-gray-600 dark:text-gray-400">
                        <Icon name="svg-spinners:ring-resize" class="w-6 h-6" />
                        <span>Loading more jobs...</span>
                    </div>
                </div>

                <!-- End of Results -->
                <div v-else-if="endOfFile && jobs.length > 0" class="flex justify-center py-8">
                    <div class="text-gray-500 dark:text-gray-400">
                        <Icon name="ic:baseline-check-circle" class="w-6 h-6 inline-block mr-2" />
                        You've reached the end of the job listings
                    </div>
                </div>

                <!-- No Results -->
                <div
                    v-else-if="!isLoadingMore && jobs.length === 0 && !isInitialLoad"
                    class="flex flex-col items-center justify-center py-12"
                >
                    <Icon name="ic:baseline-work-off" class="w-16 h-16 text-gray-400 mb-4" />
                    <p class="text-gray-600 dark:text-gray-400 text-lg font-medium">
                        No jobs found
                    </p>
                    <p class="text-gray-500 dark:text-gray-500 text-sm mt-2">
                        Try adjusting your search filters
                    </p>
                </div>
            </section>
        </section>
        <!-- Expanded Job Post -->
        <section v-if="selectedIndex !== null && selectedIndex < jobs.length" class="flex">
            <!-- Normal Expand job (bigger than tablet) -->
            <USeparator
                orientation="vertical"
                class="w-fit mx-5 hidden tablet:block"
                color="neutral"
                size="lg"
            />
            <section aria-label="job expand normal">
                <JobPostExpanded
                    v-if="jobs.length > 0"
                    class="hidden tablet:block"
                    :is-viewer="userRole === 'viewer'"
                    :is-selected="true"
                    :data="jobs[selectedIndex]!"
                />
                <JobPostDrawer
                    v-if="jobs.length > 0"
                    class="block tablet:hidden"
                    :is-viewer="userRole === 'viewer'"
                    :is-selected="true"
                    :data="jobs[selectedIndex]!"
                    @close="setSelectedIndex(null)"
                />
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import type { JobPost } from "~/data/mockData";
import type { CheckboxGroupValue } from "@nuxt/ui";
import { useInfiniteScroll, watchDebounced } from "@vueuse/core";

definePageMeta({
    layout: "viewer",
});

// Jobs
const jobs = ref<JobPost[]>([]);
const selectedIndex = ref<number | null>(null);
const userRole = ref<string>("viewer");
const isLoadingMore = ref(false);
const isInitialLoad = ref(true);

// Search and Location
const search = ref("");
const location = ref<string | null>(null);
// More filters
const jobType = ref<CheckboxGroupValue[] | null>(null);
const expType = ref<CheckboxGroupValue[] | null>(null);
const salaryRange = ref<number[] | null>(null);

const jobListElement = useTemplateRef<HTMLElement>("jobListElement");
useInfiniteScroll(
    jobListElement,
    () => {
        if (!isLoadingMore.value && !endOfFile) {
            fetchJobs(currentJobOffset).then((_) => {});
        }
    },
    {
        distance: 200,
        canLoadMore: (_) => !endOfFile && !isLoadingMore.value,
    }
);

watchDebounced(
    [search, location, jobType, expType, salaryRange],
    () => {
        endOfFile = false;
        currentJobOffset = 0;
        isInitialLoad.value = true;
        fetchJobs().then((_) => {});
    },
    {
        debounce: 300,
        maxWait: 2000,
    }
);

// API call to fetch jobs
const api = useApi();
let currentJobOffset = 0;
let endOfFile = false;
const jobsLimitPerFetch = 30;

const fetchJobs = async (offset?: number) => {
    // Only invoke fetch jobs on client-side
    if (!import.meta.client) return;

    // Prevent multiple simultaneous requests
    if (isLoadingMore.value) return;

    // Set loading state
    if (offset !== undefined && offset > 0) {
        isLoadingMore.value = true;
    }

    const jobForm = new URLSearchParams();
    jobForm.append("limit", jobsLimitPerFetch as unknown as string);
    jobForm.append("offset", (offset ?? 0) as unknown as string);
    // Optionally add parameters, prevent cluttering the request parameters with default values
    if (location.value) jobForm.append("location", location.value as unknown as string);
    if (search.value) jobForm.append("keyword", search.value as unknown as string);
    if (jobType.value) jobType.value.map(String).map((x) => jobForm.append("jobType", x));
    if (expType.value) expType.value.map(String).map((x) => jobForm.append("experience", x));
    if (salaryRange.value) {
        jobForm.append("minSalary", salaryRange.value[0] as unknown as string);
        jobForm.append("maxSalary", salaryRange.value[1] as unknown as string);
    }

    try {
        const response = await api.get("/jobs", {
            params: jobForm,
        });

        const fetchedJobs = response.data.jobs || [];
        const totalCount = response.data.total || 0;

        console.log("Fetched jobs response:", {
            fetchedCount: fetchedJobs.length,
            totalCount: totalCount,
            currentOffset: offset ?? 0,
            isAppending: offset !== undefined && offset > 0,
        });

        if (offset !== undefined && offset > 0) {
            // Append new jobs for infinite scroll
            jobs.value.push(...fetchedJobs);
        } else {
            // Replace jobs for initial load or filter change
            jobs.value = fetchedJobs;
            currentJobOffset = jobsLimitPerFetch; // Reset offset for fresh start
        }

        // Update offset for next fetch (only when appending)
        if (offset !== undefined && offset > 0) {
            currentJobOffset = offset + jobsLimitPerFetch;
        }

        // Check if we've reached the end
        if (fetchedJobs.length < jobsLimitPerFetch || jobs.value.length >= totalCount) {
            endOfFile = true;
            console.log("Reached end of job listings");
        } else {
            endOfFile = false;
        }
    } catch (error) {
        console.error("Error fetching jobs:", error);
        endOfFile = true; // Stop trying to load more on error
    } finally {
        isLoadingMore.value = false;
        isInitialLoad.value = false;
    }
};

function setSelectedIndex(index: number | null) {
    if (selectedIndex.value === index) {
        // Deselect if the same index is clicked
        selectedIndex.value = null;
    } else {
        selectedIndex.value = index;
    }
}

// Get user role from localStorage
onMounted(() => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
});

await fetchJobs();
</script>
