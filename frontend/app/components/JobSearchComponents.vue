<template>
    <div>
        <div class="flex flex-wrap sm:flex-nowrap gap-3">
            <!-- Search Menu -->
            <UInput
                v-model="searchValue"
                color="primary"
                highlight
                placeholder="Enter Keyword"
                icon="material-symbols:search-rounded"
                class="w-full"
            >
                <template v-if="searchValue?.length" #trailing>
                    <UButton
                        color="neutral"
                        variant="link"
                        icon="iconoir:xmark"
                        aria-label="Clear input"
                        @click="searchValue = ''"
                    />
                </template>
            </UInput>
            <!-- Location Picker -->
            <USelectMenu
                v-model="selectedValue"
                color="primary"
                highlight
                placeholder="Location"
                value-key="id"
                :items="items"
                class="w-[15em]"
                icon="material-symbols:location-on-outline-rounded"
            />
        </div>

        <!-- More options -->
        <div>
            <SearchMoreButton />
        </div>
    </div>
</template>

<script setup lang="ts">
import { UButton } from "#components";
import SearchMoreButton from "./SearchMoreButton.vue";

const emit = defineEmits<{
    (e: "update:search", value: string): void;
    (e: "update:location", value: string | null): void;
}>();

const props = defineProps<{
    locations?: string[];
}>();

const items = computed(() => {
    const unique = [...new Set(props.locations)];
    return unique.map((loc) => ({
        label: loc,
        id: loc.toLowerCase().replace(/\s+/g, "-"),
    }));
});

const selectedValue = ref("");
const searchValue = ref("");

watch(searchValue, (val) => emit("update:search", val));
watch(selectedValue, (val) => emit("update:location", val));
</script>
