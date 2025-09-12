<template>
    <Transition
        name="progress-bar"
        enter-active-class="transition-opacity duration-200 ease-in"
        leave-active-class="transition-opacity duration-300 ease-out"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
    >
        <div v-if="showProgress" :class="containerClass">
            <!-- Main Progress Bar using NuxtUI UProgress -->
            <UProgress
                :value="progressValue"
                :color="color"
                :size="size"
                :animation="animation"
                :class="progressClass"
            >
                <!-- Shimmer Effect Overlay -->
                <template #indicator>
                    <div
                        class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-shimmer"
                        :class="{ 'animate-pulse': progressValue < 90 }"
                    />
                </template>
            </UProgress>
        </div>
    </Transition>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
    showRequestCount?: boolean;
    showLoadingText?: boolean;
    loadingText?: string;
    color?: string;
    size?: "xs" | "sm" | "md" | "lg" | "xl";
    position?: "top" | "bottom";
    animation?: "carousel" | "swing" | "elastic" | "pulse";
    variant?: "default" | "floating" | "embedded";
}

const props = withDefaults(defineProps<Props>(), {
    showRequestCount: false,
    showLoadingText: false,
    loadingText: "Loading...",
    color: "primary",
    size: "sm",
    position: "top",
    animation: "carousel",
    variant: "default",
});

const { showProgress, progressValue } = useApiLoading();

// Container class based on variant and position
const containerClass = computed(() => {
    const baseClasses = [];

    if (props.variant === "floating") {
        baseClasses.push("fixed left-0 right-0 z-50 px-4");
        if (props.position === "bottom") {
            baseClasses.push("bottom-4");
        } else {
            baseClasses.push("top-4");
        }
    } else if (props.variant === "embedded") {
        baseClasses.push("relative w-full");
    } else {
        // default - fixed to edge
        baseClasses.push("fixed left-0 right-0 z-50");
        if (props.position === "bottom") {
            baseClasses.push("bottom-0");
        } else {
            baseClasses.push("top-0");
        }
    }

    return baseClasses.join(" ");
});

// Progress bar specific classes
const progressClass = computed(() => {
    const classes = ["transition-all duration-300 ease-out"];

    if (props.variant === "floating") {
        classes.push("rounded-full shadow-lg bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm");
    } else if (props.variant === "embedded") {
        classes.push("rounded-md");
    } else {
        classes.push("rounded-none");
    }

    return classes.join(" ");
});

</script>

<style scoped>
/* Custom shimmer animation */
@keyframes shimmer {
    0% {
        transform: translateX(-100%);
    }
    100% {
        transform: translateX(100%);
    }
}

.animate-shimmer {
    animation: shimmer 2s infinite;
}

/* Enhanced transitions for progress changes */
.progress-bar-enter-active,
.progress-bar-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.progress-bar-enter-from {
    opacity: 0;
    transform: translateY(-100%);
}

.progress-bar-leave-to {
    opacity: 0;
    transform: translateY(-100%);
}

/* Floating variant specific animations */
.progress-bar-enter-from.floating {
    transform: translateY(-20px) scale(0.95);
}

.progress-bar-leave-to.floating {
    transform: translateY(-20px) scale(0.95);
}

/* Ensure proper layering */
.z-50 {
    z-index: 9999;
}

/* Backdrop blur support */
@supports (backdrop-filter: blur(12px)) {
    .backdrop-blur-sm {
        backdrop-filter: blur(12px);
    }
}
</style>
