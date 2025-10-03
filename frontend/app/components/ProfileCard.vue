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
            <div class="text-xl">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ `${profile.firstName} ${profile.lastName}` }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ profile.major }}
                </p>

                <p class="mt-2 text-gray-800 dark:text-gray-200 font-semibold">
                    Age:
                    <span class="font-normal">{{
                        Math.abs(
                            new Date(Date.now() - Date.parse(profile.birthDate)).getUTCFullYear() -
                                1970
                        )
                    }}</span>
                </p>
                <p v-if="profile.phone" class="text-gray-800 dark:text-gray-200 font-semibold">
                    Phone:
                    <span class="font-normal">{{ profile.phone }}</span>
                </p>
                <p class="text-gray-800 dark:text-gray-200 font-semibold">
                    Email: <span class="font-normal">{{ profile.email }}</span>
                </p>
            </div>

            <!-- Edit Button -->
            <button
                class="px-4 py-2 border border-gray-400 rounded-md text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center mt-4 sm:mt-0 sm:ml-10 mr-auto mb-auto cursor-pointer"
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
                    <li v-if="profile.linkedIn">
                        <a
                            :href="profile.linkedIn"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:linkedin" class="size-[2em]" />
                            <span class="w-[10em] truncate">{{
                                `${profile.firstName} ${profile.lastName}`
                            }}</span>
                        </a>
                    </li>
                    <li v-if="profile.github">
                        <a
                            :href="profile.github"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:github" class="size-[2em] bg-white rounded-full" />
                            <span class="w-[10em] truncate">{{
                                `${profile.firstName} ${profile.lastName}`
                            }}</span>
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div class="flex-1">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About me</h3>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                    {{ profile.aboutMe }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
// import { computed } from "vue";
// import { mockUserData } from "~/data/mockData";

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
});
// const data = mockUserData;

// // Compute age
// const age = computed(() => {
//     const birth = new Date(data.profile.birthDate);
//     const today = new Date();
//     let years = today.getFullYear() - birth.getFullYear();
//     const m = today.getMonth() - birth.getMonth();
//     if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) {
//         years--;
//     }
//     return years;
// });

// const email = "john.doe@ku.th";
const config = useRuntimeConfig();
const api = useApi();

onMounted(async () => {
    try {
        const response = await api.get("/students");
        if (response.status === 200) {
            console.log("Successfully fetched student profile:", response.data);
            response.data.profile.photo = `${config.public.apiBaseUrl}/files/${response.data.profile.photoId}`;
            profile.value = response.data.profile;
        } else {
            console.error("Failed to fetch student profile:", response.message);
        }
    } catch (error) {
        console.error("Error fetching student profile:", error);
    }
});
</script>
