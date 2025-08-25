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
                ref="stepOneRef"
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
                ref="stepTwoRef"
                v-model:student-id="form.studentId"
                v-model:major="form.major"
                v-model:student-status="form.studentStatus"
                v-model:verification-file="form.verificationFile"
                @update:verification-file="onVerificationFileUpdate"
            />
            <div v-if="currentStep === 3" class="text-center py-8">
                <icon name="material-symbols:check-circle" class="text-6xl text-green-500 mb-4" />
                <h2 class="text-2xl font-bold text-green-600 mb-2">Registration Successful!</h2>
                <p class="text-gray-600">
                    Your student registration has been submitted successfully.
                </p>
            </div>
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
                :disabled="!canProceedToNext"
                :ui="{
                    base: 'justify-center',
                }"
                @click="handleNext"
            />
            <UButton
                v-if="currentStep == totalSteps - 1"
                class="w-[8em] h-fit text-xl text-white text-center rounded-md font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                label="Submit"
                :loading="isSubmitting"
                :disabled="!canSubmit"
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
import { ref, reactive, computed } from "vue";

const toast = useToast();
const config = useRuntimeConfig();
const totalSteps = 3;
const currentStep = ref(1);
const isSubmitting = ref(false);

const stepOneRef = ref<{ isValid: boolean } | null>(null);
const stepTwoRef = ref<{ isValid: boolean } | null>(null);

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

const canProceedToNext = computed(() => {
    if (currentStep.value === 1 && stepOneRef.value) {
        return stepOneRef.value.isValid;
    }
    if (currentStep.value === 2 && stepTwoRef.value) {
        return stepTwoRef.value.isValid;
    }
    return false;
});

const canSubmit = computed(() => {
    if (currentStep.value === 2 && stepTwoRef.value) {
        return stepTwoRef.value.isValid;
    }
    return false;
});

const handleNext = () => {
    if (!canProceedToNext.value) {
        toast.add({
            title: "Validation Error",
            description: "Please fill in all required fields correctly before proceeding",
            color: "error",
        });
        return;
    }

    if (currentStep.value < totalSteps) {
        currentStep.value++;
        toast.add({
            title: "Progress Saved",
            description: `Step ${currentStep.value - 1} completed successfully`,
            color: "success",
        });
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

const onSubmit = async () => {
    if (!canSubmit.value) {
        toast.add({
            title: "Validation Error",
            description: "Please fix all validation errors before submitting",
            color: "error",
        });
        return;
    }

    isSubmitting.value = true;

    try {
        const token = localStorage.getItem("jwt_token");
        if (!token) {
            toast.add({
                title: "Authentication Error",
                description: "Please log in to submit your registration",
                color: "error",
            });
            navigateTo("/");
            return;
        }

        const formData = new FormData();

        formData.append("studentId", form.studentId);
        formData.append("major", form.major);
        formData.append("studentStatus", form.studentStatus);
        formData.append("photo", form.avatar as File);
        formData.append("statusPhoto", form.verificationFile as File);

        if (form.phone) {
            formData.append("phone", form.phone);
        }
        if (form.dateOfBirth) {
            const date = new Date(form.dateOfBirth);
            formData.append("birthDate", date.toISOString());
        }
        if (form.aboutMe) {
            formData.append("aboutMe", form.aboutMe);
        }
        if (form.githubURL) {
            formData.append("github", form.githubURL);
        }
        if (form.linkedinURL) {
            formData.append("linkedIn", form.linkedinURL);
        }

        await $fetch("/students/register", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            body: formData,
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });

        currentStep.value++;

        toast.add({
            title: "Registration Successful",
            description: "Your student registration has been submitted successfully!",
            color: "success",
        });
    } catch (error: unknown) {
        console.error("Submission error:", error);

        let errorMessage = "There was an error submitting your registration. Please try again.";
        const apiError = error as { status?: number; data?: { error?: string }; message?: string };

        if (apiError.status === 401) {
            errorMessage = "Authentication failed. Please log in again.";
            localStorage.removeItem("jwt_token");
            navigateTo("/login");
            return;
        } else if (apiError.status === 409) {
            errorMessage = "Student registration already exists or Student ID is already taken.";
        } else if (apiError.data?.error) {
            errorMessage = apiError.data.error;
        }

        toast.add({
            title: "Submission Failed",
            description: errorMessage,
            color: "error",
        });
    } finally {
        isSubmitting.value = false;
    }
};

const handleDone = () => {
    toast.add({
        title: "Welcome!",
        description: "Registration complete. Redirecting to jobs page...",
        color: "success",
    });
    navigateTo("/jobs");
};
</script>
