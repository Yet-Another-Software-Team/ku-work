<template>
    <div
        class="relative flex justify-between items-center w-full min-h-[8em] border-b border-gray-400"
        :class="
            !isSelected
                ? 'bg-transparent hover:bg-gray-50 dark:hover:bg-[#1f2937] cursor-pointer'
                : 'bg-[#fdfdfd] dark:bg-[#013B49]'
        "
        @click="$emit('click')"
    >
        <!-- Green line positioned outside the flow -->
        <div
            class="absolute left-0 top-0 w-[0.3em] h-full"
            :class="isSelected ? 'bg-green-500' : 'bg-transparent'"
        ></div>

        <!-- Left side -->
        <div class="flex flex-row items-center gap-4 pl-[1em] flex-1 min-w-0">
            <!-- Icon -->
            <div class="flex items-center justify-center w-20 h-20 rounded-md">
                <img
                    v-if="data.photoId"
                    :src="`${runtimeConfig.public.apiBaseUrl}/files/${data.photoId}`"
                    alt="Company Logo"
                    class="size-full rounded-md object-cover"
                />
                <img
                    v-else
                    src="/images/background.png"
                    alt="Company Logo"
                    class="size-full rounded-md object-cover"
                />
            </div>

            <!-- Job info -->
            <!-- Job Details -->
            <div class="flex flex-col py-5 flex-1 min-w-0">
                <UTooltip :text="data.position">
                    <h2 class="font-semibold text-gray-900 dark:text-[#fdfdfd] truncate">
                        {{ data.position }}
                    </h2>
                </UTooltip>
                <UTooltip :text="`${data.companyName} • ${data.location}`">
                    <p class="capitalize text-sm text-gray-500 dark:text-gray-200 truncate">
                        {{ data.companyName }} • {{ data.location }}
                    </p>
                </UTooltip>
                <div class="flex flex-wrap items-center gap-2 mt-2">
                    <UBadge
                        v-if="data.jobType"
                        color="primary"
                        variant="subtle"
                        size="sm"
                        class="capitalize"
                    >
                        {{ formatJobType(data.jobType) }}
                    </UBadge>
                    <UBadge
                        v-if="data.experience"
                        color="neutral"
                        variant="subtle"
                        size="sm"
                        class="capitalize"
                    >
                        {{ formatExperience(data.experience) }}
                    </UBadge>
                </div>
            </div>
        </div>

        <!-- Right side -->
        <div class="text-xs text-gray-400 whitespace-nowrap p-5 pb-1 flex-shrink-0">
            {{ timeAgo(data.createdAt) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";
import { formatJobType, formatExperience } from "~/utils/formatter";

const runtimeConfig = useRuntimeConfig();

defineProps<{
    data: JobPost;
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
