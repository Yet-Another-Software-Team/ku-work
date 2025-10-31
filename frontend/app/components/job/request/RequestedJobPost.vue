<template>
    <div class="py-6">
        <!-- Header -->
        <a href="/admin/dashboard">
            <h1
                class="flex items-center text-2xl text-primary-800 dark:text-primary font-bold mt-6 mb-4 gap-1 cursor-pointer"
            >
                <Icon name="iconoir:nav-arrow-left" class="items-center" />
                <span>Back</span>
            </h1>
        </a>

        <div class="px-4 bg-white dark:bg-gray-800 shadow-md max-w-7xl mx-auto w-full">
            <div v-if="isLoading" class="space-y-4">
                <USkeleton class="h-10 w-1/2" />
                <USkeleton class="h-6 w-1/3" />
                <USkeleton class="h-40 w-full" />
            </div>
            <div v-else-if="job" class="w-full">
                <!-- First section -->
                <div class="flex items-center mb-7">
                    <!-- Profile -->
                    <span class="flex-shrink-0">
                        <img
                            v-if="job.photoId"
                            :src="photo"
                            alt="Company Logo"
                            class="rounded-full size-[6em] mt-2 object-cover"
                        />
                        <img
                            v-else
                            src="~/assets/images/background.png"
                            alt="Company Logo"
                            class="rounded-full size-[6em] object-cover"
                        />
                    </span>
                    <span class="mx-3 space-y-1">
                        <h1 class="text-xl font-bold">{{ job.name }}</h1>
                        <span>
                            <NuxtLink :to="`/jobs/company?id=${job.companyId}`">
                                <h2 class="text-primary-700 text-base font-semibold">
                                    {{ job.companyName }}
                                </h2>
                            </NuxtLink>
                            <p class="text-xs">{{ timeAgo(job.createdAt) }}</p>
                        </span>
                    </span>
                </div>
                <!-- Divider -->
                <USeparator class="mt-2" />
                <div class="flex flex-row justify-center text-sm w-full">
                    <span class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Job Type</p>
                        <p class="font-bold">{{ job.jobType }}</p>
                    </span>
                    <span>
                        <USeparator
                            orientation="vertical"
                            color="neutral"
                            class="h-full items-stretch"
                            size="sm"
                        />
                    </span>
                    <span class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Experience</p>
                        <p class="font-bold">{{ job.experience }}</p>
                    </span>
                    <span>
                        <USeparator
                            orientation="vertical"
                            color="neutral"
                            class="h-full items-stretch"
                            size="sm"
                        />
                    </span>
                    <span class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Salary Range</p>
                        <p class="font-bold">
                            {{ formatSalary(job.minSalary) }}
                            -
                            {{ formatSalary(job.maxSalary) }}
                        </p>
                    </span>
                </div>
                <USeparator class="mb-2" />
                <!-- About job -->
                <div>
                    <h2 class="font-semibold">About This Job</h2>
                    <p>
                        <span class="text-primary-700 font-semibold">Position Title: </span>
                        {{ job.position }}
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold capitalize">Location: </span>
                        {{ job.location }}
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold">Duration: </span>
                        {{ job.duration }}
                    </p>
                    <p class="whitespace-pre-line">
                        <span class="text-primary-700 font-semibold">Description:</span>
                        <br />
                        {{ job.description }}
                    </p>
                </div>

                <!-- Admin actions -->
                <div class="flex flex-col items-end justify-end ml-3 p-2 h-full w-full">
                    <div class="flex gap-3 mt-6">
                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl"
                            variant="outline"
                            color="error"
                            label="Decline"
                            :icon="'iconoir:xmark'"
                            :loading="actionLoading === 'decline'"
                            @click.stop="showRejectModal = true"
                        />
                        <UModal
                            v-model:open="showRejectModal"
                            :dismissible="false"
                            :ui="{ overlay: 'fixed inset-0 bg-black/50' }"
                        >
                            <UCard
                                title="Reject Post?"
                                :ui="{ header: 'pb-2', title: 'text-xl font-semibold text-primary-800 dark:text-primary' }"
                            >
                                <p>Are you sure you want to <strong>decline</strong> {{ job?.name }}?</p>
                                <template #footer>
                                    <div class="flex justify-end gap-2">
                                        <UButton variant="outline" color="neutral" label="Cancel" @click="showRejectModal = false" />
                                        <UButton color="error" label="Decline" @click="() => { declineRequest(); showRejectModal = false; }" />
                                    </div>
                                </template>
                            </UCard>
                        </UModal>

                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl"
                            variant="outline"
                            color="primary"
                            label="Accept"
                            :icon="'iconoir:check'"
                            :loading="actionLoading === 'accept'"
                            @click.stop="showAcceptModal = true"
                        />
                        <UModal
                            v-model:open="showAcceptModal"
                            :dismissible="false"
                            :ui="{ overlay: 'fixed inset-0 bg-black/50' }"
                        >
                            <UCard
                                title="Accept Post?"
                                :ui="{ header: 'pb-2', title: 'text-xl font-semibold text-primary-800 dark:text-primary' }"
                            >
                                <p>Are you sure you want to <strong>accept</strong> {{ job?.name }}?</p>
                                <template #footer>
                                    <div class="flex justify-end gap-2">
                                        <UButton variant="outline" color="neutral" label="Cancel" @click="showAcceptModal = false" />
                                        <UButton color="primary" label="Accept" @click="() => { acceptRequest(); showAcceptModal = false; }" />
                                    </div>
                                </template>
                            </UCard>
                        </UModal>
                    </div>
                </div>
            </div>
            <div v-else class="text-gray-500">No pending job post</div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";
const props = defineProps<{ requestId: string }>();
const api = useApi();
const config = useRuntimeConfig();
const toast = useToast();

const isLoading = ref(true);


const showAcceptModal = ref(false);
const showRejectModal = ref(false);

const job = ref<JobPost | null>(null);
const photo = ref<string>("");

function formatSalary(salary: number): string {
    return new Intl.NumberFormat("en", { notation: "compact" }).format(salary);
}

//Compute time ago
function timeAgo(createdAt: string): string {
    const createdDate = new Date(createdAt);
    const now = new Date();
    const diff = Math.floor((now.getTime() - createdDate.getTime()) / 1000);
    const min = Math.floor(diff / 60);
    const hour = Math.floor(min / 60);
    const day = Math.floor(hour / 24);
    const month = Math.floor(day / 30);
    if (month > 0) return `${month} month${month > 1 ? "s" : ""} ago`;
    if (day > 0) return `${day} day${day > 1 ? "s" : ""} ago`;
    if (hour > 0) return `${hour} hour${hour > 1 ? "s" : ""} ago`;
    if (min > 0) return `${min} minute${min > 1 ? "s" : ""} ago`;
    return "just now";
}

async function loadJob() {
    const idParam = props.requestId ?? (useRoute().params.id as string);
    if (!idParam) return;
    isLoading.value = true;
    try {
        const id = Number(idParam);
        const response = await api.get<JobPost | { job: JobPost }>(`/jobs/${id}`, {
            withCredentials: true,
        });
        const payload = response.data as JobPost | { job: JobPost };
        job.value = ("job" in payload ? payload.job : payload) as JobPost;
        photo.value = config.public.apiBaseUrl + "/files/" + job.value.photoId;
        console.log("Fetched job data:", job.value);
    } catch (error) {
        console.error("Error fetching job data:", error);
        toast.add({
            title: "Error fetching job data",
            description: String(error),
            color: "error",
        });
    } finally {
        isLoading.value = false;
    }
}

// Accept request
async function acceptRequest() {
    try {
        actionLoading.value = "accept";
        await api.post(
            `/jobs/${props.requestId}/approval`,
            { approve: true },
            { withCredentials: true }
        );
        toast.add({ title: "Job approved!", description: job.value?.name || "", color: "success" });
        navigateTo("/admin/dashboard", { replace: true });
    } catch (e: unknown) {
        api.showErrorToast(api.handleError(e), "Approve failed");
    } finally {
        actionLoading.value = null;
    }
}

// Decline request
async function declineRequest() {
    try {
        actionLoading.value = "decline";
        await api.post(
            `/jobs/${props.requestId}/approval`,
            { approve: false },
            { withCredentials: true }
        );
        toast.add({ title: "Job declined!", description: job.value?.name || "", color: "error" });
        navigateTo("/admin/dashboard", { replace: true });
    } catch (e: unknown) {
        api.showErrorToast(api.handleError(e), "Decline failed");
    } finally {
        actionLoading.value = null;
    }
}

onMounted(loadJob);
</script>
