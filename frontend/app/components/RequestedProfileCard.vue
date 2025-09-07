<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>

        <!-- Top Section -->
        <div class="flex flex-wrap">
            <!-- Profile Image -->
            <div class="w-[12em] mr-5">
                <div v-if="data.profile.photo" class="w-40 h-40 flex-shrink-0">
                    <img
                        :src="data.profile.photo"
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
                    {{ data.profile.major }}
                </p>

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

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />

        <!-- Bottom Section -->
        <div class="flex text-xl">
            <!-- Connections -->
            <div class="w-[12rem] mr-5">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">Connections</h3>
                <ul class="space-y-2 text-primary-600">
                    <li>
                        <a
                            :href="data.profile.linkedIn"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:linkedin" class="size-[2em]" />
                            <span class="w-[10em] truncate">{{ data.profile.name }}</span>
                        </a>
                    </li>
                    <li>
                        <a
                            :href="data.profile.github"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:github" class="size-[2em] bg-white rounded-full" />
                            <span class="w-[10em] truncate">{{ data.profile.name }}</span>
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div>
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About me</h3>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                    {{ data.profile.aboutMe }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { mockData } from "~/data/mockData";
import { multipleMockData } from "~/data/mockData";

const route = useRoute();
const query = route.query.id ?? 0;
const data: typeof mockData =
    multipleMockData[Number(query) % 3] ?? (multipleMockData[0] as typeof mockData);

// Compute age
const age = computed(() => {
    const birth = new Date(data?.profile?.birthDate ?? "");
    const today = new Date();
    let years = today.getFullYear() - birth.getFullYear();
    const m = today.getMonth() - birth.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) {
        years--;
    }
    return years;
});

const email = "john.doe@ku.th";
</script>
