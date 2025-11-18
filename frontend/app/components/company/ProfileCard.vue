<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>
        <!-- Banner -->
        <div class="bg-gray-300 h-[10rem] rounded-t-lg relative overflow-hidden">
            <img :src="profile.banner" alt="Banner" class="object-cover size-full" />
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
            <div v-if="isActive" class="text-xl">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ profile.name }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ profile.address }} {{ profile.city }} {{ profile.country }}
                </p>
            </div>

            <!-- User Options -->
            <UDropdownMenu
                v-if="canEdit && isActive"
                :items="items"
                :content="{ align: 'end' }"
                class="p-1 text-sm hover:cursor-pointer flex items-center mt-4 ml-auto mb-auto"
            >
                <UButton color="neutral" variant="ghost" icon="ic:baseline-more-vert" />
            </UDropdownMenu>

            <div v-else-if="!isActive" class="text-xl">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
                    Profile Inactive
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    Your profile is currently deactivated. Please reactivate your profile to access
                    all features.
                </p>
            </div>
        </div>

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />

        <!-- Bottom Section -->
        <div v-if="isActive" class="flex flex-wrap md:flex-nowrap text-xl overflow-x-hidden">
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
                            :href="`mailto:${profile.email}`"
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
            <div class="flex-1 overflow-x-hidden">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About us</h3>
                <p
                    class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed whitespace-pre-wrap break-all overflow-x-hidden max-w-full"
                >
                    {{ profile.about || "None" }}
                </p>
            </div>
        </div>
        <!-- Reactivate button -->
        <div v-else>
            <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
                Reactivate Profile
            </h2>
            <UButton
                variant="subtle"
                color="primary"
                class="h-10 self-center font-semibold"
                @click="openReactivateModal = true"
            >
                Reactivate
            </UButton>
        </div>

        <!-- Edit Modal -->
        <UModal
            v-model:open="openEditModal"
            :ui="{
                overlay: 'fixed inset-0 bg-black/50',
                content: 'w-full max-w-6xl',
            }"
        >
            <template #content>
                <EditCompanyProfileCard
                    :profile="profile"
                    :saving="isSaving"
                    @close="openEditModal = false"
                    @saved="onSaved"
                />
            </template>
        </UModal>
        <!-- Deactivate Modal -->
        <DeactivateModal
            v-model:open="openDeactivateModal"
            @update:close="(value) => (openDeactivateModal = value)"
        />
        <!-- Reactivate Modal -->
        <ReactivateModal
            v-model:open="openReactivateModal"
            @update:close="(value) => (openReactivateModal = value)"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { DropdownMenuItem } from "@nuxt/ui";
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
const isSaving = ref(false);
const openDeactivateModal = ref(false);
const openReactivateModal = ref(false);
const isActive = ref(true);

const api = useApi();
const config = useRuntimeConfig();
const toast = useToast();

// Allow editing when owner or when viewing own company ID
const canEdit = computed(() => {
    const uid = import.meta.client ? localStorage.getItem("userId") : null;
    return props.isOwner || (!!props.companyId && props.companyId === uid);
});

// Fetch company profile
async function fetchCompanyProfile() {
    try {
        let idToFetch: string | null = props.companyId;
        if (props.isOwner && !idToFetch) {
            idToFetch = localStorage.getItem("userId");
        }
        if (!idToFetch) {
            console.error("No company ID provided or found");
            return;
        }
        const response = await api.get(`/company/${idToFetch}`);
        if (response.status === 200) {
            response.data.banner = `${config.public.apiBaseUrl}/files/${response.data.bannerId}`;
            response.data.photo = `${config.public.apiBaseUrl}/files/${response.data.photoId}`;
            profile.value = response.data;
        } else {
            console.error("Failed to fetch company profile:", response.message);
        }
    } catch (error) {
        console.error("Error fetching company profile:", error);
        isActive.value = false;
    }
}

// Dropdown menu items
const items: DropdownMenuItem[] = [
    {
        label: "Edit Profile",
        icon: "material-symbols:edit-square-outline-rounded",
        onClick: () => {
            openEditModal.value = true;
        },
    },
    {
        label: "Deactivate Profile",
        icon: "material-symbols:delete-outline",
        onClick: () => {
            openDeactivateModal.value = true;
        },
    },
];

// Handle Saves, Deactivation, and Deletion
async function onSaved(
    updated: {
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
    },
    cfToken: string
) {
    isSaving.value = true;
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
    // wait for backend confirmation before updating local state
    try {
        await api.patch("/me", formData, {
            headers: {
                "Content-Type": "multipart/form-data",
                "X-Turnstile-Token": cfToken,
            },
        });
        await fetchCompanyProfile();
        toast.add({
            title: "Saved",
            description: "Company profile updated successfully.",
            color: "success",
        });
        openEditModal.value = false;
    } catch (error) {
        console.log(error);
        toast.add({
            title: "Failed to update profile",
            description: (error as { message: string }).message,
            color: "error",
        });
    } finally {
        isSaving.value = false;
    }
}

onMounted(async () => {
    await fetchCompanyProfile();
});
</script>
