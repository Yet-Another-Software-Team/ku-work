<!-- Job Post component that shows on company dashboard -->
<template>
    <UCard class="bg-white dark:bg-[#1f2937]">
        <template #header>
            <div class="flex justify-between">
                <div class="flex gap-x-4 items-center">
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
                <div class="flex items-center ml-auto">
                    <UButton
                        v-if="data.approvalStatus === 'accepted'"
                        variant="ghost"
                        color="neutral"
                        class="cursor-pointer"
                        :aria-pressed="emailNotifyOn"
                        aria-label="Toggle applicant email notifications"
                        @click.stop.prevent="toggleEmailNotification"
                    >
                        <Icon
                            :name="
                                emailNotifyOn
                                    ? 'fluent:mail-24-regular'
                                    : 'fluent:mail-off-24-regular'
                            "
                            size="24"
                            class="hover:text-primary transition-colors duration-300"
                        />
                    </UButton>
                    <UModal v-model:open="openJobEditForm">
                        <UButton
                            variant="ghost"
                            color="neutral"
                            class="cursor-pointer"
                            aria-label="Edit job"
                            @click.stop.prevent="openJobEditForm = true"
                        >
                            <Icon
                                name="fluent:edit-24-regular"
                                size="20"
                                class="hover:text-primary transition-colors duration-300"
                            />
                        </UButton>
                        <template #content>
                            <JobEditForm :data="data" @close="handleCloseEditForm" />
                        </template>
                    </UModal>
                </div>
            </div>
            <div @click="selectJob">
                <UTooltip :text="data.position">
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
            </div>
        </template>

        <div class="flex flex-col gap-y-4">
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-check-circle" size="18" class="text-primary-500" />
                    <span class="text-neutral-500 dark:text-neutral-200">Accepted</span>
                </div>
                <div>{{ data.accepted }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-access-time" size="18" class="text-warning-500" />
                    <span class="text-neutral-500 dark:text-neutral-200">Pending</span>
                </div>
                <div>{{ data.pending }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:baseline-do-disturb" size="18" class="text-error-500" />
                    <span class="text-neutral-500 dark:text-neutral-200">Rejected</span>
                </div>
                <div>{{ data.rejected }}</div>
            </div>
            <div class="flex justify-between">
                <div class="flex items-center gap-x-2">
                    <Icon name="ic:outline-insert-drive-file" size="18" class="text-neutral-500" />
                    <span class="text-neutral-500 dark:text-neutral-200">Total Applicants</span>
                </div>
                <div>{{ (data.rejected ?? 0) + (data.accepted ?? 0) + (data.pending ?? 0) }}</div>
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import type { JobPost } from "~/data/datatypes";

const emit = defineEmits<{
    "update:open": [value: boolean];
    "update:notifyOnApplication": [value: boolean];
    close: [];
}>();

const openJobEditForm = ref(false);

const props = defineProps<{
    data: JobPost;
}>();

function getNotifyFlag(d: unknown): boolean {
    // Accept both camelCase and snake_case from API just in case
    const o = (d as Record<string, unknown>) ?? {};
    const vRaw = (o["notifyOnApplication"] ?? o["notify_on_application"]) as unknown;
    return typeof vRaw === "boolean" ? vRaw : Boolean(vRaw);
}

const emailNotifyOn = ref(getNotifyFlag(props.data));
watch(
    () => props.data,
    (val) => (emailNotifyOn.value = getNotifyFlag(val))
);

const { add: addToast } = useToast();

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

function toggleEmailNotification() {
    const next = !emailNotifyOn.value;
    emailNotifyOn.value = next; // instant UI feedback
    emit("update:notifyOnApplication", next);
    addToast({
        title: next ? "Notifications enabled" : "Notifications disabled",
        description: next
            ? "You will receive emails for new applicants."
            : "You will not receive emails for new applicants.",
        color: "success",
    });
}
</script>
