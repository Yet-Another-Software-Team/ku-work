<template>
    <div class="flex">
        <section class="w-full">
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
                    <UButton
                        color="primary"
                        size="lg"
                        icon="ic:baseline-refresh"
                        @click="retryFetch"
                    >
                        Retry
                    </UButton>
                </div>
            </div>

            <!-- Company Not Found -->
            <div v-else-if="!companyExists" class="pt-5 pb-2">
                <div class="text-center py-16">
                    <Icon
                        name="ic:baseline-business"
                        class="w-20 h-20 text-gray-400 mx-auto mb-4"
                    />
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
                    <div class="mb-6">
                        <h2
                            class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-2"
                        >
                            <span>Open Positions</span>
                            <UBadge
                                v-if="companyJobs.length > 0"
                                size="lg"
                                color="primary"
                                variant="solid"
                            >
                                {{ companyJobs.length }}
                            </UBadge>
                        </h2>
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
                            This company doesn't have any job openings at the moment. Check back
                            later!
                        </p>
                    </div>

                    <!-- Jobs List using JobPostComponent -->
                    <section v-else>
                        <section v-for="(job, index) in companyJobs" :key="job.id">
                            <JobPostComponent
                                :is-selected="selectedIndex === index"
                                :data="job"
                                @click="selectedIndex = index"
                            />
                        </section>
                    </section>
                </div>
            </div>

            <!-- Access denied for other roles -->
            <div v-else class="text-center py-8">
                <Icon name="ic:baseline-lock" class="w-20 h-20 text-gray-400 mx-auto mb-4" />
                <h2 class="text-2xl font-semibold text-gray-600 dark:text-gray-400">
                    Access Denied
                </h2>
                <p class="text-gray-500 mt-2">You don't have permission to view this page.</p>
            </div>
        </section>

        <!-- Expanded Job Post -->
        <section v-if="selectedIndex !== null && selectedIndex < companyJobs.length" class="flex">
            <USeparator orientation="vertical" class="w-fit mx-5" color="neutral" size="lg" />
            <section>
                <JobPostExpanded
                    v-if="companyJobs.length > 0"
                    :is-viewer="userRole === 'viewer'"
                    :is-selected="true"
                    :data="companyJobs[selectedIndex]!"
                />
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";

const userRole = ref<string>("viewer");
const route = useRoute();
const companyId = computed(() => route.params.companyId as string);
const isLoading = ref(true);
const loadingJobs = ref(false);
const error = ref(false);
const errorMessage = ref("");
const companyExists = ref(true);
const companyJobs = ref<JobPost[]>([]);
const selectedIndex = ref<number | null>(null);

const api = useApi();
const { add: addToast } = useToast();

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
        // Use API query params to filter on server-side
        const params = new URLSearchParams();
        params.append("companyID", companyId.value);
        params.append("open", "true");
        params.append("approvalStatus", "accepted");
        params.append("limit", "100"); // Get all jobs for this company

        const response = await api.get("/jobs", { params });

        if (response.status === 200 && response.data) {
            companyJobs.value = (response.data.jobs || []).sort(
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
