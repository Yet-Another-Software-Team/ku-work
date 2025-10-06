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
            <section v-for="(job, index) in jobs" :key="index">
                <JobPostComponent
                    :is-selected="selectedIndex === index"
                    :data="job"
                    @click="selectedIndex = index"
                />
            </section>
        </section>
        <!-- Expanded Job Post -->
        <section v-if="selectedIndex !== null && selectedIndex < jobs.length" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <JobPostExpanded
                    v-if="jobs.length > 0"
                    :is-viewer="userRole === 'viewer'"
                    :is-selected="true"
                    :data="jobs[selectedIndex]!"
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
        fetchJobs(currentJobOffset).then((_) => {});
    },
    {
        distance: 10,
        canLoadMore: (_) => !endOfFile,
    }
);

// const filteredJobs = computed(() => {
//     return jobs.value.filter((job) => {
//         const companyName = job.companyName || job.name;
//         const matchesSearch =
//             job.position.toLowerCase().includes(search.value.toLowerCase()) ||
//             job.name.toLowerCase().includes(search.value.toLowerCase()) ||
//             companyName.toLowerCase().includes(search.value.toLowerCase());

//         const matchesLocation =
//             !location.value || job.location.toLowerCase().includes(location.value.toLowerCase());

//         const matchesSalary =
//             !salaryRange.value ||
//             (job.minSalary >= (salaryRange.value[0] ?? 0) &&
//                 job.maxSalary <= (salaryRange.value[1] ?? Infinity));

//         const matchesJobType =
//             !jobType.value || jobType.value.length === 0 || jobType.value.includes(job.jobType);

//         const matchesExpType =
//             !expType.value ||
//             expType.value.length === 0 ||
//             expType.value.includes(job.experienceType);

//         return (
//             matchesSearch && matchesLocation && matchesSalary && matchesJobType && matchesExpType
//         );
//     });
// });

// const adebuf = refDebounced(search);

watchDebounced(
    [search, location, jobType, expType, salaryRange],
    () =>
        fetchJobs().then((_) => {
            endOfFile = false;
            currentJobOffset = 0;
        }),
    {
        debounce: 300,
        maxWait: 2000,
    }
);

// API call to fetch jobs
const api = useApi();
let currentJobOffset = 0;
let endOfFile = false;
const jobsLimitPerFetch = 10;

const fetchJobs = async (offset?: number) => {
    // Only invoke fetch jobs on client-side
    if (!import.meta.client) return;
    const jobForm = new URLSearchParams();
    jobForm.append("limit", jobsLimitPerFetch as unknown as string);
    jobForm.append("offset", (offset ?? 0) as unknown as string);
    // Optionally add parameters, prevent cluttering the request parameters with default values
    if (location.value) jobForm.append("location", location.value as unknown as string);
    if (search.value) jobForm.append("keyword", search.value as unknown as string);
    if (jobType.value) jobType.value.map(String).map((x) => jobForm.append("jobtype", x));
    if (expType.value) expType.value.map(String).map((x) => jobForm.append("experience", x));
    if (salaryRange.value) {
        jobForm.append("minsalary", salaryRange.value[0] as unknown as string);
        jobForm.append("maxsalary", salaryRange.value[1] as unknown as string);
    }
    try {
        const response = await api.get("/job", {
            params: jobForm,
        });
        currentJobOffset += jobsLimitPerFetch;
        if (response.data.jobs.length < jobsLimitPerFetch) endOfFile = true;
        if (offset !== undefined) {
            jobs.value.push(...response.data.jobs);
        } else {
            jobs.value = response.data.jobs;
        }
    } catch (error) {
        console.error("Error fetching jobs:", error);
    }
};

// Get user role from localStorage
onMounted(() => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
});

await fetchJobs();
</script>
