<template>
    <div class="mt-[4.5rem] sticky top-0 overflow-y-auto max-h-dvh rounded-xl shadow-md/25 p-5">
        <!-- First section -->
        <div class="flex mb-7 justify-between items-center">
            <!-- Top Left Side -->
            <div class="flex items-center gap-5">
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
                <span class="px-10 space-y-1">
                    <h1 class="text-3xl font-bold">{{ data.position }}</h1>
                    <span>
                        <NuxtLink to="/jobs/company">
                            <h2 class="text-[#15543A] text-xl font-semibold">
                                {{ data.name }}
                            </h2>
                        </NuxtLink>
                        <div class="flex space-x-2 mt-2">
                            <span
                                class="px-2 py-1 text-md bg-green-100 text-green-700 rounded-full"
                            >
                                {{ data.jobType }}
                            </span>
                            <span
                                class="px-2 py-1 text-md bg-green-100 text-green-700 rounded-full"
                            >
                                {{ data.experienceType }}
                            </span>
                        </div>
                    </span>
                </span>
            </div>
            <!-- Top Right Side -->
            <div>
                <span class="flex items-center justify-between">
                    <USwitch
                        v-model="isOpen"
                        size="xl"
                        @update:model-value="
                            (value) => {
                                emit('update:open', value);
                            }
                        "
                    />
                    <span
                        v-if="isOpen"
                        class="text-primary-800 text-lg dark:text-primary font-bold"
                    >
                        Open
                    </span>
                    <span v-else class="text-error-800 text-lg dark:text-error font-bold"
                        >Closed</span
                    >
                </span>
                <!-- Edit Button -->
                <button
                    class="px-4 py-2 border border-gray-400 rounded-md text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center mt-4 gap-2"
                >
                    <Icon
                        name="material-symbols:edit-square-outline-rounded"
                        class="size-[1.5em]"
                    />
                    Edit Post
                </button>
            </div>
        </div>

        <USeparator class="mt-2" />
        <!-- Information -->
        <div class="space-y-3">
            <h2 class="text-[#15543A] text-xl font-bold">About This Job</h2>
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
    </div>
</template>

<script setup lang="ts">
import type { JobApplication } from "~/data/mockData";

const props = defineProps<{
    data: JobApplication;
    open: boolean;
}>();

const emit = defineEmits(["update:open"]);

const isOpen = ref(false);

onMounted(() => {
    isOpen.value = props.open;
});
</script>
