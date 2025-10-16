<template>
    <div
        class="flex items-center justify-between shadow-md rounded-xl px-6 py-4 border border-gray-300 dark:border-gray-700 hover:shadow-lg transition-all cursor-pointer bg-white dark:bg-gray-800"
        @click="navigateTo(`/jobs/${jobId}`)"
    >
        <!-- Left Section: Company Logo and Job Info -->
        <div class="flex items-center gap-4 flex-1 min-w-0">
            <!-- Company Logo -->
            <div
                class="w-16 h-16 rounded-lg border border-gray-300 dark:border-gray-600 flex items-center justify-center overflow-hidden flex-shrink-0 bg-white"
            >
                <img
                    v-if="companyLogo"
                    :src="companyLogo"
                    :alt="companyName"
                    class="object-contain w-full h-full p-2"
                />
                <Icon v-else name="ic:baseline-business" class="w-10 h-10 text-gray-400" />
            </div>

            <!-- Job Details -->
            <div class="flex flex-col gap-1 flex-1 min-w-0">
                <h3 class="text-lg font-semibold truncate text-gray-900 dark:text-white">
                    {{ position }}
                </h3>
                <p class="text-sm text-gray-600 dark:text-gray-400 truncate">
                    {{ companyName }}
                </p>
                <div class="flex flex-wrap items-center gap-2 mt-1">
                    <UBadge
                        v-if="jobType"
                        color="primary"
                        variant="subtle"
                        size="xs"
                        class="capitalize"
                    >
                        {{ formatJobType(jobType) }}
                    </UBadge>
                    <UBadge
                        v-if="experience"
                        color="neutral"
                        variant="subtle"
                        size="xs"
                        class="capitalize"
                    >
                        {{ formatExperience(experience) }}
                    </UBadge>
                </div>
            </div>
        </div>

        <!-- Right Section: Salary, Status, and Actions -->
        <div class="flex items-center gap-6 flex-shrink-0">
            <!-- Salary Range -->
            <div class="text-right hidden md:block">
                <p class="text-lg font-bold text-gray-900 dark:text-white">
                    {{ formatSalary(minSalary, maxSalary) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ appliedDate }}</p>
            </div>

            <!-- Status Badge -->
            <div class="flex items-center gap-3">
                <UBadge :color="statusColor" size="md" class="capitalize">
                    {{ status }}
                </UBadge>

                <!-- Actions Menu -->
                <UDropdown :items="menuItems" :popper="{ placement: 'bottom-end' }">
                    <UButton
                        icon="ic:baseline-more-vert"
                        variant="ghost"
                        color="neutral"
                        size="sm"
                        @click.stop
                    />
                </UDropdown>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
const props = defineProps<{
    jobId: number;
    position: string;
    companyName: string;
    companyLogo?: string;
    jobType?: string;
    experience?: string;
    minSalary: number;
    maxSalary: number;
    status: "pending" | "accepted" | "rejected";
    appliedDate: string;
}>();

const emit = defineEmits<{
    (e: "withdraw" | "viewDetails"): void;
}>();

const statusColor = computed(() => {
    switch (props.status) {
        case "pending":
            return "warning";
        case "accepted":
            return "primary";
        case "rejected":
            return "error";
        default:
            return "neutral";
    }
});

const menuItems = computed(() => {
    const items = [
        [
            {
                label: "View Job Details",
                icon: "ic:baseline-visibility",
                click: () => {
                    emit("viewDetails");
                    navigateTo(`/jobs/${props.jobId}`);
                },
            },
        ],
    ];

    if (props.status === "pending") {
        items.push([
            {
                label: "Withdraw Application",
                icon: "ic:baseline-cancel",
                click: () => emit("withdraw"),
            },
        ]);
    }

    return items;
});

function formatJobType(type: string): string {
    const typeMap: Record<string, string> = {
        fulltime: "Full Time",
        parttime: "Part Time",
        contract: "Contract",
        casual: "Casual",
        internship: "Internship",
    };
    return typeMap[type.toLowerCase()] || type;
}

function formatExperience(exp: string): string {
    const expMap: Record<string, string> = {
        newgrad: "New Grad",
        junior: "Junior",
        senior: "Senior",
        manager: "Manager",
        internship: "Internship",
    };
    return expMap[exp.toLowerCase()] || exp;
}

function formatSalary(min: number, max: number): string {
    if (min === 0 && max === 0) return "Negotiable";
    if (min === max) return `${min.toLocaleString()}k Baht`;
    return `${min.toLocaleString()}k - ${max.toLocaleString()}k Baht`;
}
</script>
