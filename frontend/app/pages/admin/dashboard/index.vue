<template>
    <div>
        <section class="h-[3em] overflow-hidden border-b-1 my-5">
            <div class="flex flex-row gap-2 h-[6em] w-[28em] left-0 top-0">
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="setTailwindClasses(false)"
                    @click="selectRecruit"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Students</p>
                </div>
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="setTailwindClasses(true)"
                    @click="selectCompany"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Company</p>
                </div>
            </div>
        </section>
        <section>
            <div class="flex justify-between items-center">
                <h1 class="text-2xl font-semibold mb-2">
                    {{ isCompany ? `${companyRequests.length} Posts` : `${totalRequests} Applicants` }}
                </h1>
                <div class="flex gap-5 items-center">
                    <h1 class="text-2xl font-semibold mb-2">Sort by:</h1>
                    <USelectMenu
                        v-model="selectedValue"
                        value-key="id"
                        :items="items"
                        placement="bottom-end"
                        class="w-[10em]"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />
        </section>
        <section class="flex flex-wrap w-full h-full place-content-center">
            <!-- User acc req -->
            <template v-if="!isCompany">
                <StudentRequestComponent
                    v-for="rid in totalRequests"
                    :key="rid"
                    :request-id="rid"
                    :data="multipleMockUserData[rid % 3] ?? mockUserData"
                />
            </template>
            <!-- Company acc req -->
            <template v-else>
                <template v-if="companyRequests.length">
                    <RequestedCompanyProfileCard
                        v-for="job in companyRequests"
                        :key="job.id"
                        :data="job"
                        @resolved="onCompanyRequestResolved"
                    />
                </template>
                <template v-else>
                    <h1 class="h-full justify-center items-center">No pending company posts</h1>
                </template>
            </template>
        </section>
    </div>
</template>

<script setup lang="ts">
import { mockUserData, multipleMockUserData } from "~/data/mockData";
import type { JobPost } from "~/data/mockData";
import RequestedCompanyProfileCard from "~/components/RequestedCompanyProfileCard.vue";
import { useApi } from "~/composables/useApi";

definePageMeta({
    layout: "admin",
});

const totalRequests = 50;
const isCompany = ref(false);
const companyRequests = ref<JobPost[]>([]);
const loggedOnce = ref(false);
const { get, showErrorToast } = useApi();

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
const selectedValue = ref("item1");

function setTailwindClasses(activeCondition: boolean) {
    if (isCompany.value == activeCondition) {
        return "bg-primary-200 flex flex-col border-1 rounded-3xl w-1/2 text-primary-800 hover:bg-primary-300";
    } else {
        return "bg-gray-200 flex flex-col border-1 rounded-3xl w-1/2 text-gray-500 hover:bg-gray-300";
    }
}

function selectCompany() {
    isCompany.value = true;
    fetchPendingCompanyPosts();
}

function selectRecruit() {
    isCompany.value = false;
}

async function fetchPendingCompanyPosts() {
    try {
        const res = await get<{ jobs: JobPost[] }>("/jobs", {
            params: { approvalStatus: "pending", limit: 64 },
        });
        if (!loggedOnce.value) {
            // One-time log to help verify payload during troubleshooting
            console.log("[Admin Dashboard] Pending jobs response:", res.data);
            loggedOnce.value = true;
        }
        companyRequests.value = (res.data as any)?.jobs ?? [];
    } catch (e: any) {
        showErrorToast(e, "Failed to load company posts");
    }
}

function onCompanyRequestResolved(jobId: number) {
    companyRequests.value = companyRequests.value.filter((j) => j.id !== jobId);
}
</script>
