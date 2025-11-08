<template>
    <div ref="container"></div>
</template>

<script setup lang="ts">
const emit = defineEmits<{
    callback: [string];
}>();
const container = ref();
const config = useRuntimeConfig();
declare const turnstile: {
    render: (selector: string | HTMLElement, options: object) => string;
    remove: (widgetId: string) => void;
};
let widgetId: string = "";

onMounted(() => {
    if (import.meta.client) {
        if (!config.public.turnstileClientToken) {
            emit("callback", "disabled");
            return;
        }
        widgetId = turnstile.render(container.value, {
            sitekey: useRuntimeConfig().public.turnstileClientToken,
            callback: function (token: string) {
                emit("callback", token);
            },
        });
    }
});

onUnmounted(() => {
    if (import.meta.client && config.public.turnstileClientToken) turnstile.remove(widgetId);
});
</script>
