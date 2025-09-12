<template>
    <Transition
        name="progress-bar"
        enter-active-class="transition-opacity duration-200 ease-in"
        leave-active-class="transition-opacity duration-300 ease-out"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
    >
        <div v-if="showProgress" class="fixed top-0 left-0 right-0 z-50 h-1 bg-transparent">
            <!-- Progress Bar Background -->
            <div class="h-full bg-gray-200 dark:bg-gray-800 shadow-sm">
                <!-- Animated Progress Fill -->
                <div
                    class="h-full bg-gradient-to-r from-secondary-500 to-secondary-600 transition-all duration-300 ease-out relative overflow-hidden"
                    :style="{ width: `${progressValue}%` }"
                >
                    <!-- Shimmer Effect -->
                    <div
                        class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-shimmer"
                        :class="{ 'animate-pulse': progressValue < 90 }"
                    />
                </div>
            </div>

            <!-- Optional: Request Counter (for debugging) -->
            <div
                v-if="showRequestCount && requestCount > 0"
                class="absolute top-2 right-4 text-xs bg-secondary-100 text-secondary-800 dark:bg-secondary-800 dark:text-secondary-200 px-2 py-1 rounded-full shadow-sm"
            >
                {{ requestCount }} active request{{ requestCount !== 1 ? "s" : "" }}
            </div>
        </div>
    </Transition>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
    showRequestCount?: boolean;
    color?: "primary" | "secondary" | "green" | "red" | "yellow" | "blue";
    height?: "thin" | "normal" | "thick";
    position?: "top" | "bottom";
}

const props = withDefaults(defineProps<Props>(), {
    showRequestCount: false,
    color: "secondary",
    height: "thin",
    position: "top",
});

const { showProgress, progressValue, requestCount } = useApiLoading();

// Compute classes based on props
computed(() => {
    const classes = [];

    // Height classes
    switch (props.height) {
        case "thin":
            classes.push("h-0.5");
            break;
        case "thick":
            classes.push("h-2");
            break;
        default:
            classes.push("h-1");
    }

    // Position classes
    if (props.position === "bottom") {
        classes.push("bottom-0");
    } else {
        classes.push("top-0");
    }

    return classes.join(" ");
});

computed(() => {
    const baseClasses = "h-full transition-all duration-300 ease-out relative overflow-hidden";

    // Color classes using NuxtUI color tokens
    switch (props.color) {
        case "primary":
            return `${baseClasses} bg-gradient-to-r from-primary-500 to-primary-600`;
        case "green":
            return `${baseClasses} bg-gradient-to-r from-green-500 to-green-600`;
        case "red":
            return `${baseClasses} bg-gradient-to-r from-red-500 to-red-600`;
        case "yellow":
            return `${baseClasses} bg-gradient-to-r from-yellow-500 to-yellow-600`;
        case "blue":
            return `${baseClasses} bg-gradient-to-r from-blue-500 to-blue-600`;
        default: // secondary
            return `${baseClasses} bg-gradient-to-r from-secondary-500 to-secondary-600`;
    }
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

/* Smooth transitions for progress changes */
.progress-bar-enter-active,
.progress-bar-leave-active {
    transition: all 0.3s ease;
}

.progress-bar-enter-from {
    opacity: 0;
    transform: scaleY(0);
}

.progress-bar-leave-to {
    opacity: 0;
    transform: scaleY(0);
}

/* Ensure proper layering */
.fixed {
    z-index: 9999;
}
</style>
