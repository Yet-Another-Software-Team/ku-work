<template>
    <div v-if="isSelected" class="w-[20em] mt-[4.5rem] sticky top-10 overflow-y-auto max-h-dvh">
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
                    <h1 class="text-xl font-bold truncate max-w-[200px]">{{ data.name }}</h1>
                </UTooltip>
                <span>
                    <NuxtLink :to="`/jobs/${data.companyId}`">
                        <UTooltip :text="data.companyName">
                            <h2
                                class="text-primary-700 text-base font-semibold truncate max-w-[200px] capitalize"
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
        <span class="flex flex-row justify-center text-sm">
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-primary-700 font-semibold">Job Type</p>
                <p class="font-bold">{{ formatJobType(data.jobType) }}</p>
            </span>
            <span>
                <USeparator orientation="vertical" class="h-full items-stretch" size="sm" />
            </span>
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-primary-700 font-semibold">Experience</p>
                <p class="font-bold">{{ formatExperience(data.experience) }}</p>
            </span>
            <span>
                <USeparator orientation="vertical" class="h-full items-stretch" size="sm" />
            </span>
            <span class="text-center w-1/3 p-2 space-y-1">
                <p class="text-primary-700 font-semibold">Salary Range</p>
                <p class="font-bold">
                    {{ formatSalary(data.minSalary) }}
                    -
                    {{ formatSalary(data.maxSalary) }}
                </p>
            </span>
        </span>
        <USeparator class="mb-2" />

        <div class="overflow-x-hidden">
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
            <p class="whitespace-pre-wrap break-words overflow-x-hidden">
                <span class="text-primary-700 font-semibold">Description:</span>
                <br />
                <span class="break-words">{{ data.description }}</span>
            </p>
        </div>
        <StudentApplyButton v-if="!isViewer" :job-id="data.id" label="Apply" />
    </div>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";
import { formatSalary, formatJobType, formatExperience } from "~/utils/formatter";

const runtimeConfig = useRuntimeConfig();
const timeAgo = useTimeAgo();

defineProps<{
    data: JobPost;
    isSelected: boolean;
    isViewer: boolean;
}>();
</script>
