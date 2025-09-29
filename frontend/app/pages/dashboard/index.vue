<template>
    <div>
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">Dashboards</h1>
        <div class="flex flex-wrap gap-10">
            <JobCardCompany
                v-for="job in data"
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
                    class="absolute top-0 bottom-0 left-0 right-0 text-white"
                    @click="openJobPostForm = true"
                />
                <template #content>
                    <JobPostForm @close="handleJobFormClose" />
                </template>
            </UModal>
        </div>
    </div>
</template>

<script setup lang="ts">
definePageMeta({
    layout: "viewer",
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
    try {
        const response = await api.get("/job");
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

onMounted(fetchJobs);

const updateJobOpen = (id: number, value: boolean) => {
    const job = data.value.find((job) => job.id === id);
    if (job) {
        job.open = value;
    }
};

const handleJobFormClose = () => {
    openJobPostForm.value = false;
    // Refresh jobs list after posting
    fetchJobs();
};
</script>
