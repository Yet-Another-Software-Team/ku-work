<template>
    <div :id="'Component-' + requestId" class="flex justify-center items-center w-[20em] h-full">
        <!-- card container is clickable -->
        <div
            class="flex w-full h-[7em] m-2 shadow-md bg-[#fdfdfd] dark:bg-[#1f2937] rounded-md cursor-pointer"
            @click="() => navigateToProfile(requestId)"
        >
            <!-- profile pic -->
            <div
                v-if="data.photoId"
                class="flex items-center justify-center w-20 h-full flex-shrink-0"
            >
                <img
                    :src="photo"
                    alt="Profile photo"
                    class="w-16 h-16 object-cover rounded-full justify-center items-center m-2"
                />
            </div>
            <div v-else class="flex items-center justify-center w-20 h-full flex-shrink-0">
                <Icon name="ic:baseline-account-circle" class="size-full" />
            </div>

            <!-- user data -->
            <div class="flex flex-col justify-center flex-1 p-2">
                <!-- name -->
                <p class="overflow-hidden truncate max-w-[12em]">
                    {{ data.firstName }} {{ data.lastName }}
                </p>
                <p class="text-xs">{{ data.major }}</p>

                <!-- buttons -->
                <div class="flex gap-2 mt-2 w-full">
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-full flex-1 px-2"
                        variant="outline"
                        color="error"
                        label="Decline"
                        :icon="'iconoir:xmark'"
                        @click.stop="showRejectModal = true"
                    />
                    <UModal
                        v-model:open="showRejectModal"
                        :ui="{
                            title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                            container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
                            overlay: 'fixed inset-0 bg-black/50',
                        }"
                    >
                        <template #header>
                            <p>
                                Are you sure you want to <strong>decline</strong>
                                {{ data.firstName }} {{ data.lastName }}?
                            </p>
                        </template>
                        <template #body>
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
                                    :loading="actionLoading === 'decline'"
                                    @click="
                                        () => {
                                            actionLoading = 'decline';
                                            declineRequest();
                                            showRejectModal = false;
                                            actionLoading = null;
                                        }
                                    "
                                />
                            </div>
                        </template>
                    </UModal>

                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-full flex-1 px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        @click.stop="showAcceptModal = true"
                    />
                    <UModal
                        v-model:open="showAcceptModal"
                        :ui="{
                            title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                            container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
                            overlay: 'fixed inset-0 bg-black/50',
                        }"
                    >
                        <template #header>
                            <p>
                                Are you sure you want to <strong>accept</strong>
                                {{ data.firstName }} {{ data.lastName }}?
                            </p>
                        </template>
                        <template #body>
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
                                    :loading="actionLoading === 'accept'"
                                    @click="
                                        () => {
                                            actionLoading = 'accept';
                                            acceptRequest();
                                            showAcceptModal = false;
                                            actionLoading = null;
                                        }
                                    "
                                />
                            </div>
                        </template>
                    </UModal>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { ProfileInformation } from "~/data/datatypes";

const props = defineProps<{
    requestId: string;
    data?: ProfileInformation;
}>();

const emit = defineEmits<{
    (e: "studentApprovalStatus", id: string): void;
}>();

const data = props.data as ProfileInformation;
const api = useApi();

const toast = useToast();
const photo = ref<string>("");
const showAcceptModal = ref(false);
const showRejectModal = ref(false);
const actionLoading = ref<null | 'accept' | 'decline'>(null);

function navigateToProfile(id: string) {
    console.log("Navigating to profile of request:", props.requestId);
    navigateTo(`/admin/dashboard/profile/${id}`, { replace: true });
}

// Accept request
async function acceptRequest() {
    console.log("Accepted request:", props.requestId);
    toast.add({
        title: "Request accepted!",
        description: data.name,
        color: "success",
    });
    await api.post(
        `/students/${props.requestId}/approval`,
        { approve: true },
        { withCredentials: true }
    );
    emit("studentApprovalStatus", props.requestId);
}

// Decline request
async function declineRequest() {
    console.log("Declined request:", props.requestId);
    toast.add({
        title: "Request declined!",
        description: data.name,
        color: "error",
    });
    await api.post(
        `/students/${props.requestId}/approval`,
        { approve: false },
        { withCredentials: true }
    );
    emit("studentApprovalStatus", props.requestId);
}

watch(
    () => props.data,
    (newData) => {
        if (newData) {
            data.name = newData.name;
            data.major = newData.major;
            data.photoId = newData.photoId;
            data.phone = newData.phone;
            data.email = newData.email;
            photo.value = newData.photoId
                ? `${useRuntimeConfig().public.apiBaseUrl}/files/${newData.photoId}`
                : "";
        }
    },
    { immediate: true }
);
</script>
