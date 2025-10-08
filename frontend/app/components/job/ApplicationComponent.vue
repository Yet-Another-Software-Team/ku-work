<template>
    <div
        class="flex items-center shadow-md rounded-xl px-4 py-1 h-[10em] w-[25em] border border-gray-300 dark:border-gray-700 hover:shadow-lg transition-all gap-5"
    >
        <!-- Left Section -->
        <div class="flex">
            <!-- Avatar -->
            <div
                class="w-[6em] h-[6em] rounded-full border border-gray-700 flex items-center justify-center overflow-hidden"
            >
                <img
                    v-if="avatar"
                    :src="avatar"
                    alt="Profile photo"
                    class="object-cover w-full h-full"
                />
                <Icon v-else name="ic:baseline-account-circle" class="w-full h-full" />
            </div>
        </div>
        <!-- Right Section -->
        <div class="flex flex-col gap-y-2">
            <!-- Info -->
            <UBadge :color="colorPicker()" class="w-fit">{{ applicationData.status }}</UBadge>
            <h2 class="text-xl font-semibold">
                {{ applicationData.username }}
            </h2>
            <p class="text-sm">{{ profile.profile.major }}</p>
            <!-- Buttons -->
            <div class="flex items-center gap-2">
                <UButton
                    class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                    variant="outline"
                    color="error"
                    label="Decline"
                    :icon="'iconoir:xmark'"
                    @click.stop="emit('reject')"
                />
                <UButton
                    class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                    variant="outline"
                    color="primary"
                    label="Accept"
                    :icon="'iconoir:check'"
                    @click.stop="emit('approve')"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { mockUserData } from "~/data/mockData";
import type { JobApplication, Profile } from "~/data/mockData";

const props = defineProps<{
    applicationData: JobApplication;
}>();

// console.log("Application Data:", props.applicationData);

const api = useApi();
const config = useRuntimeConfig();
const profile = ref<Profile>(mockUserData);
const avatar = ref("");

onMounted(async () => {
    if (props.applicationData.userId) {
        try {
            const response = await api.get(`/students/${props.applicationData.userId}`);
            if (response && response.data && response.data.photoId) {
                avatar.value = `${config.public.apiBaseUrl}/files/${response.data.photoId}`;
            }
            profile.value = response.data;
            // console.log("Fetched profile data:", profile.value);
        } catch (error) {
            console.error("Error fetching user data:", error);
        }
    }
});

const emit = defineEmits<{
    (e: "approve" | "reject"): void;
}>();

function colorPicker() {
    if (props.applicationData.status === "pending") {
        return "warning";
    } else if (props.applicationData.status === "accepted") {
        return "primary";
    } else if (props.applicationData.status === "rejected") {
        return "error";
    } else {
        return "neutral";
    }
}
</script>
