<template>
    <section class="w-full overflow-x-hidden">
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mt-6 mb-6">Accounts</h1>

        <div class="h-[3em] overflow-hidden border-b-1 my-5">
            <div class="flex flex-row gap-2 h-[6em] max-w-[40em] left-0 top-0">
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="tabClass('students')"
                    @click="active = 'students'"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Students</p>
                </div>
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="tabClass('companies')"
                    @click="active = 'companies'"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Companies</p>
                </div>
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="tabClass('staff')"
                    @click="active = 'staff'"
                >
                    <p class="font-bold px-5 py-1 text-2xl">KU Staff</p>
                </div>
            </div>
        </div>

        <!-- Count and sorting -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[3em] w-full left-0 top-0 mb-5" />
            </div>
            <div v-else class="flex justify-between items-center">
                <h2 class="text-2xl font-semibold mb-2">{{ currentCount }} {{ activeLabel }}</h2>
                <div class="flex items-center gap-3">
                    <span class="text-2xl font-semibold mb-2">Sort by:</span>
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

        <div v-if="isLoading">
            <USkeleton v-for="n in 6" :key="n" class="h-[6em] w-full mb-4" />
        </div>
        <div v-else>
            <div v-if="active === 'students'" class="flex flex-col gap-3">
                <div v-if="students.length === 0" class="text-center text-neutral-500 py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p>No students found.</p>
                </div>
                <div
                    v-for="student in sortedStudents"
                    v-else
                    :key="student.userId"
                    class="text-left p-4 rounded-md bg-[#fdfdfd] dark:bg-[#1f2937] shadow-[0_4px_6px_-1px_rgba(0,0,0,0.2),0_2px_4px_-2px_rgba(0,0,0,0.2)]"
                >
                    <div class="flex items-center gap-3">
                        <div class="flex-shrink-0">
                            <img
                                v-if="student.photoId"
                                :src="`${config.public.apiBaseUrl}/files/${student.photoId}`"
                                alt="Photo"
                                class="w-12 h-12 rounded-full object-cover"
                            />
                            <Icon
                                v-else
                                name="ic:baseline-account-circle"
                                class="w-12 h-12 text-gray-400"
                            />
                        </div>
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold truncate">
                                {{ student.fullName || `${student.firstName} ${student.lastName}` }}
                            </p>
                            <p class="text-sm text-gray-500 truncate">{{ student.email }}</p>
                        </div>
                        <div class="text-sm text-gray-500">
                            {{ new Date(student.createdAt).toLocaleString() }}
                        </div>
                    </div>
                </div>
            </div>

            <div v-else-if="active === 'companies'" class="flex flex-col gap-3">
                <div v-if="companies.length === 0" class="text-center text-neutral-500 py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p>No companies found.</p>
                </div>
                <div
                    v-for="company in sortedCompanies"
                    v-else
                    :key="company.userId"
                    class="text-left p-4 rounded-md bg-[#fdfdfd] dark:bg-[#1f2937] shadow-[0_4px_6px_-1px_rgba(0,0,0,0.2),0_2px_4px_-2px_rgba(0,0,0,0.2)]"
                >
                    <div class="flex items-center gap-3">
                        <div class="flex-shrink-0">
                            <img
                                v-if="company.photoId"
                                :src="`${config.public.apiBaseUrl}/files/${company.photoId}`"
                                alt="Logo"
                                class="w-12 h-12 rounded-md object-cover"
                            />
                            <Icon
                                v-else
                                name="ic:baseline-business"
                                class="w-12 h-12 text-gray-400"
                            />
                        </div>
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold truncate">{{ company.name }}</p>
                            <p class="text-sm text-gray-500 truncate">{{ company.email }}</p>
                        </div>
                        <div class="text-sm text-gray-500">
                            {{ new Date(company.createdAt).toLocaleString() }}
                        </div>
                    </div>
                </div>
            </div>

            <div v-else class="flex flex-col gap-3">
                <div v-if="staff.length === 0" class="text-center text-neutral-500 py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p>No staff entries found.</p>
                </div>
                <div
                    v-for="a in sortedStaff"
                    v-else
                    :key="a.actorId"
                    class="text-left p-4 rounded-md bg-[#fdfdfd] dark:bg-[#1f2937] shadow-[0_4px_6px_-1px_rgba(0,0,0,0.2),0_2px_4px_-2px_rgba(0,0,0,0.2)]"
                >
                    <div class="flex items-center gap-3">
                        <Icon name="ic:baseline-badge" class="w-12 h-12 text-gray-400" />
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold truncate">{{ a.actorId }}</p>
                            <p class="text-sm text-gray-500 truncate">
                                Last activity: {{ new Date(a.latestActivity).toLocaleString() }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
definePageMeta({ layout: "admin" });

type StudentItem = {
    userId: string;
    firstName: string;
    lastName: string;
    fullName?: string;
    email: string;
    photoId?: string;
    createdAt: string;
};

type CompanyItem = {
    userId: string;
    name: string;
    email: string;
    photoId?: string;
    createdAt: string;
};

type StaffItem = { actorId: string; latestActivity: string };

const api = useApi();
const config = useRuntimeConfig();

const active = ref<"students" | "companies" | "staff">("students");
const isLoading = ref(true);

const students = ref<StudentItem[]>([]);
const companies = ref<CompanyItem[]>([]);
const staff = ref<StaffItem[]>([]);

function tabClass(tab: "students" | "companies" | "staff") {
    return active.value === tab
        ? "bg-primary-200 flex flex-col border rounded-3xl w-1/3 text-primary-800 hover:bg-primary-300"
        : "bg-gray-200 flex flex-col border rounded-3xl w-1/3 text-gray-500 hover:bg-gray-300";
}

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);
const selectSortOption = ref("latest");

const activeLabel = computed(() =>
    active.value === "students" ? "Students" : active.value === "companies" ? "Companies" : "Staff"
);
const currentCount = computed(() =>
    active.value === "students"
        ? students.value.length
        : active.value === "companies"
          ? companies.value.length
          : staff.value.length
);

const sortedStudents = computed(() => {
    const list = [...students.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) =>
                (a.fullName || `${a.firstName} ${a.lastName}`).localeCompare(
                    b.fullName || `${b.firstName} ${b.lastName}`
                )
            );
        case "name_za":
            return list.sort((a, b) =>
                (b.fullName || `${b.firstName} ${b.lastName}`).localeCompare(
                    a.fullName || `${a.firstName} ${a.lastName}`
                )
            );
        case "latest":
        default:
            return list.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
    }
});

const sortedCompanies = computed(() => {
    const list = [...companies.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) => a.name.localeCompare(b.name));
        case "name_za":
            return list.sort((a, b) => b.name.localeCompare(a.name));
        case "latest":
        default:
            return list.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
    }
});

const sortedStaff = computed(() => {
    const list = [...staff.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) =>
                    new Date(a.latestActivity).getTime() - new Date(b.latestActivity).getTime()
            );
        case "name_az":
            return list.sort((a, b) => a.actorId.localeCompare(b.actorId));
        case "name_za":
            return list.sort((a, b) => b.actorId.localeCompare(a.actorId));
        case "latest":
        default:
            return list.sort(
                (a, b) =>
                    new Date(b.latestActivity).getTime() - new Date(a.latestActivity).getTime()
            );
    }
});

async function loadStudents() {
    const res = await api.get("/students", {
        params: { limit: 64, sortBy: "latest" },
        withCredentials: true,
    });
    const arr = Array.isArray(res.data) ? (res.data as unknown[]) : [];
    students.value = arr.map((student: unknown) => ({
        userId: student.userId,
        firstName: student.firstName,
        lastName: student.lastName,
        fullName: student.fullName,
        email: student.email,
        photoId: student.photoId,
        createdAt: student.createdAt,
    }));
}

async function loadCompanies() {
    const res = await api.get("/company", { withCredentials: true });
    const arr = Array.isArray(res.data) ? (res.data as unknown[]) : [];
    companies.value = arr
        .map((company: unknown) => ({
            userId: company.userId,
            name: company.name,
            email: company.email,
            photoId: company.photoId,
            createdAt: company.createdAt,
        }))
        .sort(
            (a: CompanyItem, b: CompanyItem) =>
                new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
}

async function loadStaff() {
    const res = await api.get("/admin/audits", { withCredentials: true });
    const list: Array<{ actorId: string; createdAt: string }> = (res.data as unknown[]).map(
        (account: unknown) => ({
            actorId: account.actorId,
            createdAt: account.createdAt,
        })
    );
    const map = new Map<string, string>();
    for (const a of list) {
        const prev = map.get(a.actorId);
        if (!prev || new Date(a.createdAt).getTime() > new Date(prev).getTime()) {
            map.set(a.actorId, a.createdAt);
        }
    }
    staff.value = Array.from(map.entries())
        .map(([actorId, latestActivity]) => ({ actorId, latestActivity }))
        .sort(
            (a, b) => new Date(b.latestActivity).getTime() - new Date(a.latestActivity).getTime()
        );
}

async function loadAll() {
    isLoading.value = true;
    try {
        await Promise.all([loadStudents(), loadCompanies(), loadStaff()]);
    } finally {
        isLoading.value = false;
    }
}

onMounted(loadAll);
</script>
