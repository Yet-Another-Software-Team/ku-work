<template>
    <div
        class="flex bg-white dark:bg-[#1f2937] items-center shadow-md rounded-xl px-4 py-1 h-[10em] w-[25em] border border-gray-300 dark:border-gray-700 hover:shadow-lg transition-all gap-5"
    >
        <!-- Left Section -->
        <div class="flex">
            <!-- Avatar -->
            <div
                class="w-[6em] h-[6em] rounded-full border border-gray-700 flex items-center justify-center overflow-hidden"
            >
                <div
                    v-if="isLoadingProfile"
                    class="w-full h-full bg-gray-200 dark:bg-gray-600 animate-pulse rounded-full"
                ></div>
                <img
                    v-else-if="avatar"
                    :src="avatar"
                    alt="Profile photo"
                    class="object-cover w-full h-full"
                    @error="avatar = ''"
                />
                <Icon v-else name="ic:baseline-account-circle" class="w-full h-full" />
            </div>
        </div>
        <!-- Right Section -->
        <div class="flex flex-col gap-y-2">
            <!-- Info -->
            <UBadge :color="colorPicker()" class="w-fit">{{ applicationData.status }}</UBadge>
            <NuxtLink
                :to="{
                    name: 'dashboard-id-beforeEmail',
                    params: {
                        id: applicationData.jobId,
                        beforeEmail: beforeEmail,
                    },
                }"
            >
                <h2 class="text-xl font-semibold">
                    {{ applicationData.username }}
                </h2>
            </NuxtLink>
            <p class="text-sm">{{ major }}</p>
            <!-- Buttons -->
            <div class="flex items-center gap-2">
                <UButton
                    class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2 cursor-pointer"
                    variant="outline"
                    color="error"
                    label="Decline"
                    :icon="'iconoir:xmark'"
                    @click.stop="emit('reject')"
                />
                <UButton
                    class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2 cursor-pointer"
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
import type { JobApplication, Profile, ProfileInformation } from "~/data/datatypes";
import { computed } from "vue";

const props = defineProps<{
    applicationData: JobApplication;
}>();

const api = useApi();
const config = useRuntimeConfig();

// Normalize and store profile in a consistent shape: { profile: ProfileInformation }
const profile = ref<Profile | undefined>(undefined);
const avatar = ref("");
const isLoadingProfile = ref(false);

// Computed helpers to avoid template runtime errors when data is missing
const major = computed(() => {
    return profile.value?.profile?.major ?? "";
});

const beforeEmail = computed(() => {
    const email = props.applicationData?.email ?? "";
    return email.split("@")[0];
});

const fetchUserProfile = async (userId: string) => {
    if (!userId) return;

    console.log(`[ApplicationComponent] Fetching profile for userId: ${userId}`);
    isLoadingProfile.value = true;

    try {
        const response = await api.get(`/students`, {
            params: { id: userId },
        });
        const data = response && response.data ? response.data : null;
        if (data) {
            // API sometimes returns the profile directly or wrapped as { profile: ... }
            const student: ProfileInformation = data.profile ?? data;
            if (student?.photoId) {
                // Add cache-busting parameter to prevent browser image caching
                const timestamp = Date.now();
                const newAvatarUrl = `${config.public.apiBaseUrl}/files/${student.photoId}?t=${timestamp}`;
                console.log(`[ApplicationComponent] Setting avatar for ${userId}: ${newAvatarUrl}`);
                avatar.value = newAvatarUrl;
            } else {
                console.log(`[ApplicationComponent] No photoId for ${userId}`);
                avatar.value = "";
            }
            // Ensure callers can always access profile.profile.*
            profile.value = { profile: student };
        }
    } catch (error) {
        console.error("Error fetching user data:", error);
        avatar.value = "";
        profile.value = undefined;
    } finally {
        isLoadingProfile.value = false;
    }
};

// Watch for changes in userId and refetch profile
watch(
    () => props.applicationData.userId,
    (newUserId, oldUserId) => {
        if (newUserId !== oldUserId) {
            console.log(`[ApplicationComponent] UserId changed from ${oldUserId} to ${newUserId}`);
            // Reset avatar immediately to prevent showing wrong image
            avatar.value = "";
            profile.value = undefined;
            isLoadingProfile.value = true;
            fetchUserProfile(newUserId);
        }
    },
    { immediate: true }
);

// Watch for changes in the entire applicationData object
watch(
    () => props.applicationData,
    (newApp, oldApp) => {
        // Force refresh if the application object changes significantly
        if (
            newApp &&
            oldApp &&
            (newApp.id !== oldApp.id ||
                newApp.userId !== oldApp.userId ||
                newApp.status !== oldApp.status)
        ) {
            console.log(`[ApplicationComponent] Application data changed:`, {
                oldId: oldApp.id,
                newId: newApp.id,
                oldUserId: oldApp.userId,
                newUserId: newApp.userId,
                oldStatus: oldApp.status,
                newStatus: newApp.status,
            });
            avatar.value = "";
            profile.value = undefined;
            isLoadingProfile.value = true;
            fetchUserProfile(newApp.userId);
        }
    },
    { deep: true }
);

onMounted(async () => {
    await fetchUserProfile(props.applicationData.userId);
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
