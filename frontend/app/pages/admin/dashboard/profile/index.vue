<template>
    <div class="my-5">
        <p class="text-xl">If you are seeing this message there is a problem with</p>
        <p class="text-5xl">YOU</p>
        <p class="text-xs">(maybe it is the page navigation)</p>
    </div>
</template>

<script setup lang="ts">
definePageMeta({
    layout: "admin",
});
navigateTo("/admin/dashboard");
</script>
