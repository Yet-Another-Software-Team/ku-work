<template>
    <div class="rounded-lg">
        <!-- Header -->
        <a href="/admin/dashboard">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <Icon name="iconoir:nav-arrow-left" class="items-center" />
                <span>Back</span>
            </h1>
        </a>
        <section v-if="data && !isLoading" class="flex flex-col sm:flex-row">
            <!-- First Section -->
            <section class="w-full sm:w-[20em] flex flex-col gap-5 m-3">
                <!-- First Card -->
                <div class="r-card">
                    <!-- Profile Image -->
                    <div class="h-fit flex flex-col items-center">
                        <!-- photo -->
                        <div v-if="photo" class="size-[16em] flex-shrink-0">
                            <img
                                :src="photo"
                                alt="Profile photo"
                                class="size-full object-cover rounded-full p-5"
                            />
                        </div>

                        <!-- icon -->
                        <div
                            v-else
                            class="flex items-center justify-center w-[16em] h-[16em] flex-shrink-0"
                        >
                            <Icon
                                name="ic:baseline-account-circle"
                                class="size-full text-gray-400"
                            />
                        </div>

                        <!-- Info -->
                        <div class="text-xl mt-5">
                            <h2 class="text-2xl font-semibold text-gray-900 dark:text-[#FDFDFD]">
                                {{ data.profile.name }}
                            </h2>
                            <p class="text-gray-600 dark:text-gray-300">
                                {{ data.profile.major }}
                            </p>
                        </div>

                        <!-- Divider -->
                        <hr class="border-gray-500 dark:border-gray-600 m-5 w-full" />

                        <!-- Connections -->
                        <ul>
                            <li>
                                <a
                                    :href="data.profile.linkedIn"
                                    target="_blank"
                                    class="flex hover:underline items-center justify-center gap-2 my-5"
                                >
                                    <Icon name="devicon:linkedin" class="size-[2em]" />
                                    <span class="truncate w-[12em]">{{ data.profile.name }}</span>
                                </a>
                            </li>
                            <li>
                                <a
                                    :href="data.profile.github"
                                    target="_blank"
                                    class="flex hover:underline items-center justify-center gap-2 my-5"
                                >
                                    <Icon
                                        name="devicon:github"
                                        class="size-[2em] bg-[#FDFDFD] rounded-full"
                                    />
                                    <span class="truncate w-[12em]">{{ data.profile.name }}</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                </div>

                <!-- Second Card -->
                <div class="r-card flex justify-between">
                    <!-- File Download -->
                    <span class="text-xl">Submitted File</span>
                    <a :href="file" download>
                        <Icon name="ic:outline-file-download" class="size-[2em]" />
                    </a>
                </div>
            </section>

            <!-- Second Section -->
            <section class="flex flex-col gap-5 m-3 w-full">
                <!-- First Card -->
                <div class="r-card">
                    <div class="flex flex-col text-xl text-left whitespace-normal">
                        <p class="mt-2 text-gray-800 dark:text-gray-200 font-semibold">
                            Age: <span class="font-normal">{{ age }}</span>
                        </p>
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            Phone: <span class="font-normal">{{ data.profile.phone }}</span>
                        </p>
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            Email: <span class="font-normal">{{ email }}</span>
                        </p>
                    </div>
                </div>

                <!-- Second Card -->
                <div class="r-card">
                    <div class="flex flex-col text-xl text-left">
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            <span v-if="data.profile.aboutMe.trim() !== ''" class="font-normal">{{
                                data.profile.aboutMe
                            }}</span>
                            <span v-else class="font-normal">No description provided</span>
                        </p>
                    </div>
                </div>

                <!-- Decision Button -->
                <div class="flex gap-5">
                    <span class="c-ubutton w-full sm:w-1/2">
                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl"
                            variant="outline"
                            color="error"
                            label="Decline"
                            :icon="'iconoir:xmark'"
                            @click.stop="declineRequest"
                        />
                    </span>
                    <span class="c-ubutton w-full sm:w-1/2">
                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl"
                            variant="outline"
                            color="primary"
                            label="Accept"
                            :icon="'iconoir:check'"
                            @click.stop="acceptRequest"
                        />
                    </span>
                </div>
            </section>
        </section>

        <!-- Loading State (USkeleton) -->
        <section v-else class="flex justify-center">
            <!-- First Section -->
            <section class="w-full sm:w-[20em] flex flex-col gap-5 m-3">
                <!-- First Card -->
                <div class="r-card">
                    <div class="h-fit flex flex-col items-center">
                        <!-- Profile Image Skeleton -->
                        <USkeleton class="size-[16em] rounded-full p-5" />

                        <!-- Info Skeleton -->
                        <div class="text-xl mt-5 flex flex-col items-center gap-2">
                            <USkeleton class="h-6 w-3/4 rounded-md" />
                            <USkeleton class="h-5 w-1/2 rounded-md" />
                        </div>

                        <!-- Divider -->
                        <hr class="border-gray-500 dark:border-gray-600 m-5 w-full" />

                        <!-- Connections Skeleton -->
                        <ul class="w-full flex flex-col items-center">
                            <li class="flex items-center justify-center gap-2 my-5 w-full px-10">
                                <USkeleton class="size-[2em] rounded-full" />
                                <USkeleton class="h-5 w-[12em] rounded-md" />
                            </li>
                            <li class="flex items-center justify-center gap-2 my-5 w-full px-10">
                                <USkeleton class="size-[2em] rounded-full" />
                                <USkeleton class="h-5 w-[12em] rounded-md" />
                            </li>
                        </ul>
                    </div>
                </div>

                <!-- Second Card -->
                <div class="r-card flex justify-between items-center">
                    <USkeleton class="h-6 w-1/2 rounded-md" />
                    <USkeleton class="size-[2em] rounded-md" />
                </div>
            </section>

            <!-- Second Section -->
            <section class="flex flex-col gap-5 m-3 w-full">
                <!-- First Card -->
                <div class="r-card flex flex-col text-xl text-left gap-3">
                    <USkeleton class="h-5 w-1/3 rounded-md" />
                    <USkeleton class="h-5 w-2/3 rounded-md" />
                    <USkeleton class="h-5 w-1/2 rounded-md" />
                </div>

                <!-- Second Card -->
                <div class="r-card">
                    <USkeleton class="h-20 w-full rounded-md" />
                </div>

                <!-- Decision Buttons -->
                <div class="flex gap-5">
                    <USkeleton class="h-12 w-full sm:w-1/2 rounded-md" />
                    <USkeleton class="h-12 w-full sm:w-1/2 rounded-md" />
                </div>
            </section>
        </section>

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Profile } from "~/data/mockData";
import { mockUserData } from "~/data/mockData";

const props = defineProps<{
    requestId: string;
}>();

const data = ref<Profile>(mockUserData);
const toast = useToast();

const api = useApi();
const config = useRuntimeConfig();
const isLoading = ref(true);

const photo = ref("");
const file = ref("");
const email = computed(() => {
    if (!data.value) return "";
    return data.value.profile.email || "No email provided";
});

onMounted(async () => {
    isLoading.value = true;
    try {
        const response = await api.get(`/students`, {
            params: { id: props.requestId, approvalStatus: "pending" },
            withCredentials: true,
        });
        data.value = response.data as Profile;
        data.value.profile.name = `${data.value.profile.firstName} ${data.value.profile.lastName}`;
        photo.value = `${config.public.apiBaseUrl}/files/${data.value.profile.photoId}`;
        file.value = `${config.public.apiBaseUrl}/files/${data.value.profile.statusFileId}`;
        console.log("Fetched profile data:", data.value);
    } catch (error) {
        console.error("Error fetching profile data:", error);
        toast.add({
            title: "Error fetching profile data",
            description: String(error),
            color: "error",
        });
    } finally {
        console.log("Loading finished");
        isLoading.value = false;
    }
});

// Compute age
const age = computed(() => {
    const birth = new Date(data.value?.profile.birthDate ?? "");
    const today = new Date();
    let years = today.getFullYear() - birth.getFullYear();
    const m = today.getMonth() - birth.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) {
        years--;
    }
    return years;
});

// Accept request
async function acceptRequest() {
    console.log("Accepted request:", props.requestId);
    toast.add({
        title: "Request accepted!",
        description: data.value.profile.name,
        color: "success",
    });
    await api.post(
        `/students/${props.requestId}/approval`,
        { approve: true },
        { withCredentials: true }
    );
    navigateTo("/admin/dashboard", { replace: true });
}

// Decline request
async function declineRequest() {
    console.log("Declined request:", props.requestId);
    toast.add({
        title: "Request declined!",
        description: data.value.profile.name,
        color: "error",
    });
    await api.post(
        `/students/${props.requestId}/approval`,
        { approve: false },
        { withCredentials: true }
    );
    navigateTo("/admin/dashboard", { replace: true });
}

watch(
    () => props.requestId,
    async (newId) => {
        if (newId) {
            isLoading.value = true;
            try {
                const response = await api.get(`/students`, {
                    params: { id: newId, approvalStatus: "pending" },
                    withCredentials: true,
                });
                data.value = response.data as Profile;
                data.value.profile.name = `${data.value.profile.firstName} ${data.value.profile.lastName}`;
                photo.value = `${config.public.apiBaseUrl}/files/${data.value.profile.photoId}`;
                file.value = `${config.public.apiBaseUrl}/files/${data.value.profile.statusFileId}`;
                console.log("Fetched profile data:", data.value);
            } catch (error) {
                console.error("Error fetching profile data:", error);
                toast.add({
                    title: "Error fetching profile data",
                    description: String(error),
                    color: "error",
                });
            } finally {
                console.log("Loading finished");
                isLoading.value = false;
            }
        }
    },
    { immediate: true }
);
</script>

<style>
.r-card {
    box-shadow:
        0 4px 6px -1px rgba(0, 0, 0, 0.2),
        0 2px 4px -2px rgba(0, 0, 0, 0.2);
    text-align: center;
    padding: 1.25rem;
    border-radius: 0.5rem;
    background-color: #fdfdfd;
}

.dark .r-card {
    background-color: #1f2937;
}

.c-ubutton {
    background-color: #fdfdfd;
}

.dark .c-ubutton {
    background-color: #1f2937;
}
</style>
