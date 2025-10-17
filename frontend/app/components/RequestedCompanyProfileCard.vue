<template>
    <div class="flex justify-center items-center w-[28em] h-full">
        <div class="flex w-full m-2 p-4 shadow-md/20 bg-[#fdfdfd] dark:bg-[#1f2937] rounded-md">
            <!-- Left: logo -->
            <div class="flex items-center justify-center w-20 h-20 flex-shrink-0 mr-3">
                <img
                    v-if="data.photoId"
                    :src="`${runtimeConfig.public.apiBaseUrl}/files/${data.photoId}`"
                    alt="Company Logo"
                    class="w-20 h-20 object-cover rounded-md"
                />
                <img
                    v-else
                    src="/images/company.png"
                    alt="Company Logo"
                    class="w-20 h-20 object-cover rounded-md"
                />
            </div>

            <!-- Middle: content -->
            <div class="flex flex-col justify-center flex-1 min-w-0">
                <p class="font-semibold truncate">{{ data.name }} - {{ data.position }}</p>
                <p class="text-sm text-gray-600 dark:text-gray-300 truncate">
                    {{ data.companyName }} · {{ data.location }}
                </p>
                <div class="flex gap-2 mt-2">
                    <span class="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded-full capitalize">
                        {{ data.jobType }}
                    </span>
                    <span class="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded-full capitalize">
                        {{ data.experience }}
                    </span>
                </div>
            </div>

            <!-- Right: actions -->
            <div class="flex flex-col items-end justify-between ml-3">
                <span class="text-xs text-gray-400 whitespace-nowrap">{{ timeAgo(data.createdAt) }}</span>
                <div class="flex gap-2 mt-2">
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="error"
                        label="Decline"
                        :icon="'iconoir:xmark'"
                        :loading="loading === 'decline'"
                        @click="() => approve(false)"
                    />
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        :loading="loading === 'accept'"
                        @click="() => approve(true)"
                    />
                </div>
            </div>
        </div>
    </div>
    
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/mockData";
import { useApi } from "~/composables/useApi";

const runtimeConfig = useRuntimeConfig();

const props = defineProps<{ data: JobPost }>();
const emit = defineEmits<{ (e: "resolved", id: number): void }>();

const { post, showErrorToast, showSuccessToast } = useApi();
const loading = ref<"accept" | "decline" | null>(null);

async function approve(approve: boolean) {
    try {
        loading.value = approve ? "accept" : "decline";
        await post(`/jobs/${props.data.id}/approval`, { approve });
        showSuccessToast(approve ? "Job approved" : "Job rejected");
        emit("resolved", props.data.id);
    } catch (e: any) {
        showErrorToast(e, approve ? "Approve failed" : "Reject failed");
    } finally {
        loading.value = null;
    }
}

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
