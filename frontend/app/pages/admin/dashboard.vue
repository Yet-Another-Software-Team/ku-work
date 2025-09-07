<template>
    <div>
        <section></section>
        <section class="flex flex-wrap w-full h-full place-content-center">
            <UserRequestComponent
                v-for="rid in 50"
                :key="rid"
                :request-id="rid"
                :data="multipleMockData[rid % 3] ?? mockData"
            />
        </section>
    </div>
</template>

<script setup lang="ts">
import { mockData, multipleMockData } from "~/data/mockData";

definePageMeta({
    layout: "admin",
});
</script>
