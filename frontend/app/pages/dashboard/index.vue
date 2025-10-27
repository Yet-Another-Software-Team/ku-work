<template>
    <div class="pt-5 pb-2">
        <div v-if="isLoading">
            <USkeleton class="h-12 w-1/3 mb-5" />
            <div class="flex flex-wrap gap-10">
                <USkeleton
                    v-for="n in 10"
                    :key="n"
                    class="h-[18em] w-full lg:w-[25em] drop-shadow-md"
                />
            </div>
        </div>
        <!-- Company Dashboard -->
        <div v-else-if="userRole === 'company' && !isLoading">
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">
                Company Dashboard
            </h1>

            <!-- Tab Navigation -->
            <div class="relative mb-8">
                <div class="flex items-end gap-0">
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0"
                        :class="
                            companyActiveTab === 'accepted'
                                ? 'bg-primary-500 text-white border-primary-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="companyActiveTab = 'accepted'"
                    >
                        Accepted
                    </button>
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0 -ml-px"
                        :class="
                            companyActiveTab === 'pending'
                                ? 'bg-warning-500 text-white border-warning-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="companyActiveTab = 'pending'"
                    >
                        In Progress
                    </button>
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0 -ml-px"
                        :class="
                            companyActiveTab === 'rejected'
                                ? 'bg-error-500 text-white border-error-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="companyActiveTab = 'rejected'"
                    >
                        Rejected
                    </button>
                </div>
                <div class="border-b-2 border-gray-300 dark:border-gray-600 -mt-0.5"></div>
            </div>

            <!-- Job Count -->
            <div class="flex items-center justify-between mb-5">
                <p class="text-lg font-semibold text-gray-700 dark:text-gray-300">
                    {{ totalJobs }} Jobs
                </p>
            </div>

            <div class="flex flex-wrap gap-10">
                <div v-if="filteredJobs.length === 0" class="w-full text-center py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p class="text-gray-500 text-lg">No jobs found.</p>
                    <p class="text-gray-400 text-sm mt-2">
                        {{
                            companyActiveTab === "accepted"
                                ? "You don't have any accepted jobs yet."
                                : companyActiveTab === "pending"
                                  ? "You don't have any pending jobs."
                                  : "You don't have any rejected jobs."
                        }}
                    </p>
                </div>
                <JobCardCompany
                    v-for="job in filteredJobs"
                    v-else
                    :key="job.id"
                    :data="job"
                    class="h-[18em] w-full lg:w-[25em] drop-shadow-md"
                    :job-i-d="job.id.toString()"
                    :open="job.open"
                    :position="job.position"
                    :accepted="job.accepted"
                    :rejected="job.rejected"
                    :pending="job.pending"
                    :approval-status="job.approvalStatus"
                    @update:open="(value: boolean) => updateJobOpen(job.id, value)"
                    @close="fetchJobs"
                />
            </div>

            <!-- Company Pagination -->
            <div v-if="totalJobPages > 1" class="flex justify-center mt-8">
                <UPagination
                    v-model:page="currentJobPage"
                    :items-per-page="itemsPerPage"
                    :total="totalJobs"
                    show-edges
                    :sibling-count="1"
                />
            </div>

            <div
                class="bg-primary-500 p-2 rounded-full size-[4em] fixed bottom-5 right-[6vw] hover:bg-primary-700 transition-all duration-200"
            >
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
        <div v-else-if="userRole === 'student' && !isLoading">
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">
                Student Dashboard
            </h1>

            <!-- Tab Navigation -->
            <div class="relative mb-8">
                <div class="flex items-end gap-0">
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0"
                        :class="
                            activeTab === 'accepted'
                                ? 'bg-primary-500 text-white border-primary-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="activeTab = 'accepted'"
                    >
                        Accepted
                    </button>
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0 -ml-px"
                        :class="
                            activeTab === 'pending'
                                ? 'bg-warning-500 text-white border-warning-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="activeTab = 'pending'"
                    >
                        In Progress
                    </button>
                    <button
                        class="px-8 py-4 text-lg font-medium transition-all rounded-t-2xl border-2 border-b-0 -ml-px"
                        :class="
                            activeTab === 'rejected'
                                ? 'bg-error-500 text-white border-error-500 z-10'
                                : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400 border-gray-300 dark:border-gray-600'
                        "
                        @click="activeTab = 'rejected'"
                    >
                        Rejected
                    </button>
                </div>
                <div class="border-b-2 border-gray-300 dark:border-gray-600 -mt-0.5"></div>
            </div>

            <!-- Application Count and Sort -->
            <div class="flex items-center justify-between mb-5">
                <p class="text-lg font-semibold text-gray-700 dark:text-gray-300">
                    {{ totalApplications }} Applications
                </p>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-gray-600 dark:text-gray-400">Sort by:</span>
                    <USelectMenu
                        v-model="sortBy"
                        :options="sortOptions"
                        value-attribute="value"
                        option-attribute="label"
                        class="w-40"
                    />
                </div>
            </div>

            <!-- Applications List -->
            <div v-if="filteredApplications.length === 0" class="text-center py-10">
                <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                <p class="text-gray-500 text-lg">No applications found.</p>
                <p class="text-gray-400 text-sm mt-2">
                    {{
                        activeTab === "pending"
                            ? "You don't have any pending applications."
                            : activeTab === "accepted"
                              ? "You don't have any accepted applications yet."
                              : "You don't have any rejected applications."
                    }}
                </p>
            </div>

            <div v-else class="flex flex-wrap gap-10">
                <StudentApplicationCard
                    v-for="application in filteredApplications"
                    :key="`${application.jobId}-${application.userId}`"
                    class="h-[18em] w-full lg:w-[25em] drop-shadow-md"
                    :job-id="application.jobId"
                    :position="application.position || 'Position Not Available'"
                    :company-name="application.companyName || 'Unknown Company'"
                    :company-logo="
                        application.photoId
                            ? `${config.public.apiBaseUrl}/files/${application.photoId}`
                            : undefined
                    "
                    :job-type="application.jobType"
                    :experience="application.experience"
                    :min-salary="application.minSalary || 0"
                    :max-salary="application.maxSalary || 0"
                    :status="application.status"
                    :applied-date="formatDate(application.createdAt)"
                    @withdraw="handleWithdrawApplication(application.jobId)"
                    @view-details="handleViewDetails(application.jobId)"
                />
            </div>

            <!-- Student Pagination -->
            <div v-if="totalApplicationPages > 1" class="flex justify-center mt-8">
                <UPagination
                    v-model:page="currentApplicationPage"
                    :total="totalApplications"
                    :items-per-page="itemsPerPage"
                    show-edges
                    :sibling-count="1"
                />
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
import type { JobPost } from "~/data/datatypes";

interface JobApplicationResponse {
    createdAt: string;
    jobId: number;
    userId: string;
    phone: string;
    email: string;
    status: "pending" | "accepted" | "rejected";
    // Job details included in response
    jobName: string;
    position: string;
    companyName: string;
    photoId?: string;
    bannerId?: string;
    jobType?: string;
    experience?: string;
    minSalary?: number;
    maxSalary?: number;
    location?: string;
    approvalStatus?: string;
}

const userRole = ref<string>("viewer");

definePageMeta({
    layout: "viewer",
    middleware: "auth",
});

const openJobPostForm = ref(false);

const data = ref<JobPost[]>([]);
const studentApplications = ref<JobApplicationResponse[]>([]);
const totalJobs = ref<number>(0);
const totalApplications = ref<number>(0);
const activeTab = ref<"pending" | "accepted" | "rejected">("pending");
const companyActiveTab = ref<"pending" | "accepted" | "rejected">("accepted");
const sortBy = ref("name");
const currentJobPage = ref(1);
const currentApplicationPage = ref(1);
const itemsPerPage = 15;

const sortOptions = [
    { label: "Name", value: "name" },
    { label: "Date (Newest)", value: "date-desc" },
    { label: "Date (Oldest)", value: "date-asc" },
];

const api = useApi();
const { add: addToast } = useToast();
const config = useRuntimeConfig();

const fetchJobs = async () => {
    // Only fetch jobs for companies
    if (userRole.value !== "company") return;

    try {
        const offset = (currentJobPage.value - 1) * itemsPerPage;
        const response = await api.get("/jobs", {
            params: {
                limit: itemsPerPage,
                offset: offset,
                approvalStatus: companyActiveTab.value,
            },
        });
        data.value = response.data.jobs || [];
        totalJobs.value = response.data.total || 0;
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

const fetchStudentApplications = async () => {
    // Only fetch applications for students
    if (userRole.value !== "student") return;

    try {
        const offset = (currentApplicationPage.value - 1) * itemsPerPage;
        const response = await api.get("/applications", {
            params: {
                limit: itemsPerPage,
                offset: offset,
                status: activeTab.value,
                sortBy: sortBy.value,
            },
        });
        // Handle the response data properly
        const applications = response.data.applications || [];
        studentApplications.value = applications;
        totalApplications.value = response.data.total || 0;
    } catch (error) {
        const apiError = error as { message?: string };
        console.error("Failed to fetch applications:", apiError.message || "Unknown error", error);
        // Don't show error toast for empty results
        studentApplications.value = [];
        totalApplications.value = 0;
    }
};

// No need to filter since backend already filters by status
const filteredApplications = computed(() => {
    return studentApplications.value;
});

// No need to filter since backend already filters by approval status
const filteredJobs = computed(() => {
    return data.value;
});

// Watch sortBy changes and refetch data
watch(sortBy, () => {
    currentApplicationPage.value = 1;
    fetchStudentApplications();
});

const totalJobPages = computed(() => {
    return Math.ceil(totalJobs.value / itemsPerPage);
});

const totalApplicationPages = computed(() => {
    return Math.ceil(totalApplications.value / itemsPerPage);
});

// Watch for tab changes and reset to page 1
watch(companyActiveTab, () => {
    currentJobPage.value = 1;
    fetchJobs();
});

watch(activeTab, () => {
    currentApplicationPage.value = 1;
    fetchStudentApplications();
});

// Watch for page changes and fetch data
watch(currentJobPage, () => {
    fetchJobs();
});

watch(currentApplicationPage, () => {
    fetchStudentApplications();
});

const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - date.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
        return "Today";
    } else if (diffDays === 1) {
        return "1 day ago";
    } else if (diffDays < 7) {
        return `${diffDays} days ago`;
    } else if (diffDays < 30) {
        const weeks = Math.floor(diffDays / 7);
        return `${weeks} ${weeks === 1 ? "week" : "weeks"} ago`;
    } else {
        return date.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
        });
    }
};

const handleWithdrawApplication = async (jobId: number) => {
    if (!confirm("Are you sure you want to withdraw this application?")) {
        return;
    }

    try {
        await api.delete(`/applications/${jobId}`, {
            withCredentials: true,
        });
        addToast({
            title: "Success",
            description: "Application withdrawn successfully.",
            color: "success",
        });
        // Refresh applications
        await fetchStudentApplications();
    } catch (error) {
        const apiError = error as { message?: string };
        console.error("Failed to withdraw application:", apiError.message || "Unknown error");
        addToast({
            title: "Error",
            description: "Failed to withdraw application. Please try again.",
            color: "error",
        });
    }
};

const handleViewDetails = (jobId: number) => {
    navigateTo(`/jobs/${jobId}`);
};

const isLoading = ref(true);

onMounted(async () => {
    isLoading.value = true;
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }
    try {
        if (userRole.value === "company") {
            await fetchJobs();
        } else if (userRole.value === "student") {
            await fetchStudentApplications();
        }
    } catch (error) {
        console.error("Error during onMounted:", error);
    } finally {
        isLoading.value = false;
    }
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
