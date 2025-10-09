<template>
    <div class="mt-5">
        <ApplicantCard
            v-if="applicantData"
            :data="applicantData"
            :loading="isLoading"
            @status-changed="statusChanged"
        />
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import ApplicantCard from "~/components/job/ApplicantCard.vue";

const api = useApi();

definePageMeta({
    layout: "viewer",
    middleware: "company",
});

const route = useRoute();
const jobId = route.params.id as string;
const studentIdParam = route.params.studentId as string;

const applicantData = ref(null);
const isLoading = ref(false);

const statusChanged = async (status: string) => {
    isLoading.value = true;
    try {
        await api.patch(`/jobs/${jobId}/applications/${studentIdParam}`, { status });
        await fetchData();
    } catch (error) {
        console.error("Failed to update status:", error);
    } finally {
        isLoading.value = false;
    }
};

const fetchData = async () => {
    try {
        const response = await api.get(`/jobs/${jobId}/applications/${studentIdParam}`);
        console.log(response.data);
        applicantData.value = response.data;
    } catch (error) {
        console.error("Failed to fetch applicant data:", error);
    }
};

onMounted(async () => {
    isLoading.value = true;
    await fetchData();
    isLoading.value = false;
});
</script>
