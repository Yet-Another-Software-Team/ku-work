<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>
        <!-- Banner -->
        <div class="bg-gray-300 h-32 rounded-t-lg relative overflow-hidden">
            <img :src="data.profile.banner" alt="Banner" />
        </div>

        <!-- Top Section -->
        <div class="flex flex-wrap relative">
            <!-- Profile Image -->
            <div class="w-[12em] mr-5 -mt-20">
                <div v-if="data.profile.logo" class="w-40 h-40 flex-shrink-0">
                    <img
                        :src="data.profile.logo"
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
                    {{ data.profile.name }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ data.profile.address }}
                </p>
            </div>

            <!-- Edit Button -->
            <UButton
                v-if="isOwner"
                variant="outline"
                color="neutral"
                class="px-4 py-2 text-sm hover:cursor-pointer flex items-center mt-4 ml-auto mb-auto"
                @click="openEditModal = true"
            >
                <Icon name="material-symbols:edit-square-outline-rounded" class="size-[1.5em]" />
                Edit Profile
            </UButton>
        </div>

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />

        <!-- Bottom Section -->
        <div class="flex flex-wrap md:flex-nowrap text-xl">
            <!-- Connections -->
            <div class="w-[12rem] mr-5 mb-5">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">Connections</h3>
                <ul class="space-y-2 text-primary-600">
                    <li>
                        <a
                            :href="data.profile.website"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon
                                name="material-symbols:link-rounded"
                                class="size-[2em] text-black dark:text-white"
                            />
                            <span class="w-[10rem] text-sm truncate">{{
                                data.profile.website
                            }}</span>
                        </a>
                    </li>
                    <li>
                        <a
                            :href="email"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon
                                name="material-symbols:mail-outline"
                                class="size-[2em] text-black dark:text-white"
                            />
                            <span class="w-[10rem] text-sm truncate">{{ email }}</span>
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div class="flex-1">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About us</h3>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                    {{ data.profile.aboutUs }}
                </p>
            </div>
        </div>

        <UModal
            v-model:open="openEditModal"
            :ui="{
                container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
                overlay: 'fixed inset-0 bg-black/50',
                content: 'w-full max-w-6xl'
            }"
        >
            <template #content>
                <EditCompanyProfileCard
                    :profile="data.profile"
                    @close="openEditModal = false"
                    @saved="onSaved"
                />
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import EditCompanyProfileCard from '~/components/EditCompanyProfileCard.vue'
import { mockCompanyData } from "~/data/mockData";

withDefaults(
    defineProps<{
        isOwner?: boolean;
    }>(),
    {
        isOwner: true,
    }
);

const data = mockCompanyData;
const email = "john.doe@ku.th";

const openEditModal = ref(false)

function onSaved(updated: typeof data.profile) {
  Object.assign(data.profile, updated)
  openEditModal.value = false
}

// const api = useApi();

// onMounted(async () => {
//     try {
//         const response = await api.get("/company");
//         if (response.status === 200) {
//             console.log("Successfully fetched company profile:", response.data);
//         } else {
//             console.error("Failed to fetch company profile:", response.message);
//         }
//     } catch (error) {
//         console.error("Error fetching company profile:", error);
//     }
// });
</script>
