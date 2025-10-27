<template>
    <div class="sticky top-0 overflow-y-auto max-h-dvh rounded-xl shadow-md/25 p-5">
        <!-- First section -->
        <div class="flex mb-7 justify-between items-center">
            <!-- Top Left Side -->
            <div class="flex items-center gap-5">
                <span>
                    <img
                        v-if="logo"
                        :src="logo"
                        alt="Company Logo"
                        class="rounded-full size-[6em] object-cover"
                    />
                    <img
                        v-else
                        src="/images/background.png"
                        alt="Company Logo"
                        class="rounded-full size-[6em] object-cover"
                    />
                </span>
                <span class="px-10 space-y-1">
                    <h1 class="text-3xl font-bold">{{ data.position }}</h1>
                    <span>
                        <NuxtLink to="/jobs/company">
                            <h2 class="text-[#15543A] text-xl font-semibold">
                                {{ data.name }}
                            </h2>
                        </NuxtLink>
                        <div class="flex space-x-2 mt-2">
                            <span
                                class="px-2 py-1 text-md bg-green-100 text-green-700 rounded-full"
                            >
                                {{ data.jobType }}
                            </span>
                            <span
                                class="px-2 py-1 text-md bg-green-100 text-green-700 rounded-full"
                            >
                                {{ data.experience }}
                            </span>
                        </div>
                    </span>
                </span>
            </div>
            <!-- Top Right Side -->
            <div>
                <span class="flex items-center justify-between">
                    <USwitch
                        v-model="isOpen"
                        size="xl"
                        :disabled="patchWaiting"
                        @change="handleToggle"
                        @update:model-value="
                            (value) => {
                                emit('update:open', value);
                            }
                        "
                    />
                    <span
                        v-if="isOpen"
                        class="text-primary-800 text-lg dark:text-primary font-bold"
                    >
                        Open
                    </span>
                    <span v-else class="text-error-800 text-lg dark:text-error font-bold"
                        >Closed</span
                    >
                </span>
                <!-- Edit Button -->
                <UModal v-model:open="openJobEditForm">
                    <UButton
                        class="px-4 py-2 border border-gray-400 rounded-md text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center mt-4 gap-2"
                        @click="openJobEditForm = true"
                    >
                        <Icon
                            name="material-symbols:edit-square-outline-rounded"
                            class="size-[1.5em]"
                        />
                        Edit Post
                    </UButton>
                    <template #content>
                        <JobEditForm :data="data" @close="handleCloseEditForm" />
                    </template>
                </UModal>
            </div>
        </div>

        <USeparator class="mt-2" />
        <!-- Information -->
        <div class="space-y-3">
            <h2 class="text-[#15543A] text-xl font-bold">About This Job</h2>
            <p>
                <span class="text-[#15543A] font-semibold">Position Title: </span>
                {{ data.position }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold capitalize">Location: </span>
                {{ data.location }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold">Duration: </span>
                {{ data.duration }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold">Description:</span>
                <br />
                {{ data.description }}
            </p>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";

const props = defineProps<{
    data: JobPost;
    open: boolean;
}>();

const openJobEditForm = ref(false);
const logo = ref("");
const data = ref<JobPost>(props.data);

const emit = defineEmits(["update:open", "close"]);

const isOpen = ref(false);
const patchWaiting = ref(false);

const config = useRuntimeConfig();
const api = useApi();
const toast = useToast();

onMounted(() => {
    isOpen.value = props.open;
    logo.value = `${config.public.apiBaseUrl}/files/${props.data.photoId}`;
});

interface PatchJobForm {
    name?: string;
    position?: string;
    duration?: string;
    description?: string;
    location?: string;
    jobType?: "fulltime" | "parttime" | "contract" | "casual" | "internship";
    experience?: "newgrad" | "junior" | "senior" | "manager" | "internship";
    minSalary?: number;
    maxSalary?: number;
    open?: boolean;
}

async function handleToggle() {
    // Set up for the patch request toggle switch
    patchWaiting.value = true;
    const token = localStorage.getItem("token");
    if (!token) {
        toast.add({
            title: "Not authenticated",
            description: "You must be logged in to perform this action.",
            color: "error",
        });
        patchWaiting.value = false;
        return;
    }

    const role = localStorage.getItem("role");
    if (role !== "company") {
        toast.add({
            title: "Unauthorized",
            description: "Only companies can update job status.",
            color: "error",
        });
        patchWaiting.value = false;
        return;
    }

    const toggleOpen: PatchJobForm = {
        open: isOpen.value,
    };
    try {
        await api.patch(`/jobs/${props.data.id}`, toggleOpen, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
        toast.add({
            title: "Job status updated",
            description: `The job is now ${isOpen.value ? "open" : "closed"}.`,
            color: "success",
        });
    } catch (error) {
        toast.add({
            title: "Error updating job status",
            description: "Failed to update job status. Please try again.",
            color: "error",
        });
        console.error("Error updating job status:", error);
        // Revert the toggle switch state
        isOpen.value = !isOpen.value;
    } finally {
        patchWaiting.value = false;
    }
}

function handleCloseEditForm() {
    openJobEditForm.value = false;
    emit("close");
}

// Watch for prop changes (parent update data after fetched job)
watch(
    () => props.data,
    (newData) => {
        data.value = newData;
    }
);
</script>
