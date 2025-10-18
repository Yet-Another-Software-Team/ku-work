<template>
    <div class="rounded-lg">
        <NuxtLink :to="`/dashboard/${route.params.id}`">
            <h1
                class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
            >
                <Icon name="iconoir:nav-arrow-left" class="items-center" />
                <span>Back</span>
            </h1>
        </NuxtLink>
        <section class="flex flex-col sm:flex-row">
            <section class="w-full sm:w-[20em] flex flex-col gap-5 m-3">
                <div class="r-card">
                    <div class="h-fit flex flex-col items-center">
                        <div v-if="data.photoId" class="size-[16em] flex-shrink-0">
                            <img
                                :src="`${config.public.apiBaseUrl}/files/${data.photoId}`"
                                alt="Profile photo"
                                class="size-full object-cover rounded-full p-5"
                            />
                        </div>
                        <div
                            v-else
                            class="flex items-center justify-center w-[16em] h-[16em] flex-shrink-0"
                        >
                            <Icon
                                name="ic:baseline-account-circle"
                                class="size-full text-gray-400"
                            />
                        </div>
                        <div class="text-xl mt-5">
                            <h2 class="text-2xl font-semibold text-gray-900 dark:text-[#FDFDFD]">
                                {{ data.username }}
                            </h2>
                            <p class="text-sm text-gray-500 dark:text-gray-400">
                                {{ data.studentId }}
                            </p>
                            <p class="text-gray-600 dark:text-gray-300">
                                {{ data.major }}
                            </p>
                            <div v-if="data.status" class="mt-4">
                                <span
                                    :class="statusBadgeClass"
                                    class="text-sm font-medium px-3 py-1 rounded-full capitalize"
                                >
                                    {{ data.status }}
                                </span>
                            </div>
                        </div>
                        <hr class="border-gray-500 dark:border-gray-600 m-5 w-full" />
                        <ul>
                            <li v-if="data.linkedIn">
                                <a
                                    :href="data.linkedIn"
                                    target="_blank"
                                    class="flex hover:underline items-center justify-center gap-2 my-5"
                                >
                                    <Icon name="devicon:linkedin" class="size-[2em]" />
                                    <span class="truncate w-[12em]">{{ data.username }}</span>
                                </a>
                            </li>
                            <li v-if="data.github">
                                <a
                                    :href="data.github"
                                    target="_blank"
                                    class="flex hover:underline items-center justify-center gap-2 my-5"
                                >
                                    <Icon
                                        name="devicon:github"
                                        class="size-[2em] bg-[#FDFDFD] rounded-full"
                                    />
                                    <span class="truncate w-[12em]">{{ data.username }}</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                </div>
                <div class="r-card">
                    <div
                        v-for="(file, index) in data.files"
                        :key="file.id"
                        class="flex justify-between items-center mb-2"
                    >
                        <span class="text-xl">File {{ index + 1 }}</span>
                        <a :href="`${config.public.apiBaseUrl}/files/${file.id}`" target="_blank">
                            <Icon
                                name="ic:outline-file-download"
                                class="size-[2em] hover:text-primary transition-all duration-200"
                            />
                        </a>
                    </div>
                </div>
            </section>

            <section class="flex flex-col gap-5 m-3 w-full">
                <div class="r-card">
                    <div class="flex flex-col text-xl text-left whitespace-normal">
                        <p class="mt-2 text-gray-800 dark:text-gray-200 font-semibold">
                            Age: <span class="font-normal">{{ age }}</span>
                        </p>
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            Phone: <span class="font-normal">{{ data.phone }}</span>
                        </p>
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            Email: <span class="font-normal">{{ data.email }}</span>
                        </p>
                    </div>
                </div>
                <div class="r-card">
                    <div class="flex flex-col text-xl text-left">
                        <p class="text-gray-800 dark:text-gray-200 font-semibold">
                            <span class="font-normal">{{ data.aboutMe }}</span>
                        </p>
                    </div>
                </div>

                <div class="flex gap-5">
                    <span class="c-ubutton w-full sm:w-1/2">
                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl cursor-pointer"
                            variant="outline"
                            color="error"
                            label="Reject"
                            :icon="'iconoir:xmark'"
                            :disabled="loading"
                            @click="showRejectModal = true"
                        />
                        <UModal v-model:open="showRejectModal">
                            <template #header>
                                <p>
                                    Are you sure you want to <strong>reject</strong>
                                    {{ data.username }}?
                                </p>
                            </template>
                            <template #body>
                                <div class="flex flex-col gap-2">
                                    <div class="flex justify-end gap-2">
                                        <UButton
                                            class="cursor-pointer"
                                            variant="outline"
                                            color="neutral"
                                            label="Cancel"
                                            @click="closeModal"
                                        />
                                        <UButton
                                            class="cursor-pointer"
                                            variant="solid"
                                            color="error"
                                            label="Reject"
                                            @click="
                                                () => {
                                                    updateStatus('rejected');
                                                    closeModal();
                                                }
                                            "
                                        />
                                    </div>
                                </div>
                            </template>
                        </UModal>
                    </span>
                    <span class="c-ubutton w-full sm:w-1/2">
                        <UButton
                            class="size-full font-bold p-1 rounded flex items-center gap-1 px-2 text-xl cursor-pointer"
                            variant="outline"
                            color="primary"
                            label="Accept"
                            :icon="'iconoir:check'"
                            :disabled="loading"
                            @click="showAcceptModal = true"
                        />
                        <UModal v-model:open="showAcceptModal">
                            <template #header>
                                <p>
                                    Are you sure you want to <strong>accept</strong>
                                    {{ data.username }}?
                                </p>
                            </template>
                            <template #body>
                                <div class="flex flex-col gap-2">
                                    <div class="flex justify-end gap-2">
                                        <UButton
                                            class="cursor-pointer"
                                            variant="outline"
                                            color="neutral"
                                            label="Cancel"
                                            @click="closeModal"
                                        />
                                        <UButton
                                            class="cursor-pointer"
                                            variant="solid"
                                            color="primary"
                                            label="Accept"
                                            @click="
                                                () => {
                                                    updateStatus('accepted');
                                                    closeModal();
                                                }
                                            "
                                        />
                                    </div>
                                </div>
                            </template>
                        </UModal>
                    </span>
                </div>
            </section>
        </section>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const showAcceptModal = ref(false);
const showRejectModal = ref(false);

const closeModal = () => {
    showAcceptModal.value = false;
    showRejectModal.value = false;
};

const props = defineProps({
    data: {
        type: Object,
        required: true,
    },
    loading: {
        type: Boolean,
        required: true,
    },
});

const emit = defineEmits(["status-changed"]);

const route = useRoute();
const config = useRuntimeConfig();

const age = computed(() => {
    if (!props.data?.birthDate) return "N/A"; // Return default value if no date
    const birth = new Date(props.data.birthDate);
    if (isNaN(birth.getTime())) return "N/A"; // Check for invalid date

    const today = new Date();
    let years = today.getFullYear() - birth.getFullYear();
    const m = today.getMonth() - birth.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) {
        years--;
    }
    return years;
});

const statusBadgeClass = computed(() => {
    switch (props.data.status) {
        case "accepted":
            return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
        case "rejected":
            return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
        case "pending":
        default:
            return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
    }
});

async function updateStatus(status: "accepted" | "rejected") {
    emit("status-changed", status);
}
</script>

<style scoped>
/* Styles remain the same */
.r-card {
    box-shadow:
        0 4px 6px -1px rgba(0, 0, 0, 0.2),
        0 2px 4px -2px rgba(0, 0, 0, 0.2);
    text-align: center;
    padding: 1.25rem;
    border-radius: 0.5rem;
    background-color: #fdfdfd;
}

.dark .r-card {
    background-color: #1f2937;
}

.c-ubutton {
    background-color: #fdfdfd;
}

.dark .c-ubutton {
    background-color: #1f2937;
}
</style>
