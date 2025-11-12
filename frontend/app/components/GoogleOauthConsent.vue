<template>
    <div class="flex flex-col gap-3">
        <div
            class="rounded-md border border-gray-200 dark:border-gray-700 p-3 bg-white dark:bg-[#013B49]"
        >
            <label class="flex items-start gap-3 cursor-pointer">
                <input
                    :checked="accepted"
                    type="checkbox"
                    class="mt-[2px] h-4 w-4 accent-primary-600"
                    @change="onToggle"
                />
                <span class="text-sm text-gray-900 dark:text-white">
                    I have read and consent to Google Sign-In data sharing for authentication
                    <span class="text-red-600">*</span>
                    <UButton
                        size="xs"
                        variant="ghost"
                        color="neutral"
                        class="ml-2 px-0 text-primary-600 hover:underline"
                        :ui="{ base: 'px-1' }"
                        @click="openViewer"
                    >
                        View details
                    </UButton>
                    <span v-if="loadingDoc" class="ml-2 text-xs text-gray-500">Loading…</span>
                    <span v-else-if="docError" class="ml-2 text-xs text-error">Failed to load</span>
                </span>
            </label>
            <p class="mt-2 text-xs text-gray-600 dark:text-gray-300">
                - We use an email service to send transactional/service communications.
                <br />
                - We do not do behavior-tracking analytics.
                <br />
                - We may use AI to assist certain decisions with appropriate human oversight.
            </p>
        </div>

        <UButton
            class="size-fit text-xl rounded-md px-10 gap-2 font-medium py-3 bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
            icon="cib:google"
            :label="buttonLabel"
            :loading="isLoggingIn"
            :disabled="!accepted || isLoggingIn"
            @click="continueWithGoogle"
        />

        <UModal
            v-model:open="viewerOpen"
            :ui="{
                overlay: 'fixed inset-0 bg-black/50',
                content: 'w-full max-w-3xl',
                title: 'text-xl font-semibold text-primary-800 dark:text-primary',
            }"
            :dismissible="true"
        >
            <template #header>
                <div class="flex items-center justify-between w-full">
                    <h3 class="text-lg font-semibold text-primary-800 dark:text-primary">
                        Google Sign‑In Notice and Consent
                    </h3>
                    <UButton variant="ghost" color="neutral" size="xs" @click="viewerOpen = false"
                        >Close</UButton
                    >
                </div>
            </template>
            <template #body>
                <div class="max-h-[60vh] overflow-y-auto">
                    <pre
                        v-if="docText !== null"
                        class="whitespace-pre-wrap break-words text-sm text-gray-900 dark:text-white"
                        >{{ docText }}</pre
                    >
                    <div v-else class="text-sm text-gray-500">No content available.</div>
                </div>
            </template>
            <template #footer>
                <div class="flex justify-end gap-2">
                    <UButton variant="outline" color="neutral" @click="viewerOpen = false"
                        >Close</UButton
                    >
                    <UButton color="primary" :disabled="accepted" @click="acceptAndClose"
                        >Accept</UButton
                    >
                </div>
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { googleAuthCodeLogin } from "vue3-google-login";
import { useApi, type LoginResponse } from "~/composables/useApi";

const props = withDefaults(
    defineProps<{
        noticeSrc?: string;
        buttonLabel?: string;
    }>(),
    {
        noticeSrc: "/terms/google_oauth_notice.txt",
        buttonLabel: "Continue with Google",
    }
);

const emit = defineEmits<{
    (
        e: "success",
        payload: { status: number; role: string; userId?: string; isRegistered: boolean }
    ): void;
}>();

const api = useApi();
const toast = useToast();

const isLoggingIn = ref(false);
const accepted = ref(false);

const viewerOpen = ref(false);
const docText = ref<string | null>(null);
const loadingDoc = ref(false);
const docError = ref(false);

async function loadDoc() {
    if (docText.value !== null || loadingDoc.value) return;
    loadingDoc.value = true;
    docError.value = false;
    try {
        const res = await fetch(props.noticeSrc, { method: "GET" });
        if (!res.ok) throw new Error(`Failed to fetch (${res.status})`);
        docText.value = await res.text();
    } catch {
        docError.value = true;
        docText.value = null;
    } finally {
        loadingDoc.value = false;
    }
}

function openViewer() {
    viewerOpen.value = true;
    void loadDoc();
}

function acceptAndClose() {
    accepted.value = true;
    viewerOpen.value = false;
}

function onToggle(ev: Event) {
    accepted.value = (ev.target as HTMLInputElement).checked;
}

async function continueWithGoogle() {
    if (!accepted.value || isLoggingIn.value) {
        if (!accepted.value) {
            toast.add({
                title: "Consent required",
                description: "Please review and accept the Google Sign‑In notice to continue.",
                color: "warning",
            });
        }
        return;
    }

    isLoggingIn.value = true;

    interface OauthLoginResponse extends LoginResponse {
        isRegistered: boolean;
    }

    try {
        const oauthResponse = await googleAuthCodeLogin();

        try {
            const response = await api.post<OauthLoginResponse>(
                "/auth/google/login",
                { code: oauthResponse.code },
                { withCredentials: true }
            );

            localStorage.setItem("token", response.data.token);
            localStorage.setItem("username", response.data.username);
            localStorage.setItem("isRegistered", response.data.isRegistered.toString());
            if (response.data.userId) {
                localStorage.setItem("userId", response.data.userId);
            }
            const role = response.data.role;
            localStorage.setItem("role", role);

            // Navigate based on status (201 => new account)
            if (response.status === 201) {
                navigateTo("/register/student", { replace: true });
            } else {
                navigateTo("/jobs", { replace: true });
            }

            emit("success", {
                status: response.status,
                role,
                userId: response.data.userId,
                isRegistered: response.data.isRegistered,
            });
        } catch (error: unknown) {
            let description = "Failed to log in with Google.";
            const e = error as { status?: number; message?: string };

            if (e?.status === 401) {
                description = "Google account not authorized. Please use a valid account.";
            } else if (e?.status === 500) {
                description = "Server error. Please try again later.";
            } else if (e?.message) {
                description = e.message;
            }

            toast.add({
                title: "Login Failed",
                description,
                color: "error",
            });
        }
    } catch {
        // User closed popup or Google flow failed before code exchange
        toast.add({
            title: "Login Cancelled",
            description: "Google Sign‑In was cancelled or failed to start.",
            color: "warning",
        });
    } finally {
        isLoggingIn.value = false;
    }
}

onMounted(() => {
    // Optional: prefetch doc in background
    void loadDoc();
});
</script>
