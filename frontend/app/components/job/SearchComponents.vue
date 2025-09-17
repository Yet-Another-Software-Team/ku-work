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

const items = ref([
    {
        label: "Item1",
        id: "item1",
        icon: "eos-icons:rotating-gear",
    },
    {
        label: "Item2",
        id: "item2",
        icon: "eos-icons:rotating-gear",
    },
    {
        label: "Item3",
        id: "item3",
        icon: "eos-icons:rotating-gear",
    },
]);

const selectedValue = ref();
const searchValue = ref("");
</script>
