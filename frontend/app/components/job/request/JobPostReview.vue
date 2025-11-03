<template>
    <div :id="'Job-' + requestId" class="flex justify-center items-center w-[28em] h-full">
        <!-- card container is clickable -->
        <div
            class="flex items-stretch w-full h-[7em] m-2 shadow-md bg-[#fdfdfd] dark:bg-[#1f2937] rounded-md cursor-pointer"
            @click="() => navigateToJob(requestId)"
        >
            <!-- logo / image -->
            <div class="flex items-center justify-center w-20 h-full flex-shrink-0">
                <img
                    v-if="logoSrc"
                    :src="logoSrc"
                    alt="Company Logo"
                    class="w-16 h-16 object-cover rounded-md justify-center items-center m-2"
                />
                <img
                    v-else
                    src="/images/company.png"
                    alt="Company Logo"
                    class="w-16 h-16 object-cover rounded-md m-2"
                />
            </div>

            <!-- job data -->
            <div class="flex flex-col justify-center flex-1 p-2 min-w-0">
                <!-- title / position -->
                <p class="overflow-hidden truncate font-semibold max-w-[14em]">
                    {{ job.name }}
                </p>
                <p class="text-xs text-gray-600 dark:text-gray-300 truncate">
                    {{ job.companyName }}
                </p>
            </div>

            <!-- buttons -->
            <div class="flex flex-col items-end justify-end ml-3 p-2 h-full">
                <div class="flex gap-2 mt-2">
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="error"
                        label="Decline"
                        :icon="'iconoir:xmark'"
                        @click.stop="showRejectModal = true"
                    />
                    <UModal
                        v-model:open="showRejectModal"
                        :dismissible="false"
                        :ui="{ overlay: 'fixed inset-0 bg-black/50' }"
                    >
                        <UCard
                            title="Reject Post?"
                            :ui="{
                                header: 'pb-2',
                                title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                            }"
                        >
                            <p>
                                Are you sure you want to <strong>decline</strong> {{ job?.name }}?
                            </p>
                            <template #footer>
                                <div class="flex justify-end gap-2">
                                    <UButton
                                        variant="outline"
                                        color="neutral"
                                        label="Cancel"
                                        @click="showRejectModal = false"
                                    />
                                    <UButton
                                        color="error"
                                        label="Decline"
                                        @click="
                                            () => {
                                                declineJob();
                                                showRejectModal = false;
                                            }
                                        "
                                    />
                                </div>
                            </template>
                        </UCard>
                    </UModal>
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        @click.stop="showAcceptModal = true"
                    />
                    <UModal
                        v-model:open="showAcceptModal"
                        :dismissible="false"
                        :ui="{ overlay: 'fixed inset-0 bg-black/50' }"
                    >
                        <UCard
                            title="Accept Post?"
                            :ui="{
                                header: 'pb-2',
                                title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                            }"
                        >
                            <p>Are you sure you want to <strong>accept</strong> {{ job?.name }}?</p>
                            <template #footer>
                                <div class="flex justify-end gap-2">
                                    <UButton
                                        variant="outline"
                                        color="neutral"
                                        label="Cancel"
                                        @click="showAcceptModal = false"
                                    />
                                    <UButton
                                        color="primary"
                                        label="Accept"
                                        @click="
                                            () => {
                                                approveJob();
                                                showAcceptModal = false;
                                            }
                                        "
                                    />
                                </div>
                            </template>
                        </UCard>
                    </UModal>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";

const props = defineProps<{
    requestId: string;
    data?: JobPost;
}>();

const emit = defineEmits<{
    (e: "jobApprovalStatus", id: string): void;
}>();

const job = ref<JobPost>(props.data!);
const api = useApi();

const toast = useToast();
// Modal visibility state
const showAcceptModal = ref(false);
const showRejectModal = ref(false);
const logoSrc = computed<string>(() =>
    job.value?.photoId ? `${useRuntimeConfig().public.apiBaseUrl}/files/${job.value.photoId}` : ""
);

function navigateToJob(id: string) {
    console.log("Navigating to job post of request:", props.requestId);
    navigateTo(`/admin/dashboard/jobs/${id}`, { replace: false });
}

// Accept request
async function approveJob() {
    console.log("Approved request:", props.requestId);
    toast.add({ title: "Job Approved!", description: job.value?.name, color: "success" });
    await api.post(
        `/jobs/${props.requestId}/approval`,
        { approve: true },
        { withCredentials: true }
    );
    emit("jobApprovalStatus", props.requestId);
}

// Decline request
async function declineJob() {
    console.log("Declined request:", props.requestId);
    toast.add({ title: "Job declined!", description: job.value?.name, color: "error" });
    await api.post(
        `/jobs/${props.requestId}/approval`,
        { approve: false },
        { withCredentials: true }
    );
    emit("jobApprovalStatus", props.requestId);
}
watch(
    () => props.data,
    (newData) => {
        if (newData) job.value = newData;
    }
);
</script>
