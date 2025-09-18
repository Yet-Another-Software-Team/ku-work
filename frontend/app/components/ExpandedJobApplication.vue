<template>
    <div v-if="isSelected" class="w-[20em] mt-[4.5rem] sticky top-0 overflow-y-auto max-h-dvh">
        <!-- First section -->
        <div class="flex mb-7">
            <!-- Profile -->
            <span>
                <img
                    v-if="data.logo"
                    :src="data.logo"
                    alt="Company Logo"
                    class="rounded-full size-[6em]"
                />
                <img
                    v-else
                    src="/images/background.png"
                    alt="Company Logo"
                    class="rounded-full size-[6em]"
                />
            </span>
            <span class="mx-3 space-y-1">
                <h1 class="text-xl font-bold">{{ data.position }}</h1>
                <span>
                    <h2 class="text-[#15543A] text-md font-semibold">{{ data.name }}</h2>
                    <p class="text-xs">{{ timeAgo(data.createdAt) }}</p>
                </span>
            </span>
        </div>

        <USeparator class="mt-2" />
        <span class="flex flex-row justify-center text-sm">
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-[#807D89] font-semibold">Job Type</p>
                <p class="font-bold">{{ data.jobType }}</p>
            </span>
            <span>
                <USeparator orientation="vertical" class="h-full items-stretch" size="sm" />
            </span>
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-[#807D89] font-semibold">Experience</p>
                <p class="font-bold">{{ data.experienceType }}</p>
            </span>
            <span>
                <USeparator orientation="vertical" class="h-full items-stretch" size="sm" />
            </span>
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-[#807D89] font-semibold">Salary Range</p>
                <p class="font-bold">
                    {{ formatSalary(data.minSalary) }}
                    -
                    {{ formatSalary(data.maxSalary) }}
                </p>
            </span>
        </span>
        <USeparator class="mb-2" />

        <div>
            <h2 class="font-semibold">About This Job</h2>
            <p>
                <span class="text-[#15543A] font-semibold">Position Title: </span>
                {{ data.position }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold capitalize">Location: </span>
                {{ data.location }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold">Duration: </span>
                {{ data.duration }}
            </p>
            <p>
                <span class="text-[#15543A] font-semibold">Description:</span>
                <br />
                {{ data.description }}
            </p>
        </div>
        <StudentApplyButton v-if="!isViewer" label="Apply" />
    </div>
</template>

<script setup lang="ts">
import { USeparator } from "#components";
import type { JobApplication } from "~/data/mockData";
import StudentApplyButton from "./StudentApplyButton.vue";

defineProps<{
    data: JobApplication;
    isSelected: boolean;
    isViewer: boolean;
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

function formatSalary(salary: number): string {
    return new Intl.NumberFormat("en", { notation: "compact" }).format(salary);
}
</script>
