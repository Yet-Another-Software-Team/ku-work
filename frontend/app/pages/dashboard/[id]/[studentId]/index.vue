<template>
    <div>
        <ApplicantCard v-if="data" :data="data" />
    </div>
</template>

<script lang="ts" setup>
import ApplicantCard from "~/components/job/ApplicantCard.vue";
import { multipleMockUserData } from "~/data/mockData";
import { computed } from "vue";

definePageMeta({
    layout: "viewer",
    middleware: "company",
});

const route = useRoute();
const studentIdParam = route.params.studentId as string;

const studentProfile = computed(() => {
    return (
        multipleMockUserData.find((d) => d.profile.id === studentIdParam) ?? multipleMockUserData[0]
    );
});

const data = computed(() => {
    if (!studentProfile.value) {
        return null;
    }
    const { studentId, ...profile } = studentProfile.value.profile;
    return {
        profile,
        email: "john.doe@ku.th",
        studentId: studentIdParam,
    };
});
</script>
