<template>
    <div class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-between">
        <div>
            <div class="flex justify-between items-center">
                <h1 class="text-xl md:text-3xl font-bold text-left text-black py-4">
                    Student Registration
                </h1>
                <router-link to="/jobs">
                    <icon
                        name="material-symbols:skip-next-rounded"
                        class="text-2xl md:text-4xl text-gray-500 hover:text-primary transition-all duration-150 ease-in-out"
                    />
                </router-link>
            </div>
            <div class="flex flex-1 justify-center items-center gap-x-2">
                <UProgress
                    v-for="i in totalSteps"
                    :key="i"
                    :model-value="currentStep >= i ? 100 : 0"
                    size="lg"
                />
            </div>
            <StudentRegisterStepOne
                v-if="currentStep === 1"
                v-model:full-name="form.fullName"
                v-model:date-of-birth="form.dateOfBirth"
                v-model:phone="form.phone"
                v-model:about-me="form.aboutMe"
                v-model:github-u-r-l="form.githubURL"
                v-model:linkedin-u-r-l="form.linkedinURL"
                v-model:avatar="form.avatar"
                @update:avatar="onAvatarUpdate"
            />
            <StudentRegisterStepTwo
                v-if="currentStep === 2"
                v-model:student-id="form.studentId"
                v-model:major="form.major"
                v-model:student-status="form.studentStatus"
                v-model:verification-file="form.verificationFile"
                @update:verification-file="onVerificationFileUpdate"
            />
        </div>
        <div class="flex justify-center gap-x-2 py-4">
            <UButton
                v-if="currentStep > 1 && currentStep < totalSteps"
                class="w-[8em] h-fit text-xl rounded-md font-medium hover:cursor-pointer"
                color="neutral"
                variant="outline"
                label="Previous"
                :ui="{
                    base: 'justify-center bg-white text-black hover:bg-gray-200 hover:text-black',
                }"
                @click="handlePrevious"
            />

            <UButton
                v-if="currentStep < totalSteps - 1"
                class="w-[8em] h-fit text-xl text-white text-center rounded-md font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                label="Next"
                :ui="{
                    base: 'justify-center',
                }"
                @click="handleNext"
            />
            <UButton
                v-if="currentStep == totalSteps - 1"
                class="w-[8em] h-fit text-xl text-white text-center rounded-md font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                label="Submit"
                :ui="{
                    base: 'justify-center',
                }"
                @click="onSubmit"
            />
            <UButton
                v-if="currentStep == totalSteps"
                class="w-[8em] h-fit text-xl text-white text-center rounded-md font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                label="Done"
                :ui="{
                    base: 'justify-center',
                }"
                @click="handleDone"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";

const totalSteps = 3;
const currentStep = ref(1);

const form = reactive({
    fullName: "",
    dateOfBirth: "",
    phone: "",
    aboutMe: "",
    githubURL: "",
    linkedinURL: "",
    avatar: null as File | null,

    studentId: "",
    major: "",
    studentStatus: "",
    verificationFile: null as File | null,
});

const handleNext = () => {
    if (currentStep.value < totalSteps) {
        currentStep.value++;
    }
};

const handlePrevious = () => {
    if (currentStep.value > 1) {
        currentStep.value--;
    }
};

const onAvatarUpdate = (newAvatarFile: File) => {
    form.avatar = newAvatarFile;
};

const onVerificationFileUpdate = (newFile: File) => {
    form.verificationFile = newFile;
};

const onSubmit = () => {
    // TODO: Implement form submission logic
    console.log("Form submitted");
    for (const key in form) {
        console.log(`${key}: ${form[key]}`);
    }
    currentStep.value++;
};

const handleDone = () => {
    navigateTo("/jobs");
};
</script>
