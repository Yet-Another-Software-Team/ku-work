<template>
    <UModal
        v-model:open="open"
        :ui="{
            overlay: 'fixed inset-0 bg-black/60',
            content: 'max-w-lg w-full',
        }"
    >
        <template #content>
            <UCard class="bg-white dark:bg-gray-900">
                <div class="flex items-start gap-3">
                    <div
                        class="rounded-full bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 p-2"
                    >
                        <Icon name="material-symbols:warning-rounded" class="size-6" />
                    </div>
                    <div class="space-y-2">
                        <p class="text-lg font-semibold text-gray-900 dark:text-white">
                            You are leaving this site
                        </p>
                        <p class="text-sm text-gray-600 dark:text-gray-300">
                            You are about to open
                            <span class="font-semibold text-primary-700 dark:text-primary-300">
                                {{ externalHost }}
                            </span>
                            . Make sure you trust this website before continuing.
                        </p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 break-all">
                            {{ pendingUrl }}
                        </p>
                    </div>
                </div>

                <template #footer>
                    <div class="flex justify-end gap-2">
                        <UButton variant="ghost" color="neutral" @click="cancelRedirect">
                            Stay here
                        </UButton>
                        <UButton color="primary" @click="proceedRedirect">Continue</UButton>
                    </div>
                </template>
            </UCard>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

const open = ref(false);
const pendingUrl = ref<string>("");
const pendingTarget = ref<string>("_self");

const externalHost = computed(() => {
    if (!pendingUrl.value) return "";
    try {
        return new URL(pendingUrl.value).host;
    } catch {
        return "";
    }
});

function isExternalUrl(href: string): { external: boolean; fullUrl: string } {
    try {
        const url = new URL(href, window.location.href);
        const isHttpProtocol = url.protocol === "http:" || url.protocol === "https:";
        const isSameHost = url.host === window.location.host;

        return { external: isHttpProtocol && !isSameHost, fullUrl: url.href };
    } catch {
        return { external: false, fullUrl: href };
    }
}

function handleClick(event: MouseEvent) {
    if (event.defaultPrevented) return;
    const targetElement = event.target as HTMLElement | null;
    const anchor = targetElement?.closest("a");
    if (!anchor) return;

    const href = anchor.getAttribute("href");
    if (!href || href.startsWith("#")) return;

    const rel = anchor.getAttribute("rel") || "";
    if (rel.includes("no-external-guard")) return;

    const { external, fullUrl } = isExternalUrl(href);
    if (!external) return;

    event.preventDefault();
    pendingUrl.value = fullUrl;
    pendingTarget.value = anchor.getAttribute("target") || "_self";
    open.value = true;
}

function resetState() {
    open.value = false;
    pendingUrl.value = "";
    pendingTarget.value = "_self";
}

function proceedRedirect() {
    if (!pendingUrl.value) return;

    const target = pendingTarget.value || "_self";
    if (target === "_self") {
        window.location.assign(pendingUrl.value);
    } else {
        window.open(pendingUrl.value, target, "noopener");
    }

    resetState();
}

function cancelRedirect() {
    resetState();
}

onMounted(() => {
    document.addEventListener("click", handleClick, { capture: true });
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleClick, { capture: true });
});
</script>
