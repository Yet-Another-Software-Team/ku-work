<template>
    <UCard>
        <template #header>
            <div class="flex justify-between">
                <div class="flex gap-x-4">
                    <USwitch
                        :model-value="data.open"
                        :ui="{ base: 'cursor-pointer' }"
                        @update:model-value="
                            (value) => {
                                emit('update:open', value);
                            }
                        "
                    />
                    <span v-if="data.open" class="text-primary-800 dark:text-primary font-bold">
                        Open
                    </span>
                    <span v-else class="text-error-800 dark:text-error font-bold">Closed</span>
                </div>
                <div class="flex gap-x-2 items-center">
                    <UBadge :color="colorPicker()">{{ data.approvalStatus }}</UBadge>
                    <UDropdownMenu class="cursor-pointer" :items="menuItems">
                        <Icon
                            name="ic:baseline-more-vert"
                            size="18"
                            class="hover:text-primary transition-colors duration-300"
                        />
                    </UDropdownMenu>
                    <UModal v-model:open="openJobEditForm">
                        <template #content>
                            <JobEditForm :data="data" @close="handleCloseEditForm" />
                        </template>
                    </UModal>
                </div>
            </div>
            <UTooltip :text="data.position" @click="selectJob">
                <div
                    class="flex mt-2 font-semibold text-lg truncate"
                    :class="
                        data.approvalStatus === 'accepted'
                            ? 'cursor-pointer hover:underline'
                            : 'text-gray-500 cursor-not-allowed'
                    "
                >
                    {{ data.position }}
                </div>
            </UTooltip>
        </template>

        <div class="flex flex-col gap-y-4">
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-check-circle" size="18" class="text-primary-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Accepted</span>
                </div>
                <div>{{ data.accepted }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-access-time" size="18" class="text-warning-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Pending</span>
                </div>
                <div>{{ data.pending }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:baseline-do-disturb" size="18" class="text-error-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Rejected</span>
                </div>
                <div>{{ data.rejected }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-insert-drive-file" size="18" class="text-neutral-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Total Applicants</span>
                </div>
                <div>{{ data.rejected! + data.accepted! + data.pending! }}</div>
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import type { DropdownMenuItem } from "@nuxt/ui";
import type { JobPost } from "~/data/mockData";

const emit = defineEmits<{
    (e: "update:open", value: boolean): void;
    (e: "close"): void;
}>();

const openJobEditForm = ref(false);

const props = defineProps<{
    data: JobPost;
}>();

const menuItems = ref<DropdownMenuItem[]>([
    {
        label: "Edit",
        icon: "i-lucide-pencil",
        class: "cursor-pointer",
        onClick: () => {
            openJobEditForm.value = true;
        },
    },
]);

function colorPicker() {
    if (props.data.approvalStatus === "pending") {
        return "warning";
    } else if (props.data.approvalStatus === "accepted") {
        return "primary";
    } else if (props.data.approvalStatus === "rejected") {
        return "error";
    } else {
        return "neutral";
    }
}

function selectJob() {
    console.log("Job ID:", props.data.id);
    if (props.data.approvalStatus === "accepted") {
        navigateTo(`/dashboard/${props.data.id}`);
    }
}

function handleCloseEditForm() {
    openJobEditForm.value = false;
    emit("close");
}
</script>
