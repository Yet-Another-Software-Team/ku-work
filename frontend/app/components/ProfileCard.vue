<template>
    <div class="rounded-lg">
        <!-- Header -->
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-6">Profile</h1>

        <!-- Top Section -->
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <!-- Profile Image -->
            <div class="size-fit rounded-full flex items-center justify-center overflow-hidden">
                <div v-if="mockData.profile.photo">
                    <img
                        :src="mockData.profile.photo"
                        alt="Profile photo"
                        class="w-full h-full object-cover"
                    />
                </div>
                <div v-else>
                    <Icon name="ic:baseline-account-circle" class="size-[10em]" />
                </div>
            </div>

            <!-- Info -->
            <div class="col-span-1 lg:col-span-2 text-xl">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                    {{ mockData.profile.name }}
                </h2>
                <p class="text-gray-600 dark:text-gray-300">
                    {{ mockData.profile.major }}
                </p>

                <p class="mt-2 text-gray-800 dark:text-gray-200 font-semibold">
                    Age: <span class="font-normal">{{ age }}</span>
                </p>
                <p class="text-gray-800 dark:text-gray-200 font-semibold">
                    Phone: <span class="font-normal">{{ mockData.profile.phone }}</span>
                </p>
                <p class="text-gray-800 dark:text-gray-200 font-semibold">
                    Email: <span class="font-normal">{{ email }}</span>
                </p>
            </div>

            <!-- Edit Button -->
            <button
                class="mb-auto mr-auto px-4 py-2 border border-gray-400 rounded-md text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
            >
                <Icon name="material-symbols:edit-square-outline-rounded" class="size-[1.5em]" />
                Edit Profile
            </button>
        </div>

        <!-- Divider -->
        <hr class="my-6 border-gray-300 dark:border-gray-600" />

        <!-- Bottom Section -->
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 text-xl">
            <!-- Connections -->
            <div>
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">Connections</h3>
                <ul class="space-y-2 text-primary-600">
                    <li>
                        <a
                            :href="mockData.profile.linkedIn"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon name="devicon:linkedin" class="size-[1.5em]" />
                            {{ mockData.profile.name }}
                        </a>
                    </li>
                    <li>
                        <a
                            :href="mockData.profile.github"
                            target="_blank"
                            class="flex items-center gap-2 hover:underline"
                        >
                            <Icon
                                name="devicon:github"
                                class="size-[1.5em] bg-white rounded-full"
                            />
                            {{ mockData.profile.name }}
                        </a>
                    </li>
                </ul>
            </div>

            <!-- About Me -->
            <div class="col-span-1 md:col-span-2 lg:col-span-3">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">About me</h3>
                <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                    {{ mockData.profile.aboutMe }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Profile {
    profile: {
        name: string;
        id: string;
        approved: boolean;
        created: string;
        phone: string;
        photo: string;
        birthDate: string;
        aboutMe: string;
        github: string;
        linkedIn: string;
        studentId: string;
        major: "Software Engineering" | "Computer Science";
        status: "Graduated" | "Undergraduate";
        statusFile: string;
    };
}

const mockData: Profile = {
    profile: {
        name: "John Doe",
        id: "123456",
        approved: true,
        created: "2023-01-01",
        phone: "012-345-6789",
        photo: "",
        birthDate: "2003-01-01",
        aboutMe:
            "Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions. REALLY LONG TEXT TO TEST THE LAYOUT. Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions.",
        github: "https://github.com",
        linkedIn: "https://linkedin.com/",
        studentId: "6xxxxxxxxx",
        major: "Software Engineering",
        status: "Undergraduate",
        statusFile: "https://example.com/status.pdf",
    },
};

// Compute age
const age = computed(() => {
    const birth = new Date(mockData.profile.birthDate);
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
