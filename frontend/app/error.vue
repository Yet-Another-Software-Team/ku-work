<template>
    <div class="h-screen w-full content-center gap-2">
        <NuxtImg src="/images/base.png" height="80" alt="ku work icon" class="block m-auto" />

        <h1 class="mt-1 text-8xl flex justify-center items-center">
            {{ error?.statusCode ?? "Error" }}
        </h1>
        <NuxtLink to="/" class="mt-3 flex justify-center items-center">
            Go back home <Icon name="ic:round-home" class="ml-1" />
        </NuxtLink>
    </div>
</template>

<script setup lang="ts">
import type { NuxtError } from "#app";

defineProps<{ error?: NuxtError }>();
</script>
