<template>
    <!-- Reactivate Modal -->
    <UModal
        v-model:open="openReactivateModal"
        :ui="{
            overlay: 'fixed inset-0 bg-black/50',
            content: 'w-full max-w-md',
        }"
    >
        <template #content>
            <div class="p-6">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">
                    Confirm Reactivation
                </h3>
                <p class="text-gray-700 dark:text-gray-300">
                    Are you sure you want to reactivate your profile?
                </p>
                <p class="mt-2">
                    To confirm, type
                    <strong class="text-green-600">REACTIVATE</strong> below.
                </p>
                <UInput
                    v-model="confirmReactivate"
                    class="mt-4 w-full"
                    placeholder="Type REACTIVATE to confirm"
                    aria-label="Confirmation input: type REACTIVATE to confirm"
                />
                <div class="flex justify-center md:col-span-2 mt-6">
                    <TurnstileWidget @callback="(tk) => (cfToken = tk)" />
                </div>
                <div class="flex justify-end gap-2 mt-4">
                    <UButton
                        variant="outline"
                        color="neutral"
                        @click="
                            openReactivateModal = false;
                            confirmReactivate = '';
                            emit('update:close', false);
                        "
                    >
                        Cancel
                    </UButton>
                    <UButton
                        :disabled="confirmReactivate !== 'REACTIVATE' || !cfToken"
                        variant="solid"
                        color="primary"
                        class="ml-2"
                        @click="onReactivated"
                    >
                        Reactivate
                    </UButton>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
const openReactivateModal = ref(false);

const props = defineProps<{
    open: boolean;
}>();

watch(
    () => props.open,
    (newVal) => {
        openReactivateModal.value = newVal;
    }
);

const emit = defineEmits<{
    (e: "update:close", value: boolean): void;
}>();

const confirmReactivate = ref<string>("");
const cfToken = ref<string>("");

const api = useApi();
const toast = useToast();

// Handle Reactivation
async function onReactivated() {
    try {
        const response = await api.post("/me/reactivate", null, {
            headers: {
                "X-Turnstile-Token": cfToken.value,
            },
            withCredentials: true,
        });
        toast.add({
            title: "Account Reactivated",
            description: "Your account has been reactivated.",
            color: "success",
        });
        navigateTo("/", { replace: true });
    } catch (error) {
        toast.add({
            title: "Failed to reactivate profile",
            description: (error as { message: string }).message,
            color: "error",
        });
    } finally {
        openReactivateModal.value = false;
        confirmReactivate.value = "";
        cfToken.value = "";
        emit("update:close", false);
    }
}
</script>
