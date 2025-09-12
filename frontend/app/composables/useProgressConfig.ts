export interface ProgressConfig {
    color: string;
    size: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    position: 'top' | 'bottom';
    variant: 'default' | 'floating' | 'embedded';
    animation: 'carousel' | 'swing' | 'elastic' | 'pulse';
    showRequestCount: boolean;
    showLoadingText: boolean;
    loadingText: string;
}

const defaultConfig: ProgressConfig = {
    color: 'secondary',
    size: 'sm',
    position: 'top',
    variant: 'default',
    animation: 'carousel',
    showRequestCount: false,
    showLoadingText: false,
    loadingText: 'Loading...'
};

// Global configuration state
const globalConfig = ref<ProgressConfig>({ ...defaultConfig });

export const useProgressConfig = () => {
    const appConfig = useAppConfig();
    const colorMode = useColorMode();

    // Get theme-aware color based on current color mode
    const getThemeAwareColor = (baseColor: string): string => {
        if (colorMode.value === 'dark') {
            // Adjust colors for dark mode if needed
            switch (baseColor) {
                case 'secondary':
                    return 'secondary';
                case 'primary':
                    return 'primary';
                default:
                    return baseColor;
            }
        }
        return baseColor;
    };

    // Get configuration with theme awareness
    const getProgressConfig = (): ProgressConfig => {
        const config = { ...globalConfig.value };

        // Apply theme-aware color
        config.color = getThemeAwareColor(config.color);

        // Override with app config if available
        if (appConfig.progressBar) {
            Object.assign(config, appConfig.progressBar);
        }

        return config;
    };

    // Update global configuration
    const updateProgressConfig = (newConfig: Partial<ProgressConfig>): void => {
        Object.assign(globalConfig.value, newConfig);
    };

    // Reset to defaults
    const resetProgressConfig = (): void => {
        globalConfig.value = { ...defaultConfig };
    };

    // Preset configurations
    const presets = {
        minimal: {
            color: 'secondary',
            size: 'xs' as const,
            variant: 'default' as const,
            showRequestCount: false,
            showLoadingText: false
        },
        detailed: {
            color: 'secondary',
            size: 'sm' as const,
            variant: 'floating' as const,
            showRequestCount: true,
            showLoadingText: true,
            loadingText: 'Processing request...'
        },
        embedded: {
            color: 'primary',
            size: 'md' as const,
            variant: 'embedded' as const,
            showRequestCount: false,
            showLoadingText: true,
            loadingText: 'Loading content...'
        },
        elegant: {
            color: 'secondary',
            size: 'sm' as const,
            variant: 'floating' as const,
            animation: 'swing' as const,
            showRequestCount: false,
            showLoadingText: false
        }
    };

    // Apply preset
    const applyPreset = (presetName: keyof typeof presets): void => {
        updateProgressConfig(presets[presetName]);
    };

    // Reactive getters
    const currentConfig = computed(() => getProgressConfig());
    const isMinimalMode = computed(() =>
        !currentConfig.value.showRequestCount && !currentConfig.value.showLoadingText
    );

    return {
        // State
        currentConfig: readonly(currentConfig),
        isMinimalMode: readonly(isMinimalMode),

        // Methods
        getProgressConfig,
        updateProgressConfig,
        resetProgressConfig,
        applyPreset,
        getThemeAwareColor,

        // Presets
        presets: readonly(presets)
    };
};

// Export types
export type { ProgressConfig };
