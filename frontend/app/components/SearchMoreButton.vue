<template>
    <div
        class="rounded-lg my-2 transition-transform duration-200"
        :class="showFilters ? 'border-1 p-2' : 'border-0'"
    >
        <!-- If filters are expanded -->
        <transition name="fade">
            <div v-if="showFilters" class="grid grid-cols-3 gap-6">
                <!-- Job Type -->
                <div>
                    <h3 class="font-bold mb-2">Job Type</h3>
                    <div class="space-y-1">
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Full Time</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Part Time</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Contract</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Casual</span>
                        </label>
                    </div>
                </div>

                <!-- Experience -->
                <div>
                    <h3 class="font-bold mb-2">Experience</h3>
                    <div class="space-y-1">
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">New Grad</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Junior</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Senior</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" /> <span class="ml-2">Manager</span>
                        </label>
                    </div>
                </div>

                <!-- Salary Range -->
                <div>
                    <h3 class="font-bold mb-2">Salary Range</h3>
                    <USlider v-model="sliderValues" :max="maxSalary" :step="1000" class="w-full" />
                    <div class="text-sm text-right pt-2">
                        {{ formatSalary(sliderValues[0] ?? 0) }}
                        -
                        {{ formatSalary(sliderValues[1] ?? 0) }} ฿
                    </div>
                </div>
            </div>
        </transition>

        <!-- Toggle button -->
        <div class="flex mt-4" :class="showFilters ? 'justify-end' : 'justify-start'">
            <UButton
                label="More"
                color="primary"
                variant="solid"
                class="h-full"
                :trailing-icon="showFilters ? 'i-lucide:chevron-up' : 'i-lucide:chevron-down'"
                @click="showFilters = !showFilters"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const maxSalary = 2000000;
const sliderValues = ref([0, 750000]);
const showFilters = ref(false);

function formatSalary(salary: number): string {
    return new Intl.NumberFormat("en", { notation: "compact" }).format(salary);
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
