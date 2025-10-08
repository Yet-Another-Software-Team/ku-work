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
                    {{ profile.address }} {{ profile.city }} {{ profile.country }}
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
                    <li v-if="profile.phone">
                        <span class="flex items-center gap-2">
                            <Icon
                                name="material-symbols:call-outline"
                                class="size-[2em] text-black dark:text-white"
                            />
                            <span class="w-[10rem] text-sm truncate">{{ profile.phone }}</span>
                        </span>
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

        <UModal
            v-model:open="openEditModal"
            :ui="{
                container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
                overlay: 'fixed inset-0 bg-black/50',
                content: 'w-full max-w-6xl',
            }"
        >
            <template #content>
                <EditCompanyProfileCard
                    :profile="profile"
                    @close="openEditModal = false"
                    @saved="onSaved"
                />
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import EditCompanyProfileCard from "~/components/EditCompanyProfileCard.vue";

const props = withDefaults(
    defineProps<{
        isOwner?: boolean;
        companyId?: string;
    }>(),
    {
        isOwner: true,
        companyId: "",
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
    city: "",
    country: "",
    phone: "",
});

const openEditModal = ref(false);
const api = useApi();
const config = useRuntimeConfig();
const toast = useToast();

async function onSaved(updated: {
    name?: string;
    address?: string;
    website?: string;
    banner?: string;
    photo?: string;
    about?: string;
    city?: string;
    country?: string;
    email?: string;
    phone?: string;
    _logoFile?: File | null;
    _bannerFile?: File | null;
}) {
    openEditModal.value = false;
    const formData = new FormData();
    if (profile.value.name !== updated.name) formData.append("username", updated.name!);
    if (profile.value.address !== updated.address) formData.append("address", updated.address!);
    if (profile.value.website !== updated.website) formData.append("website", updated.website!);
    if (profile.value.city !== updated.city) formData.append("city", updated.city!);
    if (profile.value.about !== updated.about) formData.append("about", updated.about!);
    if (profile.value.country !== updated.country) formData.append("country", updated.country!);
    if (profile.value.email !== updated.email) formData.append("email", updated.email!);
    if (profile.value.phone !== updated.phone) formData.append("phone", updated.phone!);
    if (updated._logoFile) formData.append("photo", updated._logoFile!);
    if (updated._bannerFile) formData.append("banner", updated._bannerFile!);
    Object.assign(profile, updated);
    try {
        await api.patch("/me", formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        });
    } catch (error) {
        console.log(error);
        toast.add({
            title: "Failed to update profile",
            description: (error as { message: string }).message,
            color: "error",
        });
    }
}

onMounted(async () => {
    try {
        let idToFetch: string | null = props.companyId;
        if (props.isOwner && !idToFetch) {
            // Get user ID from localStorage for owner view
            idToFetch = localStorage.getItem("userId");
        }

        if (!idToFetch) {
            console.error("No company ID provided or found");
            return;
        }

        // Fetch full company profile using company ID
        const response = await api.get(`/company/${idToFetch}`);
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
