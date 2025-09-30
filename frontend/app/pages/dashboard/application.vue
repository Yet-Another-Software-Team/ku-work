<template>
    <div class="flex">
        <section class="w-full">
            <!-- Title -->
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <span>Job Dashboard</span>
            </h1>
            <!-- Job applications -->
            <section>
                <CompanyApplicationComponent
                    v-if="job"
                    :data="job"
                    :open="job.open ?? false"
                    @update:open="(value: boolean) => job && updateJobOpen(Number(job.id), value)"
                />
                <div v-else class="text-center text-neutral-500 dark:text-neutral-400">
                    No job applications available.
                </div>
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { mockJobData, type JobApplication } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

// Jobs
const job: JobApplication | null = mockJobData.jobs[0] ? mockJobData.jobs[0] : null;

// API call to fetch jobs
// const api = useApi();

// interface getIndividualJobApplicationForm {}

// const fetchJobs = async () => {
//     const jobForm: getIndividualJobApplicationForm = {};
//     try {
//         const response = await api.get("/job", {
//             params: { jobForm },
//         });
//         console.log("Jobs fetched:", response.data);
//         currentJobOffset += jobsLimitPerFetch;
//         jobs.push(...response.data.jobs);
//     } catch (error) {
//         console.error("Error fetching jobs:", error);
//     }
// };
// await fetchJobs();

const updateJobOpen = (id: number, value: boolean) => {
    if (job && Number(job.id) === id) {
        job.open = value;
    }
};
</script>
