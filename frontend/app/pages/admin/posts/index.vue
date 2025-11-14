<template>
    <section class="w-full overflow-x-hidden">
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mt-6 mb-6">Posts</h1>

        <!-- Count and sorting -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[3em] w-full left-0 top-0 mb-5" />
            </div>
            <div v-else class="flex justify-between items-center">
                <h2 class="text-2xl font-semibold mb-2">{{ jobs.length }} Posts</h2>
                <div class="flex items-center gap-3">
                    <span class="text-2xl font-semibold mb-2">Sort by:</span>
                    <USelectMenu
                        v-model="selectSortOption"
                        value-key="id"
                        :items="sortOptions"
                        placement="bottom-end"
                        class="w-[10em]"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />
        </section>

        <div v-if="isLoading">
            <USkeleton v-for="n in 8" :key="n" class="h-[8em] w-full mb-4" />
        </div>
        <div v-else>
            <div v-if="errorMessage" class="text-center text-red-500 py-10">
                <p>{{ errorMessage }}</p>
                <UButton
                    class="mt-4"
                    color="primary"
                    variant="soft"
                    label="Retry"
                    @click="loadJobs"
                />
            </div>
            <div v-else-if="jobs.length === 0" class="text-center text-neutral-500 py-10">
                <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                <p>No posts found.</p>
            </div>
            <div v-else class="flex flex-col gap-3">
                <div v-for="job in sortedJobs" :key="job.id" class="r-card">
                    <div class="flex items-center justify-between">
                        <div class="min-w-0">
                            <p class="font-semibold truncate">{{ job.name }}</p>
                            <p class="text-sm text-gray-500 truncate">
                                {{ job.companyName }} | {{ job.position }}
                            </p>
                            <p class="text-xs text-gray-500 truncate">
                                Status: {{ job.approvalStatus }}
                            </p>
                        </div>
                        <div class="text-right text-sm text-gray-500">
                            <div>Updated: {{ new Date(job.updatedAt).toLocaleString() }}</div>
                            <UButton
                                variant="ghost"
                                color="primary"
                                label="Open"
                                @click="navigateTo(`/admin/dashboard/jobs/${job.id}`)"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
definePageMeta({ layout: "admin" });

type JobItem = {
    id: number;
    name: string;
    companyName: string;
    position: string;
    approvalStatus: string;
    updatedAt: string;
};

const api = useApi();
const isLoading = ref(true);
const errorMessage = ref("");
const jobs = ref<JobItem[]>([]);

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);
const selectSortOption = ref("latest");

const sortedJobs = computed(() => {
    const list = [...jobs.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) => a.name.localeCompare(b.name));
        case "name_za":
            return list.sort((a, b) => b.name.localeCompare(a.name));
        case "latest":
        default:
            return list.sort(
                (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
            );
    }
});

function isJobItem(job: unknown): job is JobItem {
    if (!job || typeof job !== "object") {
        return false;
    }
    const record = job as Record<string, unknown>;
    return (
        typeof record.id === "number" &&
        typeof record.name === "string" &&
        typeof record.companyName === "string" &&
        typeof record.position === "string" &&
        typeof record.approvalStatus === "string" &&
        typeof record.updatedAt === "string"
    );
}

function parseJobPayload(data: unknown): JobItem[] {
    if (Array.isArray(data)) {
        return data.filter(isJobItem);
    }
    if (data && typeof data === "object" && Array.isArray((data as { jobs?: unknown[] }).jobs)) {
        return ((data as { jobs?: unknown[] }).jobs ?? []).filter(isJobItem);
    }
    return [];
}

async function loadJobs() {
    errorMessage.value = "";
    try {
        const res = await api.get("/jobs", { withCredentials: true });
        jobs.value = parseJobPayload(res.data).map((job) => ({
            id: job.id,
            name: job.name,
            companyName: job.companyName,
            position: job.position,
            approvalStatus: job.approvalStatus,
            updatedAt: job.updatedAt,
        }));
    } catch (error) {
        console.error("Failed to load jobs", error);
        jobs.value = [];
        errorMessage.value = "Unable to load posts. Please try again.";
    }
}

onMounted(async () => {
    isLoading.value = true;
    try {
        await loadJobs();
    } finally {
        isLoading.value = false;
    }
});
</script>

<style scoped>
.r-card {
    box-shadow:
        0 4px 6px -1px rgba(0, 0, 0, 0.2),
        0 2px 4px -2px rgba(0, 0, 0, 0.2);
    text-align: left;
    padding: 1rem;
    border-radius: 0.5rem;
    background-color: #fdfdfd;
}
.dark .r-card {
    background-color: #1f2937;
}
</style>
