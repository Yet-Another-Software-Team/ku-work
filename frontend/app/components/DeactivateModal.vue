<template>
    <!-- Deactivate Modal -->
    <UModal
        v-model:open="openDeactivateModal"
        :ui="{
            overlay: 'fixed inset-0 bg-black/50',
            content: 'w-full max-w-md',
        }"
    >
        <template #content>
            <div class="p-6">
                <h3 class="font-semibold text-gray-800 dark:text-white mb-2">
                    Confirm Deactivation
                </h3>
                <p class="text-gray-700 dark:text-gray-300">
                    Are you sure you want to deactivate your profile?
                </p>
                <p class="mt-2">
                    To confirm, type
                    <strong class="text-red-600">DEACTIVATE</strong> below.
                </p>
                <UInput
                    v-model="confirmDeactivate"
                    class="mt-4 w-full"
                    placeholder="Type DEACTIVATE to confirm"
                    aria-label="Confirmation input: type DEACTIVATE to confirm"
                />
                <div class="flex justify-center md:col-span-2">
                    <TurnstileWidget @callback="(tk) => (cfToken = tk)" />
                </div>
                <div class="flex justify-end gap-2 mt-4">
                    <UButton
                        variant="outline"
                        color="neutral"
                        :disabled="!cfToken"
                        @click="
                            openDeactivateModal = false;
                            confirmDeactivate = '';
                            emit('update:close', false);
                        "
                    >
                        Cancel
                    </UButton>
                    <UButton
                        :disabled="confirmDeactivate !== 'DEACTIVATE' || !cfToken"
                        variant="solid"
                        color="error"
                        class="ml-2"
                        @click="onDeactivated"
                    >
                        Deactivate
                    </UButton>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
const openDeactivateModal = ref(false);

const props = defineProps<{
    open: boolean;
}>();

watch(
    () => props.open,
    (newVal) => {
        openDeactivateModal.value = newVal;
    }
);

const emit = defineEmits<{
    (e: "update:close", value: boolean): void;
}>();

const confirmDeactivate = ref<string>("");
const cfToken = ref<string>("");

const api = useApi();
const toast = useToast();

// Handle Deactivation
async function onDeactivated() {
    try {
        // send headers as the request config (third argument) and use the ref's value
        const response = await api.post("/me/deactivate", null, {
            headers: {
                "X-Turnstile-Token": cfToken.value,
            },
            withCredentials: true,
        });
        console.log(response);
        toast.add({
            title: "Account Deactivated",
            description: "Your account has been deactivated.",
            color: "success",
        });
        navigateTo("/", { replace: true });
    } catch (error) {
        console.log(error);
        toast.add({
            title: "Failed to deactivate profile",
            description: (error as { message: string }).message,
            color: "error",
        });
    } finally {
        openDeactivateModal.value = false;
        confirmDeactivate.value = "";
        cfToken.value = "";
        emit("update:close", false);
    }
}
</script>
