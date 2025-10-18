<template>
    <div class="relative">
        <StudentRequestProfileCard :request-id="id" />
    </div>
</template>

<script setup lang="ts">
definePageMeta({
    layout: "admin",
});

const route = useRoute();

const id = ref<string>("");

onMounted(() => {
    if (route.params.id) {
        id.value = String(route.params.id);
    }
});
</script>
