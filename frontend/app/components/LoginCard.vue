<template>
    <div class="relative translate-y-[-2em] mb-10">
        <div
            class="flex flex-row gap-2 h-[7em] w-[95vw] lg:w-[28em] absolute left-1/2 -translate-x-1/2 top-0"
        >
            <div
                class="hover:cursor-pointer transition-all duraion-150"
                :class="setTailwindClasses(false)"
                @click="selectRecruit"
            >
                <p class="font-bold px-5 py-5 text-2xl">KU Recruits</p>
            </div>
            <div
                class="hover:cursor-pointer transition-all duration-150"
                :class="setTailwindClasses(true)"
                @click="selectCompany"
            >
                <p class="font-bold px-5 py-5 text-2xl">Company</p>
            </div>
        </div>

        <div
            class="h-[30em] w-[95vw] lg:w-[28em] rounded-xl bg-white pt-10 text-black relative top-[3.5em] mx-auto z-10"
        >
            <div v-if="!isCompany" class="flex flex-col h-full w-full">
                <h2 class="text-xl font-semibold mx-auto mb-5">KU Students/Staffs Login</h2>
                <GoogleOauthButton class="mx-auto text-white" />
                <p class="mt-2 text-xs text-gray-500 text-center">
                    By continuing, you acknowledge our
                    <a
                        href="#"
                        class="text-primary-600 underline hover:text-primary-800"
                        @click.prevent="openNoticeViewer"
                        >Google Sign‑In Notice</a
                    >.
                </p>

                <UModal
                    v-model:open="noticeViewerOpen"
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
                                Google Sign‑In Notice
                            </h3>
                            <UButton
                                variant="ghost"
                                color="neutral"
                                size="xs"
                                @click="noticeViewerOpen = false"
                                >Close</UButton
                            >
                        </div>
                    </template>
                    <template #body>
                        <div class="max-h-[60vh] overflow-y-auto">
                            <div v-if="noticeLoading" class="text-sm text-gray-500">Loading...</div>
                            <div v-else-if="noticeError" class="text-sm text-error">
                                Failed to load document.
                            </div>
                            <pre
                                v-else-if="noticeDocText !== null"
                                class="whitespace-pre-wrap break-words text-sm text-gray-900 dark:text-white"
                                >{{ noticeDocText }}</pre
                            >
                            <div v-else class="text-sm text-gray-500">No content available.</div>
                        </div>
                    </template>
                    <template #footer>
                        <div class="flex justify-end gap-2">
                            <UButton
                                variant="outline"
                                color="neutral"
                                @click="noticeViewerOpen = false"
                                >Close</UButton
                            >
                        </div>
                    </template>
                </UModal>
            </div>
            <div v-else class="flex flex-col h-full w-full">
                <h2 class="text-xl font-semibold mx-auto mb-5">Company Login</h2>
                <CompanyLoginField />
                <p class="text-center w-full mt-auto mb-2">
                    Don't have an account?
                    <a
                        class="text-primary-600 font-semibold underline hover:text-primary-800"
                        href="/register/company"
                    >
                        Sign Up
                    </a>
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
const isCompany = ref(false);

const noticeViewerOpen = ref(false);
const noticeDocText = ref(null);
const noticeLoading = ref(false);
const noticeError = ref(false);

async function loadNoticeDoc() {
    if (noticeDocText.value !== null || noticeLoading.value) return;
    noticeLoading.value = true;
    noticeError.value = false;
    try {
        const res = await fetch("/terms/google_oauth_notice.txt", { method: "GET" });
        if (!res.ok) throw new Error(`Failed to fetch (${res.status})`);
        noticeDocText.value = await res.text();
    } catch {
        noticeError.value = true;
        noticeDocText.value = null;
    } finally {
        noticeLoading.value = false;
    }
}

function openNoticeViewer() {
    noticeViewerOpen.value = true;
    void loadNoticeDoc();
}

function setTailwindClasses(activeCondition) {
    if (isCompany.value == activeCondition) {
        return "bg-primary-200 flex flex-col rounded-3xl w-1/2 text-primary-800 hover:bg-primary-300";
    } else {
        return "bg-gray-200 flex flex-col rounded-3xl w-1/2 text-gray-500 hover:bg-gray-300";
    }
}

function selectCompany() {
    isCompany.value = true;
}

function selectRecruit() {
    isCompany.value = false;
}
</script>
