<template>
    <div
        class="rounded-lg my-2 transition-transform duration-200"
        :class="
            showFilters ? 'border-1 border-primary p-2 bg-[#fdfdfd] dark:bg-[#013B49]' : 'border-0'
        "
    >
        <!-- If filters are expanded -->
        <transition name="fade">
            <div v-if="showFilters" class="grid grid-cols-3 gap-6">
                <!-- Job Type -->
                <div>
                    <h3 class="font-bold mb-2">Job Type</h3>
                    <div class="space-y-1">
                        <UCheckboxGroup
                            v-model="jobTypeValue"
                            :items="jobTypeItems"
                            size="lg"
                            :ui="{
                                base: 'ring-2 dark:ring-gray-400',
                            }"
                        />
                    </div>
                </div>

                <!-- Experience -->
                <div>
                    <h3 class="font-bold mb-2">Experience</h3>
                    <div class="space-y-1">
                        <UCheckboxGroup
                            v-model="expTypeValue"
                            :items="expTypeItems"
                            size="lg"
                            :ui="{
                                base: 'ring-2 dark:ring-gray-400',
                            }"
                        />
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
import type { CheckboxGroupItem, CheckboxGroupValue } from "@nuxt/ui";

const maxSalary = 2000000;
const sliderValues = ref([0, 750000]);
const showFilters = ref(false);

const jobTypeItems = ref<CheckboxGroupItem[]>(["Full Time", "Part Time", "Contract", "Casual"]);
const jobTypeValue = ref<CheckboxGroupValue[]>([]);

const expTypeItems = ref<CheckboxGroupItem[]>(["Internship", "Freelance", "Temporary"]);
const expTypeValue = ref<CheckboxGroupValue[]>([]);

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
