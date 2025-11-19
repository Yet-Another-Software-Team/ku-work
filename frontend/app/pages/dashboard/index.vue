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

            <!-- Tab Navigation (match job detail tabs) -->
            <section class="h-[3em] overflow-hidden border-b-1 my-5">
                <div class="flex flex-row gap-2 h-[6em] max-w-[40em] left-0 top-0">
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="companyTabClasses('accepted')"
                        @click="companyActiveTab = 'accepted'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Accepted</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="companyTabClasses('pending')"
                        @click="companyActiveTab = 'pending'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">In Progress</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="companyTabClasses('rejected')"
                        @click="companyActiveTab = 'rejected'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Rejected</p>
                    </div>
                </div>
            </section>

            <!-- Job Count -->
            <div class="flex items-center justify-between mb-5">
                <p class="text-lg font-semibold text-gray-700 dark:text-gray-300">
                    {{ totalJobs }} Jobs
                </p>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-gray-600 dark:text-gray-400">Sort by:</span>
                    <USelectMenu
                        v-model="companySortBy"
                        :options="companySortOptions"
                        value-attribute="value"
                        option-attribute="label"
                        class="w-40"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />

            <div class="flex flex-wrap gap-10">
                <UModal v-model:open="openJobPostCreateForm">
                    <div
                        class="cursor-pointer h-[18em] w-full lg:w-[25em] drop-shadow-md border-primary-700 hover:bg-primary-600/10 border-dashed border-3 rounded-xl flex items-center justify-center flex-col transition-all duration-150"
                    >
                        <Icon name="ic:baseline-plus" size="4em" mode="svg" />
                        <p class="text-xl font-semibold">Create new job post</p>
                    </div>
                    <template #content>
                        <JobPostCreateForm @close="handleJobFormClose" />
                    </template>
                </UModal>

                <JobPostDashboard
                    v-for="job in filteredJobs"
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
                    @update:notify-on-application="
                        (value: boolean) => updateJobNotify(job.id, value)
                    "
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
        </div>

        <!-- Student Dashboard -->
        <div v-else-if="userRole === 'student' && !isLoading">
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">
                Student Dashboard
            </h1>

            <!-- Tab Navigation (match job detail tabs) -->
            <section class="h-[3em] overflow-hidden border-b-1 my-5">
                <div class="flex flex-row gap-2 h-[6em] max-w-[40em] left-0 top-0">
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="studentTabClasses('accepted')"
                        @click="activeTab = 'accepted'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Accepted</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="studentTabClasses('pending')"
                        @click="activeTab = 'pending'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">In Progress</p>
                    </div>
                    <div
                        class="hover:cursor-pointer transition-all duration-150 text-center"
                        :class="studentTabClasses('rejected')"
                        @click="activeTab = 'rejected'"
                    >
                        <p class="font-bold px-5 py-1 text-2xl">Rejected</p>
                    </div>
                </div>
            </section>

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
            <hr class="w-full my-5" />

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

const openJobPostCreateForm = ref(false);

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

const companySortOptions = [
    { label: "Latest", value: "latest" },
    { label: "Oldest", value: "oldest" },
    { label: "Name A-Z", value: "name_az" },
    { label: "Name Z-A", value: "name_za" },
];
const companySortBy = ref("latest");

const api = useApi();
const { add: addToast } = useToast();
const config = useRuntimeConfig();

// Shared tab class helpers (match job detail page)
function classesFor(active: boolean, kind: "pending" | "accepted" | "rejected"): string {
    if (active) {
        if (kind === "pending")
            return "bg-yellow-200 flex flex-col border rounded-3xl w-1/3 text-yellow-800 hover:bg-yellow-300";
        if (kind === "accepted")
            return "bg-green-200 flex flex-col border rounded-3xl w-1/3 text-primary-800 hover:bg-primary-300";
        return "bg-error-200 flex flex-col border rounded-3xl w-1/3 text-error-800 hover:bg-error-300";
    }
    return "bg-gray-200 flex flex-col border rounded-3xl w-1/3 text-gray-500 hover:bg-gray-300";
}
function companyTabClasses(tab: "pending" | "accepted" | "rejected") {
    return classesFor(companyActiveTab.value === tab, tab);
}
function studentTabClasses(tab: "pending" | "accepted" | "rejected") {
    return classesFor(activeTab.value === tab, tab);
}

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
    const list = [...data.value];
    switch (companySortBy.value) {
        case "latest":
            return list.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) => (a.name || "").localeCompare(b.name || ""));
        case "name_za":
            return list.sort((a, b) => (b.name || "").localeCompare(a.name || ""));
        default:
            return list;
    }
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

const updateJobNotify = (id: number, value: boolean) => {
    const job = data.value.find((job) => job.id === id);
    if (!job) return;
    const oldValue: boolean = job.notifyOnApplication || false;
    job.notifyOnApplication = value;
    api.patch(
        `/jobs/${job.id}`,
        {
            notifyOnApplication: value,
        },
        {
            withCredentials: true,
        }
    )
        .then((x) => {
            if (x.status !== 200) return Promise.reject(x);
        })
        .catch((_) => {
            job.notifyOnApplication = oldValue;
        });
};

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
    openJobPostCreateForm.value = false;
    // Refresh jobs list after posting (only for companies)
    if (userRole.value === "company") {
        fetchJobs();
    }
};
</script>
