<template>
    <div :id="'Component-' + requestId" class="flex justify-center items-center w-[20em] h-full">
        <!-- card container is clickable -->
        <div
            class="flex w-full h-[7em] m-2 shadow-md/20 bg-[#fdfdfd] dark:bg-[#1f2937] rounded-md cursor-pointer"
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
                    class="w-17 h-17 object-cover rounded-full justify-center items-center m-2"
                />
            </div>
            <div v-else class="flex items-center justify-center w-20 h-full flex-shrink-0">
                <Icon name="ic:baseline-account-circle" class="size-full" />
            </div>

            <!-- user data -->
            <div class="flex flex-col justify-center flex-1 p-2">
                <!-- name -->
                <p class="overflow-hidden truncate max-w-[12em]">{{ data.name }}</p>
                <p class="text-xs">{{ data.major }}</p>

                <!-- buttons -->
                <div class="flex gap-2 mt-2">
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="error"
                        label="Decline"
                        :icon="'iconoir:xmark'"
                        @click.stop="declineRequest"
                    />
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        @click.stop="acceptRequest"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { ProfileInformation } from "~/data/mockData";

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
