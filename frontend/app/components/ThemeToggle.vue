<template>
    <UButton
        :icon="colorMode.value === 'dark' ? 'ic:round-light-mode' : 'ic:round-dark-mode'"
        class="cursor-pointer"
        @click="toggleDark"
    />
</template>

<script setup lang="ts">
const colorMode = useColorMode();
const toggleDark = () => {
    colorMode.preference = colorMode.value === "dark" ? "light" : "dark";
};
</script>
