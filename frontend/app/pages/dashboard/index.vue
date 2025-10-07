<template>
    <div class="pt-5 pb-2">
        <!-- Company Dashboard -->
        <div v-if="userRole === 'company'">
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">
                Company Dashboard
            </h1>
            <div class="flex flex-wrap gap-10">
                <div v-if="data.length === 0">
                    <p class="text-center text-neutral-300 dark:text-neutral-400 text-xl">
                        No jobs posted yet.
                    </p>
                </div>
                <JobCardCompany
                    v-for="job in data"
                    v-else
                    :key="job.id"
                    class="h-[18em] w-full lg:w-[25em] drop-shadow-md"
                    :job-i-d="job.id.toString()"
                    :open="job.open"
                    :position="job.position"
                    :accepted="job.accepted"
                    :rejected="job.rejected"
                    :pending="job.pending"
                    @update:open="(value: boolean) => updateJobOpen(job.id, value)"
                />
            </div>
            <div class="bg-primary p-2 rounded-full size-[4em] fixed bottom-5 right-[6vw]">
                <UModal v-model:open="openJobPostForm">
                    <Icon
                        name="ic:baseline-plus"
                        size="4em"
                        mode="svg"
                        class="absolute top-0 bottom-0 left-0 right-0 text-white cursor-pointer"
                        @click="openJobPostForm = true"
                    />
                    <template #content>
                        <JobPostForm @close="handleJobFormClose" />
                    </template>
                </UModal>
            </div>
        </div>

        <!-- Student Dashboard -->
        <div v-else-if="userRole === 'student'">
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">
                Student Dashboard
            </h1>
            <div class="text-center py-8">
                <p class="text-gray-500 text-lg">
                    Student dashboard functionality not yet implemented.
                </p>
                <p class="text-gray-400 text-sm mt-2">
                    This feature will be available in a future update.
                </p>
            </div>
        </div>

        <!-- Access Denied -->
        <div v-else class="text-center py-8">
            <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">Access Denied</h2>
            <p class="text-gray-500 mt-2">You don't have permission to view this dashboard.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
const userRole = ref<string>("viewer");

definePageMeta({
    layout: "viewer",
    middleware: "auth",
});

const openJobPostForm = ref(false);

type Job = {
    id: number;
    position: string;
    accepted: number;
    rejected: number;
    pending: number;
    open: boolean;
};

const data = ref<Job[]>([]);

const api = useApi();
const { add: addToast } = useToast();

const fetchJobs = async () => {
    // Only fetch jobs for companies
    if (userRole.value !== "company") return;

    try {
        const response = await api.get("/jobs", {
            params: {
                companyId: "self",
            },
        });
        console.log("Fetched jobs:", response.data);
        data.value = response.data.jobs || [];
    } catch (error) {
        const apiError = error as { message?: string };
        console.error("Failed to fetch jobs:", apiError.message || "Unknown error");
        addToast({
            title: "Error",
            description: "Failed to fetch jobs. Please refresh the page.",
            color: "error",
        });
    }
};

onMounted(() => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
    fetchJobs();
});

const updateJobOpen = (id: number, value: boolean) => {
    const job = data.value.find((job) => job.id === id);
    if (!job) return;
    const oldValue = job.open;
    job.open = value;
    api.patch(
        `/jobs/${job.id}`,
        {
            open: value,
        },
        {
            withCredentials: true,
        }
    )
        .then((x) => {
            if (x.status !== 200) Promise.reject(x);
        })
        .catch((_) => (job.open = oldValue));
};

const handleJobFormClose = () => {
    openJobPostForm.value = false;
    // Refresh jobs list after posting (only for companies)
    if (userRole.value === "company") {
        fetchJobs();
    }
};
</script>
