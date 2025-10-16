<template>
    <div>
        <!-- Selector -->
        <section class="h-[3em] overflow-hidden border-b-1 my-5">
            <div v-if="isLoading">
                <USkeleton class="h-[6em] w-[28em] left-0 top-0" />
            </div>
            <div v-else class="flex flex-row gap-2 h-[6em] w-[28em] left-0 top-0">
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="setTailwindClasses(false)"
                    @click="selectStudent"
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
        <!-- Total requests and sort -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[3em] w-full left-0 top-0 mb-5" />
            </div>
            <div v-else class="flex justify-between">
                <h1 class="text-2xl font-semibold mb-2">{{ totalRequests }} Applicants</h1>
                <div class="flex gap-5">
                    <!-- TODO: Implement sorting -->
                    <h1 class="text-2xl font-semibold mb-2">Sort by:</h1>
                    <USelectMenu
                        v-model="selectSortOption"
                        value-key="id"
                        :items="sortOptions"
                        placement="bottom-end"
                        class="w-[10em]"
                    />
                </div>
            </div>
            <hr class="w-full my-5" />
        </section>
        <!-- Request list -->
        <section class="flex flex-wrap w-full h-full place-content-start gap-5">
            <!-- User acc req -->
            <template v-if="!isCompany && studentData">
                <StudentRequestDashboardComponent
                    v-for="(data, index) in studentData"
                    :key="index"
                    :request-id="data.id"
                    :data="data"
                    @student-approval-status="selectStudent"
                />
            </template>
            <!-- Company acc req -->
            <template v-else>
                <h1 class="h-full justify-center items-center">Not yet implement</h1>
            </template>
        </section>
    </div>
</template>

<script setup lang="ts">
import type { Profile } from "~/data/mockData";

definePageMeta({
    layout: "admin",
});

const isLoading = ref(true);

const totalRequests = ref(0);
const isCompany = ref(false);

const api = useApi();
const studentData = ref<Profile>();

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);

const selectSortOption = ref("latest");

onMounted(async () => {
    isLoading.value = true;

    try {
        await selectStudent();
    } catch (error) {
        console.error("Error fetching student data:", error);
    } finally {
        isLoading.value = false;
    }
});

function setTailwindClasses(activeCondition: boolean) {
    if (isCompany.value == activeCondition) {
        return "bg-primary-200 flex flex-col border-1 rounded-3xl w-1/2 text-primary-800 hover:bg-primary-300";
    } else {
        return "bg-gray-200 flex flex-col border-1 rounded-3xl w-1/2 text-gray-500 hover:bg-gray-300";
    }
}

function selectCompany() {
    isCompany.value = true;
    totalRequests.value = 0;
}

const studentOffset = ref(0);
const studentLimit = ref(10);

async function selectStudent() {
    isCompany.value = false;
    try {
        const response = await api.get("/students", {
            params: {
                limit: studentLimit.value,
                offset: studentOffset.value,
                approvalStatus: "pending",
            },
            withCredentials: true,
        });
        studentData.value = response.data as Profile;
        totalRequests.value = response.data.length;
    } catch (error) {
        console.error("Error fetching student data:", error);
    }
}
</script>
