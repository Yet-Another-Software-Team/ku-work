<template>
    <UDrawer :open="isSelected">
        <template #content>
            <div v-if="isSelected" class="w-[20em] mt-[4.5rem] sticky top-10 overflow-y-auto p-4">
                <!-- First section -->
                <div class="flex items-center mb-7">
                    <!-- Profile -->
                    <span class="flex-shrink-0">
                        <img
                            v-if="data.photoId"
                            :src="`${runtimeConfig.public.apiBaseUrl}/files/${data.photoId}`"
                            alt="Company Logo"
                            class="rounded-full size-[6em] object-cover"
                        />
                        <img
                            v-else
                            src="~/assets/images/background.png"
                            alt="Company Logo"
                            class="rounded-full size-[6em] object-cover"
                        />
                    </span>
                    <span class="mx-3 space-y-1">
                        <UTooltip :text="data.name">
                            <h1 class="text-xl font-bold truncate max-w-[200px]">
                                {{ data.name }}
                            </h1>
                        </UTooltip>
                        <span>
                            <NuxtLink :to="`/jobs/${data.companyId}`">
                                <UTooltip :text="data.companyName">
                                    <h2
                                        class="text-primary-700 text-md font-semibold truncate max-w-[200px]"
                                    >
                                        {{ data.companyName }}
                                    </h2>
                                </UTooltip>
                            </NuxtLink>
                            <p class="text-xs">{{ timeAgo(data.createdAt) }}</p>
                        </span>
                    </span>
                </div>

                <USeparator class="mt-2" />

                <div class="flex flex-row justify-center text-sm">
                    <div class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Job Type</p>
                        <p class="font-bold">{{ formatJobType(data.jobType) }}</p>
                    </div>
                    <USeparator orientation="vertical" class="h-full" size="sm" />
                    <div class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Experience</p>
                        <p class="font-bold">{{ formatExperience(data.experience) }}</p>
                    </div>
                    <USeparator orientation="vertical" class="h-full" size="sm" />
                    <div class="text-center w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Salary Range</p>
                        <p class="font-bold">
                            {{ formatSalary(data.minSalary) }} - {{ formatSalary(data.maxSalary) }}
                        </p>
                    </div>
                </div>

                <USeparator class="mb-2" />

                <div>
                    <h2 class="font-semibold">About This Job</h2>
                    <p>
                        <span class="text-primary-700 font-semibold">Position Title: </span>
                        <UTooltip :text="data.position">
                            <span class="truncate inline-block max-w-[250px] align-bottom">
                                {{ data.position }}
                            </span>
                        </UTooltip>
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold capitalize">Location: </span>
                        {{ data.location }}
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold">Duration: </span>
                        {{ data.duration }}
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold">Description:</span><br />
                        {{ data.description }}
                    </p>
                </div>

                <StudentApplyButton v-if="!isViewer" :job-id="data.id" label="Apply" />
            </div>
        </template>
    </UDrawer>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";

const runtimeConfig = useRuntimeConfig();

defineProps<{
    data: JobPost;
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

function formatJobType(type: string): string {
    const typeMap: Record<string, string> = {
        fulltime: "Full Time",
        parttime: "Part Time",
        contract: "Contract",
        casual: "Casual",
        internship: "Internship",
    };
    return typeMap[type.toLowerCase()] || type;
}

function formatExperience(exp: string): string {
    const expMap: Record<string, string> = {
        newgrad: "New Grad",
        junior: "Junior",
        senior: "Senior",
        manager: "Manager",
        internship: "Internship",
    };
    return expMap[exp.toLowerCase()] || exp;
}
</script>
