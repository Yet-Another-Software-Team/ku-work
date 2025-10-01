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
            <section v-for="(job, index) in filteredJobs" :key="index">
                <JobPostComponent
                    :is-selected="selectedIndex === index"
                    :data="job"
                    @click="selectedIndex = index"
                />
            </section>
        </section>
        <!-- Expanded Job Post -->
        <section v-if="selectedIndex !== null && selectedIndex < filteredJobs.length" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <JobPostExpanded
                    v-if="filteredJobs.length > 0"
                    :is-viewer="userRole === 'viewer'"
                    :is-selected="true"
                    :data="filteredJobs[selectedIndex]!"
                />
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import type { CheckboxGroupValue } from "@nuxt/ui";
import type { JobPost } from "~/data/mockData";

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

const filteredJobs = computed(() => {
    return jobs.value.filter((job) => {
        const companyName = job.company?.User?.Username || job.name;
        const matchesSearch =
            job.position.toLowerCase().includes(search.value.toLowerCase()) ||
            job.name.toLowerCase().includes(search.value.toLowerCase()) ||
            companyName.toLowerCase().includes(search.value.toLowerCase());

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

// API call to fetch jobs
const api = useApi();

interface getJobForm {
    location?: string;
    keyword?: string;
    jobtype?: string[];
    experience?: string[];
    minsalary?: number;
    maxsalary?: number;
}

const fetchJobs = async () => {
    const jobForm: getJobForm = {
        location: location.value ?? "",
        keyword: search.value ?? "",
        jobtype: jobType.value ? jobType.value.map(String) : [""],
        experience: expType.value ? expType.value.map(String) : undefined,
        minsalary: salaryRange.value ? salaryRange.value[0] : 0,
        maxsalary: salaryRange.value ? salaryRange.value[1] : 99999999,
    };
    try {
        // Only invoke fetch jobs on client-side
        if (import.meta.client) {
            const response = await api.get("/job", {
                params: { jobForm },
            });
            console.log(response);
            if (response.data.jobs && response.data.jobs.length > 0) {
                jobs.value.push(...response.data.jobs);
            }
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
