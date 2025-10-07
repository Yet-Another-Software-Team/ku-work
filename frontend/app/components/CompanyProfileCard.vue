<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>
        <!-- Banner -->
        <div class="bg-gray-300 h-32 rounded-t-lg relative overflow-hidden">
            <img :src="profile.banner" alt="Banner" />
        </div>

        <!-- Top Section -->
        <div class="flex flex-wrap relative">
            <!-- Profile Image -->
            <div class="w-[12em] mr-5 -mt-20">
                <div v-if="profile.photo" class="w-40 h-40 flex-shrink-0">
                    <img
                        :src="profile.photo"
                        alt="Profile photo"
                        class="w-40 h-40 object-cover rounded-full justify-center items-center m-2"
                    />
                </div>
                <div v-else class="flex items-center justify-center w-40 h-40 flex-shrink-0">
                    <Icon name="ic:baseline-account-circle" class="size-full" />
                </div>
            </div>

            <!-- Info -->
            <div class="text-xl">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ profile.name }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ profile.address }}
                </p>
            </div>

            <!-- Edit Button -->
            <button
                v-if="isOwner"
                class="px-4 py-2 border border-gray-400 rounded-md text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center mt-4 ml-auto mb-auto cursor-pointer"
            >
                <Icon name="material-symbols:edit-square-outline-rounded" class="size-[1.5em]" />
                Edit Profile
            </button>
        </div>

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />

        <!-- Bottom Section -->
        <div class="flex flex-wrap md:flex-nowrap text-xl">
            <!-- Connections -->
            <div class="w-[12rem] mr-5 mb-5">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">Connections</h3>
                <ul class="space-y-2 text-primary-600">
                    <li v-if="profile.website">
                        <a
                            :href="profile.website"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon
                                name="material-symbols:link-rounded"
                                class="size-[2em] text-black dark:text-white"
                            />
                            <span class="w-[10rem] text-sm truncate">{{ profile.website }}</span>
                        </a>
                    </li>
                    <li>
                        <a
                            :href="profile.email"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon
                                name="material-symbols:mail-outline"
                                class="size-[2em] text-black dark:text-white"
                            />
                            <span class="w-[10rem] text-sm truncate">{{ profile.email }}</span>
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div class="flex-1">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About us</h3>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                    {{ profile.about }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
// import { mockCompanyData } from "~/data/mockData";

withDefaults(
    defineProps<{
        isOwner?: boolean;
    }>(),
    {
        isOwner: true,
    }
);

const profile = ref({
    photo: "",
    banner: "",
    email: "",
    website: "",
    about: "",
    address: "",
    name: "",
});

const config = useRuntimeConfig();
const api = useApi();

onMounted(async () => {
    try {
        // Get user ID from localStorage
        const userId = localStorage.getItem("userId");
        if (!userId) {
            console.error("No user ID found in localStorage");
            return;
        }

        // Fetch full company profile using user ID
        const response = await api.get(`/company/${userId}`);
        if (response.status === 200) {
            console.log("Successfully fetched company profile:", response.data);
            response.data.banner = `${config.public.apiBaseUrl}/files/${response.data.bannerId}`;
            response.data.photo = `${config.public.apiBaseUrl}/files/${response.data.photoId}`;
            profile.value = response.data;
        } else {
            console.error("Failed to fetch company profile:", response.message);
        }
    } catch (error) {
        console.error("Error fetching company profile:", error);
    }
});
</script>
