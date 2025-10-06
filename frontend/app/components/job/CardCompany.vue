<template>
    <UCard>
        <template #header>
            <div class="flex justify-between">
                <div class="flex gap-x-4">
                    <USwitch
                        :model-value="props.open"
                        :ui="{ base: 'cursor-pointer' }"
                        @update:model-value="
                            (value) => {
                                emit('update:open', value);
                            }
                        "
                    />
                    <span v-if="props.open" class="text-primary-800 dark:text-primary font-bold">
                        Open
                    </span>
                    <span v-else class="text-error-800 dark:text-error font-bold">Closed</span>
                </div>
                <UDropdownMenu class="cursor-pointer" :items="menuItems">
                    <Icon
                        name="ic:baseline-more-vert"
                        size="18"
                        class="hover:text-primary transition-colors duration-300"
                    />
                </UDropdownMenu>
            </div>
            <UTooltip :text="props.position" @click="selectJob">
                <div class="flex mt-2 font-semibold text-lg truncate">
                    {{ props.position }}
                </div>
            </UTooltip>
        </template>

        <div class="flex flex-col gap-y-4">
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-access-time" size="18" class="text-warning-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Pending</span>
                </div>
                <div>{{ props.pending }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-check-circle" size="18" class="text-primary-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Accepted</span>
                </div>
                <div>{{ props.accepted }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:baseline-do-disturb" size="18" class="text-error-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Rejected</span>
                </div>
                <div>{{ props.rejected }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-insert-drive-file" size="18" class="text-neutral-700" />
                    <span class="text-neutral-500 dark:text-neutral-400">Total Applicants</span>
                </div>
                <div>{{ props.rejected + props.accepted + props.pending }}</div>
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import type { DropdownMenuItem } from "@nuxt/ui";

const emit = defineEmits<{
    (e: "update:open", value: boolean): void;
}>();

const props = defineProps({
    jobID: {
        type: String,
        required: true,
    },
    open: {
        type: Boolean,
        required: true,
    },
    position: {
        type: String,
        required: true,
    },
    accepted: {
        type: Number,
        required: true,
    },
    rejected: {
        type: Number,
        required: true,
    },
    pending: {
        type: Number,
        required: true,
    },
});

const menuItems = ref<DropdownMenuItem[]>([
    {
        label: "Edit",
        icon: "i-lucide-pencil",
        class: "cursor-pointer",
        onClick: () => {
            console.log("Action 1 clicked");
        },
    },
]);

function selectJob() {
    console.log("Job ID:", props.jobID);
    navigateTo(`/dashboard/post/${props.jobID}`);
}
</script>
