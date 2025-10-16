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
            <div class="flex flex-wrap gap-10">
                <div v-if="data.length === 0" class="w-full text-center py-10">
                    <p class="text-neutral-400 dark:text-neutral-500 text-xl">
                        No jobs posted yet.
                    </p>
                </div>
                <JobCardCompany
                    v-for="job in data"
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
            <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">Dashboard</h1>

            <!-- Tab Navigation -->
            <div class="flex items-center gap-3 mb-5">
                <UButton
                    :variant="activeTab === 'pending' ? 'solid' : 'ghost'"
                    :color="activeTab === 'pending' ? 'warning' : 'neutral'"
                    size="lg"
                    @click="activeTab = 'pending'"
                >
                    In Progress
                </UButton>
                <UButton
                    :variant="activeTab === 'accepted' ? 'solid' : 'ghost'"
                    :color="activeTab === 'accepted' ? 'primary' : 'neutral'"
                    size="lg"
                    @click="activeTab = 'accepted'"
                >
                    Accepted
                </UButton>
                <UButton
                    :variant="activeTab === 'rejected' ? 'solid' : 'ghost'"
                    :color="activeTab === 'rejected' ? 'error' : 'neutral'"
                    size="lg"
                    @click="activeTab = 'rejected'"
                >
                    Rejected
                </UButton>
            </div>

            <!-- Application Count and Sort -->
            <div class="flex items-center justify-between mb-5">
                <p class="text-lg font-semibold text-gray-700 dark:text-gray-300">
                    {{ filteredApplications.length }} Applicants
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

            <div v-else class="flex flex-col gap-4">
                <StudentApplicationCard
                    v-for="application in sortedApplications"
                    :key="`${application.jobId}-${application.userId}`"
                    :job-id="application.jobId"
                    :position="application.jobDetails?.position || 'Position Not Available'"
                    :company-name="application.jobDetails?.companyName || 'Unknown Company'"
                    :company-logo="
                        application.jobDetails?.photoId
                            ? `${config.public.apiBaseUrl}/files/${application.jobDetails.photoId}`
                            : undefined
                    "
                    :job-type="application.jobDetails?.jobType"
                    :experience="application.jobDetails?.experience"
                    :min-salary="application.jobDetails?.minSalary || 0"
                    :max-salary="application.jobDetails?.maxSalary || 0"
                    :status="application.status"
                    :applied-date="formatDate(application.createdAt)"
                    @withdraw="handleWithdrawApplication(application.jobId)"
                    @view-details="handleViewDetails(application.jobId)"
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
import type { JobPost } from "~/data/mockData";

interface JobApplicationResponse {
    createdAt: string;
    jobId: number;
    userId: string;
    phone: string;
    email: string;
    status: "pending" | "accepted" | "rejected";
}

interface JobApplicationWithDetails extends JobApplicationResponse {
    jobDetails?: {
        position: string;
        companyName: string;
        photoId?: string;
        jobType?: string;
        experience?: string;
        minSalary?: number;
        maxSalary?: number;
    };
}

const userRole = ref<string>("viewer");

definePageMeta({
    layout: "viewer",
    middleware: "auth",
});

const openJobPostForm = ref(false);

const data = ref<JobPost[]>([]);
const studentApplications = ref<JobApplicationWithDetails[]>([]);
const activeTab = ref<"pending" | "accepted" | "rejected">("pending");
const sortBy = ref("name");

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
        const response = await api.get("/jobs");
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

const fetchStudentApplications = async () => {
    // Only fetch applications for students
    if (userRole.value !== "student") return;

    try {
        const response = await api.get("/applications");
        const applications = response.data || [];

        // Fetch job details for each application
        const applicationsWithDetails = await Promise.all(
            applications.map(async (app: JobApplicationResponse) => {
                try {
                    const jobResponse = await api.get(`/jobs/${app.jobId}`);
                    return {
                        ...app,
                        jobDetails: {
                            position: jobResponse.data.position,
                            companyName: jobResponse.data.companyName,
                            photoId: jobResponse.data.photoId,
                            jobType: jobResponse.data.jobType,
                            experience:
                                jobResponse.data.experienceType || jobResponse.data.experience,
                            minSalary: jobResponse.data.minSalary,
                            maxSalary: jobResponse.data.maxSalary,
                        },
                    };
                } catch (error) {
                    console.error(`Failed to fetch job details for job ${app.jobId}:`, error);
                    return app;
                }
            })
        );

        studentApplications.value = applicationsWithDetails;
    } catch (error) {
        const apiError = error as { message?: string };
        console.error("Failed to fetch applications:", apiError.message || "Unknown error");
        addToast({
            title: "Error",
            description: "Failed to fetch applications. Please refresh the page.",
            color: "error",
        });
    }
};

const filteredApplications = computed(() => {
    return studentApplications.value.filter((app) => app.status === activeTab.value);
});

const sortedApplications = computed(() => {
    const apps = [...filteredApplications.value];

    switch (sortBy.value) {
        case "name":
            return apps.sort((a, b) => {
                const nameA = a.jobDetails?.companyName || "";
                const nameB = b.jobDetails?.companyName || "";
                return nameA.localeCompare(nameB);
            });
        case "date-desc":
            return apps.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
        case "date-asc":
            return apps.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        default:
            return apps;
    }
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
