<template>
    <div
        class="flex items-center justify-between w-full h-[8em] border-b-1 border-gray-400"
        :class="!isSelected ? 'bg-transparent hover:bg-gray-50 cursor-pointer' : 'bg-[#fdfdfd]'"
        @click="$emit('click')"
    >
        <!-- Left side -->
        <div class="flex flex-row items-center gap-4 h-full">
            <!-- Green line -->
            <div
                class="w-[0.5em] h-full"
                :class="isSelected ? 'bg-green-500' : 'bg-transparent'"
            ></div>

            <!-- Icon -->
            <div class="flex items-center justify-center w-20 h-20 rounded-md">
                <img
                    v-if="data.logo"
                    :src="data.logo"
                    alt="Company Logo"
                    class="size-full rounded-md"
                />
                <img
                    v-else
                    src="/images/background.png"
                    alt="Company Logo"
                    class="size-full rounded-md"
                />
            </div>

            <!-- Job info -->
            <div class="flex flex-col h-full py-5">
                <h2 class="font-semibold text-gray-900">{{ data.name }}</h2>
                <p class="text-sm text-gray-500">{{ data.location }}</p>
                <div class="flex space-x-2 mt-2">
                    <span class="px-2 py-1 text-xs bg-green-100 text-green-700 rounded-full">
                        {{ data.jobType }}
                    </span>
                    <span class="px-2 py-1 text-xs bg-green-100 text-green-700 rounded-full">
                        {{ data.experienceType }}
                    </span>
                </div>
            </div>
        </div>

        <!-- Right side -->
        <div class="text-xs text-gray-400 whitespace-nowrap h-full p-5">
            {{ timeAgo(data.createdAt) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobApplication } from "~/data/mockData";

defineProps<{
    data: JobApplication;
    isSelected: boolean;
}>();

defineEmits<{
    (e: "click"): void;
}>();

function timeAgo(createdAt: string): string {
    const createdDate = new Date(createdAt);
    const now = new Date();

    const diffMs = now.getTime() - createdDate.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    const diffMonth = Math.floor(diffDay / 30);

    if (diffMonth > 0) return `${diffMonth} month${diffMonth > 1 ? "s" : ""} ago`;
    if (diffDay > 0) return `${diffDay} day${diffDay > 1 ? "s" : ""} ago`;
    if (diffHour > 0) return `${diffHour} hour${diffHour > 1 ? "s" : ""} ago`;
    if (diffMin > 0) return `${diffMin} minute${diffMin > 1 ? "s" : ""} ago`;
    return "just now";
}
</script>
