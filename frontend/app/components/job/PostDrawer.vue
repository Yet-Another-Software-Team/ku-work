<template>
    <UDrawer :open="isSelected">
        <template #content>
            <div class="flex items-center justify-end">
                <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="emit('close')" />
            </div>
            <div
                v-if="isSelected"
                class="sticky top-10 overflow-y-auto p-4 sm:p-8 gap-2 max-w-[95vw] h-[95vh] sm:max-w-none"
            >
                <!-- First section -->
                <div
                    class="flex flex-col sm:flex-row items-center sm:items-start mb-7 gap-3 sm:gap-0"
                >
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
                    <span class="mx-0 sm:mx-3 space-y-1 text-center sm:text-left w-full sm:w-auto">
                        <UTooltip :text="data.name">
                            <h1
                                class="text-lg sm:text-xl font-bold truncate max-w-full sm:max-w-[200px]"
                            >
                                {{ data.name }}
                            </h1>
                        </UTooltip>
                        <span>
                            <NuxtLink :to="`/jobs/${data.companyId}`">
                                <UTooltip :text="data.companyName">
                                    <h2
                                        class="text-primary-700 text-sm sm:text-md font-semibold truncate max-w-full sm:max-w-[200px]"
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

                <div
                    class="flex flex-col sm:flex-row justify-center text-xs sm:text-sm gap-2 sm:gap-0"
                >
                    <div class="text-center w-full sm:w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Job Type</p>
                        <p class="font-bold">{{ formatJobType(data.jobType) }}</p>
                    </div>
                    <USeparator orientation="vertical" class="h-full hidden sm:block" size="sm" />
                    <div class="text-center w-full sm:w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Experience</p>
                        <p class="font-bold">{{ formatExperience(data.experience) }}</p>
                    </div>
                    <USeparator orientation="vertical" class="h-full hidden sm:block" size="sm" />
                    <div class="text-center w-full sm:w-1/3 p-2 space-y-1">
                        <p class="text-primary-700 font-semibold">Salary Range</p>
                        <p class="font-bold">
                            {{ formatSalary(data.minSalary) }} - {{ formatSalary(data.maxSalary) }}
                        </p>
                    </div>
                </div>

                <USeparator class="mb-2" />

                <div class="text-sm">
                    <h2 class="font-semibold text-base mb-2">About This Job</h2>
                    <p class="mb-2">
                        <span class="text-primary-700 font-semibold">Position Title: </span>
                        <UTooltip :text="data.position">
                            <span
                                class="truncate inline-block max-w-[200px] sm:max-w-[250px] align-bottom"
                            >
                                {{ data.position }}
                            </span>
                        </UTooltip>
                    </p>
                    <p class="mb-2">
                        <span class="text-primary-700 font-semibold capitalize">Location: </span>
                        {{ data.location }}
                    </p>
                    <p class="mb-2">
                        <span class="text-primary-700 font-semibold">Duration: </span>
                        {{ data.duration }}
                    </p>
                    <p>
                        <span class="text-primary-700 font-semibold">Description:</span><br />
                        <span class="whitespace-pre-wrap break-words">{{ data.description }}</span>
                    </p>
                </div>

                <StudentApplyButton
                    v-if="!isViewer"
                    :job-id="data.id"
                    label="Apply"
                    class="mt-4 w-full"
                />
            </div>
        </template>
    </UDrawer>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";
import { formatSalary, formatJobType, formatExperience } from "~/utils/formatter";

const runtimeConfig = useRuntimeConfig();
const timeAgo = useTimeAgo();

defineProps<{
    data: JobPost;
    isSelected: boolean;
    isViewer: boolean;
}>();

const emit = defineEmits<{
    (e: "close"): void;
}>();
</script>
