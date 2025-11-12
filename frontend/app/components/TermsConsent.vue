<template>
    <div class="space-y-4">
        <div
            v-for="doc in docs"
            :key="doc.key"
            class="flex items-start gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-[#013B49]"
        >
            <input
                :id="`consent-${doc.key}`"
                :checked="localAcceptanceMap[doc.key] || false"
                type="checkbox"
                class="mt-[2px] h-4 w-4 cursor-pointer accent-primary-600"
                @click.prevent.stop="openViewer(doc.key)"
            />
            <div class="flex-1">
                <label
                    class="block cursor-pointer text-sm text-gray-900 dark:text-white"
                    :for="`consent-${doc.key}`"
                >
                    I have read and accept
                    <span class="font-semibold">{{ doc.title }}</span>
                    <span v-if="doc.required !== false" class="text-red-600">*</span>
                </label>
                <div class="mt-1">
                    <UButton
                        size="xs"
                        variant="ghost"
                        color="neutral"
                        class="px-0 text-primary-600 hover:underline"
                        :ui="{ base: 'px-1' }"
                        @click="openViewer(doc.key)"
                    >
                        View
                    </UButton>
                    <span v-if="loadStates[doc.key]?.error" class="ml-2 text-xs text-error">
                        Failed to load document.
                    </span>
                    <span
                        v-else-if="loadStates[doc.key]?.loading"
                        class="ml-2 text-xs text-gray-500"
                    >
                        Loading...
                    </span>
                </div>
            </div>
        </div>

        <div
            v-if="showComplianceNote"
            class="rounded-md border border-amber-300 bg-amber-50 p-3 text-amber-800 text-sm dark:bg-[#2b2400] dark:text-amber-200 dark:border-amber-700"
        >
            <p class="font-semibold mb-1">Important privacy information</p>
            <ul class="list-disc pl-5 space-y-1">
                <li>
                    We use an email service to send communications related to your account and
                    activities.
                </li>
                <li>We do not use analytics to track your behavior.</li>
                <li>
                    We may process your provided information with AI to assist decisions related to
                    the service (e.g., profile and content quality checks).
                </li>
                <li>
                    This KU-Work service is designed to comply with GDPR and Thailand PDPA
                    requirements.
                </li>
            </ul>
        </div>

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
                        {{ currentTitle }}
                    </h3>
                    <UButton variant="ghost" color="neutral" size="xs" @click="viewerOpen = false"
                        >Close</UButton
                    >
                </div>
            </template>
            <template #body>
                <div
                    ref="viewerBodyRef"
                    class="max-h-[60vh] overflow-y-auto"
                    @scroll="onViewerScroll"
                >
                    <pre
                        v-if="currentContent !== null"
                        class="whitespace-pre-wrap break-words text-sm text-gray-900 dark:text-white"
                        >{{ currentContent }}</pre
                    >
                    <div v-else class="text-sm text-gray-500">No content available.</div>
                </div>
            </template>
            <template #footer>
                <div class="flex justify-end gap-2">
                    <UButton variant="outline" color="neutral" @click="viewerOpen = false"
                        >Close</UButton
                    >
                    <UButton
                        color="primary"
                        :disabled="disabled || !scrolledToBottom"
                        @click="acceptCurrentAndClose"
                        >Accept</UButton
                    >
                </div>
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted, watch, nextTick } from "vue";

type DocDef = {
    key: string; // unique key for the document, e.g., 'ku-terms' or 'company-terms'
    title: string; // human readable title shown in UI
    src: string; // public path to the text file (e.g., '/terms/ku.txt'); must be accessible via HTTP
    required?: boolean; // whether acceptance is required for "allAccepted"; default true
};

const props = withDefaults(
    defineProps<{
        docs?: DocDef[];
        // v-model:accepted -> true only when all required docs accepted
        accepted?: boolean;
        // supply initial acceptance map if desired
        acceptanceMap?: Record<string, boolean>;
        // disable interactions
        disabled?: boolean;
        // show compliance note block
        showComplianceNote?: boolean;
    }>(),
    {
        docs: () => [],
        accepted: false,
        acceptanceMap: () => ({}),
        disabled: false,
        showComplianceNote: true,
    }
);

const emit = defineEmits<{
    (e: "update:accepted", value: boolean): void;
    (e: "update:acceptanceMap", value: Record<string, boolean>): void;
}>();

// local acceptance state by doc key
const localAcceptanceMap = reactive<Record<string, boolean>>({});

// document content and load states
const contents = reactive<Record<string, string | null>>({});
const loadStates = reactive<Record<string, { loading: boolean; error: boolean }>>({});

const requiredKeys = computed(() =>
    props.docs.filter((d) => d.required !== false).map((d) => d.key)
);

const allAccepted = computed(() => requiredKeys.value.every((k) => !!localAcceptanceMap[k]));

// modal viewer
const viewerOpen = ref(false);
const currentKey = ref<string | null>(null);
const viewerBodyRef = ref<HTMLElement | null>(null);
const scrolledToBottom = ref(false);

function onViewerScroll(e: Event) {
    const el = e.target as HTMLElement;
    scrolledToBottom.value = el.scrollTop + el.clientHeight >= el.scrollHeight - 8;
}

function resetScrollState() {
    scrolledToBottom.value = false;
}

watch(viewerOpen, async (open) => {
    if (open) {
        resetScrollState();
        await nextTick();
        const el = viewerBodyRef.value;
        if (el && el.scrollHeight <= el.clientHeight + 1) {
            scrolledToBottom.value = true;
        }
    }
});

const currentTitle = computed(() => {
    const doc = props.docs.find((d) => d.key === currentKey.value);
    return doc?.title ?? "";
});

const currentContent = computed(() =>
    currentKey.value ? (contents[currentKey.value] ?? null) : null
);

function initializeState() {
    // initialize acceptance map and content placeholders
    for (const d of props.docs) {
        localAcceptanceMap[d.key] = !!props.acceptanceMap[d.key];
        if (!(d.key in contents)) contents[d.key] = null;
        if (!(d.key in loadStates)) loadStates[d.key] = { loading: false, error: false };
    }
}

async function loadDocIfNeeded(key: string) {
    if (!key || loadStates[key]?.loading || contents[key] !== null) return;
    const doc = props.docs.find((d) => d.key === key);
    if (!doc) return;
    loadStates[key] = { loading: true, error: false };
    try {
        // Ensure client-side fetch
        const res = await fetch(doc.src, { method: "GET" });
        if (!res.ok) throw new Error(`Failed to fetch ${doc.src}: ${res.status}`);
        const text = await res.text();
        contents[key] = text;
        loadStates[key] = { loading: false, error: false };
    } catch {
        loadStates[key] = { loading: false, error: true };
        contents[key] = null;
        // non-fatal; UI shows error text
    }
}

function openViewer(key: string) {
    currentKey.value = key;
    resetScrollState();
    viewerOpen.value = true;
    void loadDocIfNeeded(key);
}

function acceptCurrentAndClose() {
    if (!currentKey.value || props.disabled || !scrolledToBottom.value) {
        if (!currentKey.value || props.disabled) viewerOpen.value = false;
        return;
    }
    localAcceptanceMap[currentKey.value] = true;
    emit("update:acceptanceMap", { ...localAcceptanceMap });
    emit("update:accepted", allAccepted.value);
    viewerOpen.value = false;
}

watch(allAccepted, (v) => {
    emit("update:accepted", v);
});

onMounted(() => {
    initializeState();
    // Preload all docs in background (optional)
    props.docs.forEach((d) => void loadDocIfNeeded(d.key));
});
</script>
