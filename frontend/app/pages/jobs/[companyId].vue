<template>
    <div>
        <!-- Loading State -->
        <div v-if="isLoading" class="pt-5 pb-2">
            <USkeleton class="h-12 w-1/4 mb-6" />
            <div class="rounded-lg">
                <!-- Banner Skeleton -->
                <USkeleton class="h-32 rounded-t-lg mb-4" />

                <!-- Profile Section Skeleton -->
                <div class="flex flex-wrap gap-5 mb-6">
                    <USkeleton class="w-40 h-40 rounded-full" />
                    <div class="flex-1 space-y-3">
                        <USkeleton class="h-8 w-64" />
                        <USkeleton class="h-6 w-96" />
                    </div>
                </div>

                <hr class="my-6 border-gray-300 dark:border-gray-600" />

                <!-- Content Skeleton -->
                <div class="flex flex-wrap gap-5">
                    <div class="w-48 space-y-3">
                        <USkeleton class="h-6 w-32" />
                        <USkeleton class="h-10 w-full" />
                        <USkeleton class="h-10 w-full" />
                        <USkeleton class="h-10 w-full" />
                    </div>
                    <div class="flex-1 space-y-3">
                        <USkeleton class="h-6 w-32" />
                        <USkeleton class="h-4 w-full" />
                        <USkeleton class="h-4 w-full" />
                        <USkeleton class="h-4 w-3/4" />
                    </div>
                </div>

                <!-- Jobs Section Skeleton -->
                <div class="mt-12">
                    <USkeleton class="h-8 w-48 mb-6" />
                    <div class="space-y-4">
                        <USkeleton class="h-32 w-full" />
                        <USkeleton class="h-32 w-full" />
                        <USkeleton class="h-32 w-full" />
                    </div>
                </div>
            </div>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="pt-5 pb-2">
            <div class="text-center py-16">
                <Icon
                    name="ic:baseline-error-outline"
                    class="w-20 h-20 text-red-500 mx-auto mb-4"
                />
                <h2 class="text-2xl font-semibold text-gray-800 dark:text-gray-200 mb-2">
                    Failed to Load Company Profile
                </h2>
                <p class="text-gray-600 dark:text-gray-400 mb-6">{{ errorMessage }}</p>
                <UButton color="primary" size="lg" icon="ic:baseline-refresh" @click="retryFetch">
                    Retry
                </UButton>
            </div>
        </div>

        <!-- Company Not Found -->
        <div v-else-if="!companyExists" class="pt-5 pb-2">
            <div class="text-center py-16">
                <Icon name="ic:baseline-business" class="w-20 h-20 text-gray-400 mx-auto mb-4" />
                <h2 class="text-2xl font-semibold text-gray-800 dark:text-gray-200 mb-2">
                    Company Not Found
                </h2>
                <p class="text-gray-600 dark:text-gray-400 mb-6">
                    The company profile you're looking for doesn't exist or has been removed.
                </p>
                <UButton
                    color="primary"
                    variant="outline"
                    size="lg"
                    icon="ic:baseline-arrow-back"
                    @click="navigateTo('/jobs')"
                >
                    Back to Job Board
                </UButton>
            </div>
        </div>

        <!-- Company Profile Display -->
        <div v-else-if="['student', 'viewer', 'company'].includes(userRole)" class="pt-5 pb-2">
            <CompanyProfileCard :is-owner="false" :company-id="companyId" />

            <!-- Company Jobs Section -->
            <div class="mt-12">
                <div class="flex items-center justify-between mb-6">
                    <h2 class="text-3xl font-bold text-gray-900 dark:text-white">Open Positions</h2>
                    <UBadge v-if="companyJobs.length > 0" size="lg" color="primary" variant="solid">
                        {{ companyJobs.length }} {{ companyJobs.length === 1 ? "Job" : "Jobs" }}
                    </UBadge>
                </div>

                <!-- Loading Jobs -->
                <div v-if="loadingJobs" class="space-y-4">
                    <USkeleton class="h-32 w-full" />
                    <USkeleton class="h-32 w-full" />
                    <USkeleton class="h-32 w-full" />
                </div>

                <!-- No Jobs Available -->
                <div
                    v-else-if="companyJobs.length === 0"
                    class="text-center py-12 bg-gray-50 dark:bg-gray-800/50 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-700"
                >
                    <Icon
                        name="ic:baseline-work-off"
                        class="w-16 h-16 text-gray-400 dark:text-gray-500 mx-auto mb-4"
                    />
                    <p class="text-gray-600 dark:text-gray-400 text-lg font-medium mb-2">
                        No Open Positions
                    </p>
                    <p class="text-gray-500 dark:text-gray-500 text-sm">
                        This company doesn't have any job openings at the moment. Check back later!
                    </p>
                </div>

                <!-- Jobs List with Extended Cards -->
                <div v-else class="space-y-6">
                    <div
                        v-for="job in companyJobs"
                        :key="job.id"
                        class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-sm hover:shadow-lg transition-all duration-200 overflow-hidden cursor-pointer"
                        @click="navigateTo(`/jobs/${job.id}`)"
                    >
                        <!-- Card Header -->
                        <div class="flex items-start gap-4 p-6 pb-4">
                            <!-- Company Logo -->
                            <div
                                class="w-16 h-16 rounded-lg border border-gray-200 dark:border-gray-700 flex items-center justify-center overflow-hidden flex-shrink-0 bg-white"
                            >
                                <img
                                    v-if="job.photoId"
                                    :src="`${config.public.apiBaseUrl}/files/${job.photoId}`"
                                    alt="Company Logo"
                                    class="w-full h-full object-cover"
                                />
                                <Icon
                                    v-else
                                    name="ic:baseline-business"
                                    class="w-8 h-8 text-gray-400"
                                />
                            </div>

                            <!-- Job Info -->
                            <div class="flex-1 min-w-0">
                                <h3
                                    class="text-xl font-semibold text-gray-900 dark:text-white mb-1"
                                >
                                    {{ job.position }}
                                </h3>
                                <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">
                                    {{ job.name }}
                                </p>
                                <div class="flex flex-wrap items-center gap-2">
                                    <UBadge
                                        v-if="job.jobType"
                                        color="primary"
                                        variant="subtle"
                                        size="sm"
                                        class="capitalize"
                                    >
                                        {{ formatJobType(job.jobType) }}
                                    </UBadge>
                                    <UBadge
                                        v-if="job.experience"
                                        color="neutral"
                                        variant="subtle"
                                        size="sm"
                                        class="capitalize"
                                    >
                                        {{ formatExperience(job.experience) }}
                                    </UBadge>
                                    <span
                                        class="text-sm text-gray-600 dark:text-gray-400 flex items-center gap-1"
                                    >
                                        <Icon name="ic:baseline-location-on" class="w-4 h-4" />
                                        {{ job.location }}
                                    </span>
                                </div>
                            </div>

                            <!-- Salary and Date -->
                            <div class="text-right flex-shrink-0">
                                <p class="text-lg font-bold text-gray-900 dark:text-white mb-1">
                                    {{ formatSalary(job.minSalary, job.maxSalary) }}
                                </p>
                                <p class="text-xs text-gray-500 dark:text-gray-400">
                                    {{ timeAgo(job.createdAt) }}
                                </p>
                            </div>
                        </div>

                        <!-- Card Body -->
                        <div class="px-6 pb-6">
                            <!-- Job Description -->
                            <p class="text-sm text-gray-700 dark:text-gray-300 line-clamp-2 mb-4">
                                {{ job.description || "No description available." }}
                            </p>

                            <!-- Card Footer -->
                            <div
                                class="flex items-center justify-between pt-4 border-t border-gray-200 dark:border-gray-700"
                            >
                                <div
                                    class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400"
                                >
                                    <span class="flex items-center gap-1">
                                        <Icon name="ic:baseline-access-time" class="w-4 h-4" />
                                        {{ job.duration || "Not specified" }}
                                    </span>
                                </div>
                                <UButton
                                    color="primary"
                                    size="sm"
                                    icon="ic:baseline-arrow-forward"
                                    trailing
                                    @click.stop="navigateTo(`/jobs/${job.id}`)"
                                >
                                    View Details
                                </UButton>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Access denied for other roles -->
        <div v-else class="text-center py-8">
            <Icon name="ic:baseline-lock" class="w-20 h-20 text-gray-400 mx-auto mb-4" />
            <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">Access Denied</h2>
            <p class="text-gray-500 mt-2">You don't have permission to view this page.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";

const userRole = ref<string>("viewer");
const route = useRoute();
const companyId = computed(() => route.params.companyId as string);
const isLoading = ref(true);
const loadingJobs = ref(false);
const error = ref(false);
const errorMessage = ref("");
const companyExists = ref(true);
const companyJobs = ref<JobPost[]>([]);

const api = useApi();
const { add: addToast } = useToast();
const config = useRuntimeConfig();

definePageMeta({
    layout: "viewer",
    middleware: "auth",
});

// Validate company ID exists
const validateCompanyId = () => {
    if (!companyId.value) {
        error.value = true;
        errorMessage.value = "No company ID provided in the URL.";
        companyExists.value = false;
        return false;
    }
    return true;
};

// Check if company exists
const checkCompanyExists = async () => {
    if (!validateCompanyId()) {
        isLoading.value = false;
        return;
    }

    try {
        const response = await api.get(`/company/${companyId.value}`);

        if (response.status === 200 && response.data) {
            companyExists.value = true;
            error.value = false;
            // Fetch company jobs after confirming company exists
            await fetchCompanyJobs();
        } else {
            companyExists.value = false;
            errorMessage.value = "Company profile not found.";
        }
    } catch (err) {
        const apiError = err as { status?: number; message?: string };

        if (apiError.status === 404) {
            companyExists.value = false;
            errorMessage.value = "Company profile not found.";
        } else {
            error.value = true;
            errorMessage.value =
                apiError.message || "Failed to load company profile. Please try again.";

            addToast({
                title: "Error",
                description: errorMessage.value,
                color: "error",
            });
        }
    } finally {
        isLoading.value = false;
    }
};

// Fetch company jobs
const fetchCompanyJobs = async () => {
    if (!companyId.value) return;

    loadingJobs.value = true;
    try {
        const response = await api.get("/jobs");

        if (response.status === 200 && response.data) {
            // Filter jobs by company ID and only show open/accepted jobs
            companyJobs.value = (response.data.jobs || [])
                .filter(
                    (job: JobPost) =>
                        job.companyId === companyId.value &&
                        job.open &&
                        job.approvalStatus === "accepted"
                )
                .sort(
                    (a: JobPost, b: JobPost) =>
                        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
                );
        }
    } catch (err) {
        console.error("Error fetching company jobs:", err);
        // Don't show error for jobs, just set empty array
        companyJobs.value = [];
    } finally {
        loadingJobs.value = false;
    }
};

// Format job type
const formatJobType = (type: string): string => {
    const typeMap: Record<string, string> = {
        fulltime: "Full Time",
        parttime: "Part Time",
        contract: "Contract",
        casual: "Casual",
        internship: "Internship",
    };
    return typeMap[type.toLowerCase()] || type;
};

// Format experience
const formatExperience = (exp: string): string => {
    const expMap: Record<string, string> = {
        newgrad: "New Grad",
        junior: "Junior",
        senior: "Senior",
        manager: "Manager",
        internship: "Internship",
    };
    return expMap[exp.toLowerCase()] || exp;
};

// Format salary
const formatSalary = (min: number, max: number): string => {
    if (min === 0 && max === 0) return "Negotiable";
    if (min === max) return `${formatNumber(min)}k Baht`;
    return `${formatNumber(min)}k - ${formatNumber(max)}k Baht`;
};

const formatNumber = (num: number): string => {
    return new Intl.NumberFormat("en").format(num);
};

// Time ago helper
const timeAgo = (createdAt: string): string => {
    const createdDate = new Date(createdAt);
    const now = new Date();

    const diffMs = now.getTime() - createdDate.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    const diffMonth = Math.floor(diffDay / 30);

    if (diffMonth > 0) return `${diffMonth} month${diffMonth > 1 ? "s" : ""} ago`;
    if (diffDay > 0) return `${diffDay} day${diffDay > 1 ? "s" : ""} ago`;
    if (diffHour > 0) return `${diffHour} hour${diffHour > 1 ? "s" : ""} ago`;
    if (diffMin > 0) return `${diffMin} minute${diffMin > 1 ? "s" : ""} ago`;
    return "just now";
};

// Retry fetching company data
const retryFetch = async () => {
    isLoading.value = true;
    error.value = false;
    errorMessage.value = "";
    companyExists.value = true;
    companyJobs.value = [];

    await checkCompanyExists();
};

onMounted(async () => {
    if (import.meta.client) {
        userRole.value = localStorage.getItem("role") || "viewer";
    }

    await checkCompanyExists();
});

// Watch for company ID changes (if user navigates to different company)
watch(companyId, async (newId, oldId) => {
    if (newId !== oldId && newId) {
        isLoading.value = true;
        companyJobs.value = [];
        await checkCompanyExists();
    }
});
</script>
