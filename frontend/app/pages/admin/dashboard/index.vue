<template>
    <div>
        <!-- Selector -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mt-6 mb-6">Dashboard</h1>
        <section class="h-[3em] overflow-hidden border-b-1 my-5">
            <div v-if="isLoading">
                <USkeleton class="h-[6em] w-[28em] left-0 top-0" />
            </div>
            <div v-else class="flex flex-row gap-2 h-[6em] w-[28em] left-0 top-0">
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="setTailwindClasses(false)"
                    @click="selectStudent"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Students</p>
                </div>
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="setTailwindClasses(true)"
                    @click="selectCompany"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Companies</p>
                </div>
            </div>
        </section>
        <!-- Total requests and sort -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[3em] w-full left-0 top-0 mb-5" />
            </div>
            <div v-else class="flex justify-between">
                <h1 class="text-2xl font-semibold mb-2">{{ totalRequests }} {{ headerLabel }}</h1>
                <div class="flex gap-5">
                    <!-- TODO: Implement sorting -->
                    <h1 class="text-2xl font-semibold mb-2">Sort by:</h1>
                    <USelectMenu
                        v-model="selectSortOption"
                        value-key="id"
                        :items="sortOptions"
                        placement="bottom-end"
                        class="w-[10em]"
                        @change="fetchStudents"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />
        </section>
        <!-- Request list -->
        <section class="flex flex-wrap w-full h-full place-content-start gap-5">
            <!-- Student registration requests -->
            <template v-if="!isCompany">
                <template v-if="studentData.length">
                    <StudentRequestProfileCard
                        v-for="(data, index) in studentData"
                        :key="index"
                        :request-id="data.id"
                        :data="data"
                        @student-approval-status="selectStudent"
                    />
                </template>
                <template v-else>
                    <div class="w-full text-center py-10">
                        <Icon
                            name="ic:baseline-inbox"
                            class="w-16 h-16 text-gray-400 mx-auto mb-4"
                        />
                        <p class="text-gray-500 text-lg">No pending student registrations.</p>
                    </div>
                </template>
            </template>
            <!-- Company post requests -->
            <template v-else>
                <template v-if="companyRequests.length">
                    <JobPostRequestCard
                        v-for="job in companyRequests"
                        :key="job.id"
                        :request-id="String(job.id)"
                        :data="job"
                        @job-approval-status="onCompanyRequestResolved"
                    />
                </template>
                <template v-else>
                    <div class="w-full text-center py-10">
                        <Icon
                            name="ic:baseline-inbox"
                            class="w-16 h-16 text-gray-400 mx-auto mb-4"
                        />
                        <p class="text-gray-500 text-lg">No pending company posts.</p>
                    </div>
                </template>
            </template>
        </section>
    </div>
</template>

<script setup lang="ts">
import type { ProfileInformation, JobPost } from "~/data/datatypes";

definePageMeta({
    layout: "admin",
});

const isLoading = ref(true);

const totalRequests = ref(0);
const isCompany = ref(false);

const api = useApi();
const studentData = ref<ProfileInformation[]>([]);
const companyRequests = ref<JobPost[]>([]);
const loggedOnce = ref(false);

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);

const selectSortOption = ref("latest");

const fetchStudents = async () => {
    isLoading.value = true;

    try {
        await selectStudent();
    } catch (error) {
        console.error("Error fetching student data:", error);
    } finally {
        isLoading.value = false;
    }
};

onMounted(fetchStudents);

function setTailwindClasses(activeCondition: boolean) {
    if (isCompany.value == activeCondition) {
        return "bg-primary-200 flex flex-col border rounded-3xl w-1/2 text-primary-800 hover:bg-primary-300";
    } else {
        return "bg-gray-200 flex flex-col border rounded-3xl w-1/2 text-gray-500 hover:bg-gray-300";
    }
}

const headerLabel = computed(() => (isCompany.value ? "Posts" : "Applicants"));

async function selectCompany() {
    isCompany.value = true;
    totalRequests.value = 0;
    await fetchPendingCompanyPosts();
    totalRequests.value = companyRequests.value.length;
}

const studentOffset = ref(0);
const studentLimit = ref(10);

async function selectStudent() {
    isCompany.value = false;
    try {
        const response = await api.get<ProfileInformation[] | { profile: ProfileInformation }>(
            "/students",
            {
                params: {
                    limit: studentLimit.value,
                    offset: studentOffset.value,
                    approvalStatus: "pending",
                    sortBy: selectSortOption.value,
                },
                withCredentials: true,
            }
        );
        const data = response.data as ProfileInformation[] | { profile: ProfileInformation };
        if (Array.isArray(data)) {
            studentData.value = data;
            totalRequests.value = data.length;
        } else if (data && "profile" in data) {
            studentData.value = [data.profile];
            totalRequests.value = 1;
        } else {
            studentData.value = [];
            totalRequests.value = 0;
        }
    } catch (error) {
        console.error("Error fetching student data:", error);
    }
}
const postOffset = ref(0);
const postLimit = ref(10);

async function fetchPendingCompanyPosts() {
    try {
        const res = await api.get<JobPost[] | { jobs: JobPost[] }>("/jobs", {
            params: {
                limit: postLimit.value,
                offset: postOffset.value,
                approvalStatus: "pending",
                sortBy: selectSortOption.value,
            },
            withCredentials: true,
        });
        if (!loggedOnce.value) {
            console.log("[Admin Dashboard] Pending jobs response:", res.data);
            loggedOnce.value = true;
        }
        // Some backends return { jobs: [...] }, others return array directly
        const payload = res.data as JobPost[] | { jobs: JobPost[] };
        companyRequests.value = Array.isArray(payload) ? payload : (payload.jobs ?? []);
    } catch (e: unknown) {
        api.showErrorToast(api.handleError(e), "Failed to load company posts");
    }
}

function onCompanyRequestResolved(jobId: string) {
    companyRequests.value = companyRequests.value.filter((j) => String(j.id) !== jobId);
    if (isCompany.value) {
        totalRequests.value = companyRequests.value.length;
    }
}
</script>
