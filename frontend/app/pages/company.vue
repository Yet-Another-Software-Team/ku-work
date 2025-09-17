<template>
    <div>
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mb-5">Dashboards</h1>
        <div class="flex flex-wrap gap-10">
            <JobCardCompany
                v-for="job in data"
                :key="job.id"
                class="h-[18em] w-full lg:w-[25em] drop-shadow-md"
                :job-i-d="job.id.toString()"
                :open="job.open"
                :position="job.position"
                :accepted="job.accepted"
                :rejected="job.rejected"
                :pending="job.pending"
                @update:open="(value: boolean) => updateJobOpen(job.id, value)"
            />
        </div>
        <div class="bg-primary p-2 rounded-full size-[4em] fixed bottom-5 right-[6vw]">
            <Icon
                name="ic:baseline-plus"
                size="4em"
                mode="svg"
                class="absolute top-0 bottom-0 left-0 right-0 text-white"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
definePageMeta({
    layout: "viewer",
});

const data = [
    {
        id: 1,
        position: "Software Quality Assurance Engineer",
        accepted: 10,
        rejected: 5,
        pending: 3,
        open: true,
    },
    {
        id: 2,
        position: "Software Engineer",
        accepted: 15,
        rejected: 7,
        pending: 2,
        open: false,
    },
    {
        id: 3,
        position: "Product Manager",
        accepted: 8,
        rejected: 4,
        pending: 1,
        open: true,
    },
    {
        id: 4,
        position: "Data Scientist",
        accepted: 12,
        rejected: 6,
        pending: 4,
        open: true,
    },
    {
        id: 5,
        position: "Security Personnel",
        accepted: 2,
        rejected: 199238,
        pending: 0,
        open: false,
    },
];

const updateJobOpen = (id: number, value: boolean) => {
    const job = data.find((job) => job.id === id);
    if (job) {
        job.open = value;
    }
};
</script>
