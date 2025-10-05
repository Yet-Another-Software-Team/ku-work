<template>
    <section class="w-full">
        <!-- Back -->
        <a href="/dashboard">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <Icon name="iconoir:nav-arrow-left" class="items-center" />
                <span>Back</span>
            </h1>
        </a>
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
            <section class="h-[3em] overflow-hidden border-b-1 my-5">
                <div class="flex flex-row gap-2 h-[6em] max-w-[40em] left-0 top-0">
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="setTailwindClasses('inprogress')"
                        @click="selectInProgress"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">In Progress</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="setTailwindClasses('accepted')"
                        @click="selectAccepted"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Accepted</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="setTailwindClasses('rejected')"
                        @click="selectRejected"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Rejected</p>
                    </div>
                </div>
            </section>
        </section>
    </section>
</template>

<script setup lang="ts">
import { mockJobData, type JobPost } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

// Jobs
const job: JobPost | null = mockJobData.jobs[0] ? mockJobData.jobs[0] : null;

// API call to fetch jobs
const api = useApi();
const limit = 10;
let currentJobOffset = 0;

interface getApplicationForm {
    limit?: number;
    offset?: number;
    jobId?: number;
}

onMounted(() => {
    const token = localStorage.getItem("token");
    fetchApplication(token);
});

const fetchApplication = async (token: string | null) => {
    const jobForm: getApplicationForm = {
        limit: limit,
        offset: currentJobOffset,
        jobId: job ? Number(job.id) : undefined,
    };
    try {
        const response = await api.get("/job/application", {
            headers: {
                Authorization: `Bearer ${token}`,
            },
            params: { jobForm },
        });
        console.log("Job application fetched:", response.data);
        currentJobOffset += limit;
    } catch (error) {
        console.error("Error fetching job application:", error);
    }
};

// Toggle between inprogress, accepted, rejected
const isSelected = ref("inprogress");

function setTailwindClasses(activeCondition: string) {
    const condition = isSelected.value;
    if (condition == activeCondition) {
        if (condition == "inprogress") {
            return "bg-yellow-200 flex flex-col border-1 rounded-3xl w-1/3 text-yellow-800 hover:bg-yellow-300";
        } else if (condition == "accepted") {
            return "bg-green-200 flex flex-col border-1 rounded-3xl w-1/3 text-primary-800 hover:bg-primary-300";
        }
        return "bg-error-200 flex flex-col border-1 rounded-3xl w-1/3 text-error-800 hover:bg-error-300";
    } else {
        return "bg-gray-200 flex flex-col border-1 rounded-3xl w-1/3 text-gray-500 hover:bg-gray-300";
    }
}

function selectInProgress() {
    isSelected.value = "inprogress";
}

function selectAccepted() {
    isSelected.value = "accepted";
}

function selectRejected() {
    isSelected.value = "rejected";
}

const updateJobOpen = (id: number, value: boolean) => {
    if (job && Number(job.id) === id) {
        job.open = value;
    }
};
</script>
