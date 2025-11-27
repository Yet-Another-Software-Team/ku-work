<template>
    <div>
        <UApp>
            <!-- Global API Progress Bar with configurable theming -->
            <UApiProgressBar
                :color="progressConfig.color"
                :show-request-count="progressConfig.showRequestCount"
                :show-loading-text="progressConfig.showLoadingText"
                :loading-text="progressConfig.loadingText"
                :variant="progressConfig.variant"
                :size="progressConfig.size"
                :position="progressConfig.position"
                :animation="progressConfig.animation"
            />
            <ExternalRedirect />
            <NuxtLayout>
                <NuxtPage />
            </NuxtLayout>
        </UApp>
    </div>
</template>

<script setup lang="ts">
const { currentConfig: progressConfig } = useProgressConfig();
</script>
