import { ref, computed } from "vue";

interface RequestInfo {
    id: string;
    url: string;
    method: string;
    timestamp: number;
}

// Global state - shared across all instances
const activeRequests = ref<Map<string, RequestInfo>>(new Map());
const showProgress = ref(false);
const progressValue = ref(0);
let progressTimer: NodeJS.Timeout | null = null;

export const useApiLoading = () => {
    // Generate unique request ID
    const generateRequestId = (): string => {
        return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    };

    // Computed values
    const isLoading = computed(() => activeRequests.value.size > 0);
    const requestCount = computed(() => activeRequests.value.size);

    // Start tracking a request
    const startRequest = (url: string, method: string): string => {
        const id = generateRequestId();
        const requestInfo: RequestInfo = {
            id,
            url,
            method: method.toUpperCase(),
            timestamp: Date.now(),
        };

        activeRequests.value.set(id, requestInfo);

        // Start progress bar animation
        if (!showProgress.value) {
            startProgress();
        }

        return id;
    };

    // Stop tracking a request
    const endRequest = (id: string): void => {
        activeRequests.value.delete(id);

        // Stop progress bar if no active requests
        if (activeRequests.value.size === 0) {
            completeProgress();
        }
    };

    // Start progress bar animation
    const startProgress = (): void => {
        showProgress.value = true;
        progressValue.value = 0;

        // Simulate progress animation
        const animate = () => {
            if (progressValue.value < 90 && activeRequests.value.size > 0) {
                // Slow down as we approach 90%
                const increment = progressValue.value < 30 ? 10 : progressValue.value < 60 ? 5 : 2;
                progressValue.value = Math.min(90, progressValue.value + increment);
                progressTimer = setTimeout(animate, 200);
            }
        };

        progressTimer = setTimeout(animate, 100);
    };

    // Complete progress bar
    const completeProgress = (): void => {
        if (progressTimer) {
            clearTimeout(progressTimer);
            progressTimer = null;
        }

        // Quickly complete to 100%
        progressValue.value = 100;

        // Hide progress bar after brief delay
        setTimeout(() => {
            showProgress.value = false;
            progressValue.value = 0;
        }, 200);
    };

    // Force complete (for error cases)
    const forceComplete = (): void => {
        activeRequests.value.clear();
        completeProgress();
    };

    // Get current active requests info (for debugging)
    const getActiveRequests = (): RequestInfo[] => {
        return Array.from(activeRequests.value.values());
    };

    // Clear old requests (cleanup utility)
    const clearStaleRequests = (maxAge: number = 30000): void => {
        const now = Date.now();
        const staleIds: string[] = [];

        activeRequests.value.forEach((request, id) => {
            if (now - request.timestamp > maxAge) {
                staleIds.push(id);
            }
        });

        staleIds.forEach((id) => {
            activeRequests.value.delete(id);
        });

        if (activeRequests.value.size === 0 && showProgress.value) {
            completeProgress();
        }
    };

    // Auto cleanup stale requests every 30 seconds
    if (import.meta.client) {
        setInterval(clearStaleRequests, 30000);
    }

    // Theme-aware utilities
    const getProgressColor = () => {
        // You can extend this to read from app config or theme context
        return "secondary";
    };

    const getProgressHeight = () => {
        return "thin"; // or read from app config
    };

    return {
        // State
        isLoading: readonly(isLoading),
        requestCount: readonly(requestCount),
        showProgress: readonly(showProgress),
        progressValue: readonly(progressValue),

        // Methods
        startRequest,
        endRequest,
        forceComplete,
        getActiveRequests,
        clearStaleRequests,

        // Theme utilities
        getProgressColor,
        getProgressHeight,
    };
};

// Export types for external use
export type { RequestInfo };
