<template>
    <section class="w-full">
        <!-- Back -->
        <NuxtLink to="/dashboard">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <Icon name="iconoir:nav-arrow-left" class="items-center" />
                <span>Back</span>
            </h1>
        </NuxtLink>
        <!-- Job Post detail -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[20em] w-full mb-5" />
                <USkeleton class="h-[10em] w-full mb-5" />
                <div class="flex"></div>
            </div>
            <CompanyPostComponent
                v-else-if="job"
                :data="job"
                :open="job.open ?? false"
                @update:open="(value: boolean) => job && updateJobOpen(Number(job.id), value)"
                @close="fetchJob"
            />
            <div v-else class="text-center text-neutral-500 dark:text-neutral-400">
                No job applications available.
            </div>
        </section>
        <!-- Navigation bar -->
        <section v-if="job" class="h-[3em] overflow-hidden border-b-1 my-5">
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
        <!-- Job Applications List -->
        <section v-if="job">
            <div class="flex justify-between">
                <h1 class="text-2xl font-semibold mb-2">{{ countedApplication }} Applicants</h1>
                <div class="flex gap-5">
                    <h1 class="text-2xl font-semibold mb-2">Sort by:</h1>
                    <!-- TODO: Implement sorting later -->
                    <USelectMenu
                        v-model="selectSortOption"
                        value-key="id"
                        placement="bottom-end"
                        class="w-[10em]"
                        :items="sortOptions"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />
            <div v-if="isLoading">
                <USkeleton v-for="n in 10" :key="n" class="h-[20em] w-full mb-5" />
            </div>
            <div v-else class="flex flex-wrap gap-5 mb-10">
                <JobApplicationComponent
                    v-for="app in filteredApplications()"
                    :key="app.id"
                    :application-data="app"
                    @approve="acceptApplication(app, true)"
                    @reject="acceptApplication(app, false)"
                />
            </div>
        </section>
    </section>
</template>

<script setup lang="ts">
import { mockJobApplicationData } from "~/data/mockData";
import type { JobPost, JobApplication } from "~/data/mockData";

definePageMeta({
    layout: "viewer",
});

// test: set to true to use mock data
const test = ref(false);

// Jobs and applications
const job = ref<JobPost>();
const applications = ref<Array<JobApplication>>();
const countedApplication = computed(() => applicationCount());

// API call to fetch jobs
const api = useApi();
const route = useRoute();
const isLoading = ref(true);
const limit = 32;
let currentJobOffset = 0;

interface getApplicationForm {
    limit?: number;
    offset?: number;
    jobId?: number;
    sortBy?: string;
}

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);

const selectSortOption = ref("latest");

const loadContents = async () => {
    isLoading.value = true;
    try {
        const jobId = route.params.id ? Number(route.params.id) : -1;
        if (jobId === -1) {
            return;
        }
        const token = localStorage.getItem("token");
        await fetchJob(token, jobId);
        await fetchApplication(token, jobId);
        if (test.value) {
            applications.value = [mockJobApplicationData];
            for (let index = 0; index < 10; index++) {
                applications.value.push({ ...mockJobApplicationData, id: index + 2 });
            }
            if (job.value) job.value.pending = applications.value.length;
        }
    } catch (error) {
        console.error("Error during onMounted:", error);
    } finally {
        isLoading.value = false;
    }
};

onMounted(loadContents);

const acceptApplication = async (jobApplication: JobApplication, accept: boolean) => {
    try {
        await api.patch(
            `/jobs/${jobApplication.jobId}/applications/${jobApplication.userId}/status`,
            {
                status: accept ? "accepted" : "rejected",
            },
            {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("token")}`,
                },
            }
        );
        currentJobOffset = 0;
        applications.value = undefined;
        await loadContents();
    } catch (error) {
        console.error("Error accepting/rejecting job application:", error);
    }
};

const fetchJob = async (token: string | null, jobId: number) => {
    try {
        const response = await api.get(
            `/jobs/${jobId ?? (route.params.id ? Number(route.params.id) : -1)}`,
            {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            }
        );
        job.value = response.data;
    } catch (error) {
        console.error("Error fetching job details:", error);
    }
};

const fetchApplication = async (token: string | null, jobId: number) => {
    const jobForm: getApplicationForm = {
        limit: limit,
        offset: currentJobOffset,
        jobId: jobId,
        sortBy: selectSortOption.value,
    };
    try {
        const response = await api.get(`/jobs/${jobId}/applications`, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
            params: jobForm,
        });
        if (applications.value === undefined) {
            applications.value = response.data;
        } else {
            applications.value.push(...response.data);
        }
        currentJobOffset += limit;
    } catch (error) {
        console.error("Error fetching job application:", error);
    }
};

// Toggle between inprogress, accepted, rejected
const isSelected = ref("inprogress");

// Handlers for selecting application status
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

function filteredApplications() {
    if (isSelected.value === "inprogress") {
        return applications.value?.filter((app) => app.status === "pending");
    } else if (isSelected.value === "accepted") {
        return applications.value?.filter((app) => app.status === "accepted");
    } else if (isSelected.value === "rejected") {
        return applications.value?.filter((app) => app.status === "rejected");
    }
}

function applicationCount() {
    if (isSelected.value === "inprogress") {
        return job.value?.pending;
    } else if (isSelected.value === "accepted") {
        return job.value?.accepted;
    } else if (isSelected.value === "rejected") {
        return job.value?.rejected;
    }
    return 0;
}

const updateJobOpen = (id: number, value: boolean) => {
    if (job.value && Number(job.value.id) === id) {
        job.value.open = value;
    }
};
</script>
