<template>
    <div :id="'Job-' + requestId" class="flex justify-center items-center w-[28em] h-full">
        <!-- card container is clickable -->
        <div
            class="flex items-stretch w-full h-[7em] m-2 shadow-md/20 bg-[#fdfdfd] dark:bg-[#1f2937] rounded-md cursor-pointer"
            @click="() => navigateToJob(requestId)"
        >
            <!-- logo / image -->
            <div class="flex items-center justify-center w-20 h-full flex-shrink-0">
                <img
                    v-if="logoSrc"
                    :src="logoSrc"
                    alt="Company Logo"
                    class="w-17 h-17 object-cover rounded-md justify-center items-center m-2"
                />
                <img
                    v-else
                    src="/images/company.png"
                    alt="Company Logo"
                    class="w-17 h-17 object-cover rounded-md m-2"
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
                        @click.stop="declineJob"
                    />
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        @click.stop="approveJob"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";

const props = defineProps<{
    requestId: string;
    data?: JobPost;
}>();

const emit = defineEmits<{
    (e: "jobApprovalStatus", id: string): void;
}>();

const job = props.data as JobPost;
const api = useApi();

const toast = useToast();
const logoSrc = ref<string>("");

function navigateToJob(id: string) {
    console.log("Navigating to job post of request:", props.requestId);
    navigateTo(`/admin/dashboard/jobs/${id}`, { replace: false });
}

// Accept request
async function approveJob() {
    console.log("Approveed request:", props.requestId);
    toast.add({
        title: "Job Approved!",
        description: job.name,
        color: "success",
    });
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
    toast.add({
        title: "Job declined!",
        description: job.name,
        color: "error",
    });
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
        if (!newData) return;
        job.companyName = newData.companyName;
        job.id = newData.id;
        job.createdAt = newData.createdAt;
        job.name = newData.name;
        job.companyId = newData.companyId;
        job.photoId = newData.photoId;
        job.bannerId = newData.bannerId;
        job.position = newData.position;
        job.duration = newData.duration;
        job.description = newData.description;
        job.location = newData.location;
        job.jobType = newData.jobType;
        job.experience = newData.experience;
        job.minSalary = newData.minSalary;
        job.maxSalary = newData.maxSalary;
        job.approvalStatus = newData.approvalStatus;
        job.open = newData.open;
        logoSrc.value = newData.photoId
            ? `${useRuntimeConfig().public.apiBaseUrl}/files/${newData.photoId}`
            : "";
    },
    { immediate: true }
);
</script>
