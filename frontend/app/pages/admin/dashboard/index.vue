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
            <div class="flex justify-between">
                <h1 class="text-2xl font-semibold mb-2">{{ totalRequests }} Applicants</h1>
                <div class="flex gap-5">
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
            <template v-if="!isCompany && studentData">
                <StudentRequestComponent
                    v-for="(data, index) in studentData"
                    :key="index"
                    :request-id="index"
                    :data="data"
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

const totalRequests = 50;
const isCompany = ref(false);

const api = useApi();
const studentData = ref<Profile>();

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

onMounted(() => {
    console.log("Selected Value:", selectedValue.value);
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
}

const studentOffset = ref(0);
const studentLimit = ref(10);

async function selectRecruit() {
    isCompany.value = false;
    try {
        const response = await api.get("/students", {
            params: {
                limit: studentLimit.value,
                offset: studentOffset.value,
            },
            withCredentials: true,
        });
        studentData.value = response.data as Profile;
        console.log("Fetched student data:", studentData.value);
    } catch (error) {
        console.error("Error fetching student data:", error);
    }
}
</script>
