<template>
    <section class="w-full overflow-x-hidden">
        <h1 class="text-5xl text-primary-800 dark:text-primary font-bold mt-6 mb-6">Logs</h1>

        <div class="h-[3em] overflow-hidden border-b-1 my-5">
            <div class="flex flex-row gap-2 h-[6em] max-w-[40em] left-0 top-0">
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="tabClass('audit')"
                    @click="active = 'audit'"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Audit</p>
                </div>
                <div
                    class="hover:cursor-pointer transition-all duration-150 text-center"
                    :class="tabClass('email')"
                    @click="active = 'email'"
                >
                    <p class="font-bold px-5 py-1 text-2xl">Email</p>
                </div>
            </div>
        </div>
        <!-- Count and sorting -->
        <section>
            <div v-if="isLoading">
                <USkeleton class="h-[3em] w-full left-0 top-0 mb-5" />
            </div>
            <div v-else class="flex justify-between items-center">
                <h2 class="text-2xl font-semibold mb-2">
                    {{ currentCount }} {{ active === "audit" ? "Audit logs" : "Email logs" }}
                </h2>
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
            <!-- Audit Logs -->
            <div v-if="active === 'audit'" class="flex flex-col gap-3">
                <div v-if="audits.length === 0" class="text-center text-neutral-500 py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p>No audit logs found.</p>
                </div>
                <div v-for="a in sortedAudits" v-else :key="a.id" class="r-card text-left">
                    <div class="flex items-start gap-3">
                        <Icon name="ic:baseline-history" class="w-8 h-8 text-gray-400 mt-1" />
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold truncate">
                                {{ a.action }}
                                <span v-if="a.objectName" class="font-normal text-gray-600"
                                    >on {{ a.objectName }}</span
                                >
                                <span v-if="a.objectId" class="font-mono text-gray-500"
                                    >#{{ a.objectId }}</span
                                >
                            </p>
                            <p class="text-sm text-gray-500 truncate">
                                by {{ a.actorId }} • {{ new Date(a.createdAt).toLocaleString() }}
                            </p>
                            <p v-if="a.reason" class="text-sm text-gray-600 mt-1">
                                Reason: {{ a.reason }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Email Logs -->
            <div v-else class="flex flex-col gap-3">
                <div v-if="emails.length === 0" class="text-center text-neutral-500 py-10">
                    <Icon name="ic:baseline-inbox" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p>No email logs found.</p>
                </div>
                <div v-for="e in sortedEmails" v-else :key="e.id" class="r-card text-left">
                    <div class="flex items-start gap-3">
                        <Icon name="ic:baseline-email" class="w-8 h-8 text-gray-400 mt-1" />
                        <div class="flex-1 min-w-0">
                            <p class="font-semibold truncate">
                                {{ e.subject }}
                                <span
                                    class="ml-2 text-xs px-2 py-0.5 rounded-full"
                                    :class="badgeClass(e.status)"
                                >
                                    {{ e.status }}
                                </span>
                            </p>
                            <p class="text-sm text-gray-500 truncate">to {{ e.to }}</p>
                            <p class="text-sm text-gray-500">
                                {{ new Date(e.createdAt).toLocaleString() }}
                            </p>
                            <p
                                v-if="e.status !== 'delivered' && (e.errorCode || e.errorDesc)"
                                class="text-sm text-red-600 mt-1"
                            >
                                {{ e.errorCode }} {{ e.errorDesc }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div class="mt-4 flex gap-2">
                <UButton icon="ic:baseline-refresh" label="Refresh" @click="reload" />
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
definePageMeta({ layout: "admin" });

type AuditItem = {
    id: number;
    actorId: string;
    createdAt: string;
    action: string;
    reason?: string;
    objectName?: string;
    objectId?: string;
};

type EmailItem = {
    id: number;
    to: string;
    subject: string;
    body: string;
    createdAt: string;
    status: "delivered" | "temporary_error" | "permanent_error" | string;
    errorCode?: string;
    errorDesc?: string;
};

const api = useApi();

const active = ref<"audit" | "email">("audit");
const isLoading = ref(true);

const audits = ref<AuditItem[]>([]);
const emails = ref<EmailItem[]>([]);

const sortOptions = ref([
    { label: "Latest", id: "latest" },
    { label: "Oldest", id: "oldest" },
    { label: "Name A-Z", id: "name_az" },
    { label: "Name Z-A", id: "name_za" },
]);
const selectSortOption = ref("latest");

const currentCount = computed(() =>
    active.value === "audit" ? audits.value.length : emails.value.length
);

const sortedAudits = computed(() => {
    const list = [...audits.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) => (a.actorId || "").localeCompare(b.actorId || ""));
        case "name_za":
            return list.sort((a, b) => (b.actorId || "").localeCompare(a.actorId || ""));
        case "latest":
        default:
            return list.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
    }
});

const sortedEmails = computed(() => {
    const list = [...emails.value];
    switch (selectSortOption.value) {
        case "oldest":
            return list.sort(
                (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            );
        case "name_az":
            return list.sort((a, b) => (a.subject || "").localeCompare(b.subject || ""));
        case "name_za":
            return list.sort((a, b) => (b.subject || "").localeCompare(a.subject || ""));
        case "latest":
        default:
            return list.sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            );
    }
});

function tabClass(tab: "audit" | "email") {
    return active.value === tab
        ? "bg-primary-200 flex flex-col border rounded-3xl w-1/3 text-primary-800 hover:bg-primary-300"
        : "bg-gray-200 flex flex-col border rounded-3xl w-1/3 text-gray-500 hover:bg-gray-300";
}

function badgeClass(status: string) {
    switch (status) {
        case "delivered":
            return "bg-green-100 text-green-700";
        case "temporary_error":
            return "bg-yellow-100 text-yellow-700";
        case "permanent_error":
            return "bg-red-100 text-red-700";
        default:
            return "bg-gray-100 text-gray-700";
    }
}

async function loadAudits() {
    const res = await api.get("/admin/audits", { withCredentials: true });
    const arr: AuditItem[] = Array.isArray(res.data) ? (res.data as AuditItem[]) : [];
    audits.value = arr.map((x) => ({
        id: x.id,
        actorId: x.actorId,
        createdAt: x.createdAt,
        action: x.action,
        reason: x.reason,
        objectName: x.objectName,
        objectId: x.objectId,
    }));
}

async function loadEmails() {
    const res = await api.get("/admin/emaillog", { withCredentials: true });
    const arr: EmailItem[] = Array.isArray(res.data) ? (res.data as EmailItem[]) : [];
    emails.value = arr.map((x) => ({
        id: x.id,
        to: x.to,
        subject: x.subject,
        body: x.body,
        createdAt: x.createdAt,
        status: x.status,
        errorCode: x.errorCode,
        errorDesc: x.errorDesc,
    }));
}

async function reload() {
    isLoading.value = true;
    try {
        await Promise.all([loadAudits(), loadEmails()]);
    } finally {
        isLoading.value = false;
    }
}

onMounted(reload);
</script>

<style scoped>
.r-card {
    box-shadow:
        0 4px 6px -1px rgba(0, 0, 0, 0.2),
        0 2px 4px -2px rgba(0, 0, 0, 0.2);
    text-align: left;
    padding: 1rem;
    border-radius: 0.5rem;
    background-color: #fdfdfd;
}
.dark .r-card {
    background-color: #1f2937;
}
</style>
