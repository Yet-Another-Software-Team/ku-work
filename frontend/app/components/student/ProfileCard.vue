<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>

        <!-- Top Section -->
        <div class="flex flex-wrap">
            <!-- Profile Image -->
            <div class="w-[12em] mr-5">
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
                    {{ `${profile.firstName} ${profile.lastName}` }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ profile.major }}
                </p>

                <p v-if="hasBirthDate" class="mt-2 text-gray-800 dark:text-gray-200 font-semibold">
                    Age:
                    <span class="font-normal">{{ age }}</span>
                </p>
                <p v-if="profile.phone" class="text-gray-800 dark:text-gray-200 font-semibold">
                    Phone:
                    <span class="font-normal">{{ profile.phone }}</span>
                </p>
                <p class="text-gray-800 dark:text-gray-200 font-semibold">
                    Email: <span class="font-normal">{{ profile.email }}</span>
                </p>
            </div>

            <!-- User Options -->
            <UDropdownMenu
                v-if="isActive"
                :items="items"
                :content="{ align: 'end' }"
                class="p-1 text-sm hover:cursor-pointer flex items-center mt-4 ml-auto mb-auto"
            >
                <UButton color="neutral" variant="ghost" icon="ic:baseline-more-vert" />
            </UDropdownMenu>

            <div v-else class="text-xl">
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
        <div v-if="isActive" class="flex flex-wrap md:flex-nowrap text-xl">
            <!-- Connections -->
            <div v-if="hasAnyConnection" class="w-[12rem] mr-5 mb-5">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">Connections</h3>
                <ul class="space-y-2 text-primary-600">
                    <li v-if="hasLinkedIn">
                        <a
                            :href="profile.linkedIn"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:linkedin" class="size-[2em]" />
                            <span class="w-[10em] truncate">{{ linkedInLabel }}</span>
                        </a>
                    </li>
                    <li v-if="hasGitHub">
                        <a
                            :href="profile.github"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:github" class="size-[2em] bg-white rounded-full" />
                            <span class="w-[10em] truncate">{{ githubLabel }}</span>
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div class="flex-1">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About me</h3>
                <p
                    class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed whitespace-pre-wrap break-words overflow-x-hidden"
                >
                    {{ aboutMeDisplay }}
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

        <!-- Modals -->
        <!-- Edit Modal -->
        <UModal
            v-model:open="openEditModal"
            :ui="{
                overlay: 'fixed inset-0 bg-black/50',
                content: 'w-full max-w-2xl',
            }"
        >
            <template #content>
                <EditStudentProfileCard
                    :profile="profile"
                    :saving="isSaving"
                    @close="isEditModalOpen = false"
                    @saved="handleProfileSaved"
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
import type { DropdownMenuItem } from "@nuxt/ui";

const toast = useToast();

const isEditModalOpen = ref(false);
const isSaving = ref(false);
const openDeactivateModal = ref(false);
const openReactivateModal = ref(false);
const openEditModal = ref(false);
const isActive = ref(true);

interface StudentProfile {
    name?: string;
    birthDate?: string;
    phone?: string;
    github?: string;
    linkedIn?: string;
    aboutMe?: string;
    photo?: string;
}

type StudentProfileUpdate = StudentProfile & { _avatarFile?: File | null };

const handleProfileSaved = async (updated: StudentProfileUpdate, cfToken: string) => {
    const { _avatarFile, ...newProfile } = updated;
    openEditModal.value = false;
    const formData = new FormData();
    if (profile.value.phone !== updated.phone) formData.append("phone", updated.phone!);
    if (profile.value.birthDate !== updated.birthDate)
        formData.append("birthDate", updated.birthDate! + "T00:00:00.000Z");
    if (profile.value.aboutMe !== updated.aboutMe) formData.append("aboutMe", updated.aboutMe!);
    if (profile.value.github !== updated.github) formData.append("github", updated.github!);
    if (profile.value.linkedIn !== updated.linkedIn) formData.append("linkedIn", updated.linkedIn!);
    if (_avatarFile) formData.append("photo", _avatarFile!);
    formData.append("studentStatus", profile.value.status);
    Object.assign(profile.value, newProfile);
    try {
        await api.patch("/me", formData, {
            headers: {
                "Content-Type": "multipart/form-data",
                "X-Turnstile-Token": cfToken,
            },
        });
        await fetchStudentProfile();
        toast.add({
            title: "Saved",
            description: "Profile updated successfully.",
            color: "success",
        });
        isEditModalOpen.value = false;
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
};

const profile = ref({
    photo: "",
    birthDate: "0",
    phone: "",
    major: "",
    linkedIn: "",
    github: "",
    aboutMe: "",
    firstName: "",
    lastName: "",
    email: "",
    status: "",
    approvalStatus: "",
});

const config = useRuntimeConfig();
const api = useApi();

async function fetchStudentProfile() {
    try {
        const response = await api.get("/students");
        if (response.status === 200) {
            response.data.profile.photo = `${config.public.apiBaseUrl}/files/${response.data.profile.photoId}`;
            profile.value = response.data.profile;
        } else {
            console.error("Failed to fetch student profile:", response.message);
        }
    } catch (error) {
        console.error("Error fetching student profile:", error);
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

onMounted(async () => {
    await fetchStudentProfile();
});

// Social connections helpers
const hasLinkedIn = computed(
    () => !!profile.value.linkedIn && profile.value.linkedIn.trim() !== ""
);
const hasGitHub = computed(() => !!profile.value.github && profile.value.github.trim() !== "");
const hasAnyConnection = computed(() => hasLinkedIn.value || hasGitHub.value);

function extractLabel(url?: string) {
    if (!url) return "";
    try {
        const u = new URL(url);
        const parts = u.pathname.split("/").filter(Boolean);
        return parts[parts.length - 1] || u.hostname;
    } catch {
        return url;
    }
}
const linkedInLabel = computed(() => extractLabel(profile.value.linkedIn));
const githubLabel = computed(() => extractLabel(profile.value.github));

// Birthdate helpers
const hasBirthDate = computed(() => {
    const bd = profile.value.birthDate;
    if (!bd) return false;
    const d = new Date(bd);
    return !isNaN(d.getTime());
});
const age = computed(() => {
    if (!hasBirthDate.value) return "";
    const birth = new Date(profile.value.birthDate);
    const today = new Date();
    let years = today.getFullYear() - birth.getFullYear();
    const m = today.getMonth() - birth.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) {
        years--;
    }
    return years;
});

// About me fallback
const aboutMeDisplay = computed(() => {
    const v = profile.value.aboutMe || "";
    return v.trim() === "" ? "None" : v;
});
</script>
