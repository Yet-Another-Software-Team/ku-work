<template>
    <div
        v-if="isLoading"
        class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-center items-center"
    >
        <icon name="svg-spinners:180-ring-with-bg" class="text-[10em] text-primary mb-4" />
        <h2 class="text-4xl font-bold text-primary mb-2">Loading...</h2>
        <p class="text-neutral-500">Checking your registration status.</p>
    </div>
    <div v-else class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-between">
        <div>
            <div class="flex justify-between items-center">
                <h1 class="text-xl md:text-3xl font-bold text-left text-black py-4">
                    Student Registration
                </h1>
                <!-- Skip -->
                <router-link
                    v-if="currentStep == 1"
                    to="/jobs"
                    class="flex items-center text-neutral-500 hover:text-primary transition-all duration-150 ease-in-out"
                >
                    <p class="text-sm font-bold uppercase">Skip</p>
                    <icon name="material-symbols:skip-next-rounded" class="text-2xl md:text-4xl" />
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
                <icon name="hugeicons:checkmark-circle-03" class="text-[15em] text-primary mb-4" />
                <h2 class="text-4xl font-bold text-primary mb-2">You’re All Set</h2>
                <p class="text-gray-600">
                    You have successfully registered your account. When admin approve your profile,
                    you will be notified in your email
                </p>
            </div>
        </div>
        <div v-if="currentStep === 2" class="flex justify-center">
            <TurnstileWidget @callback="(tk) => (cfToken = tk)" />
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
import { ref, reactive, computed, onMounted } from "vue";
import { useApi } from "@/composables/useApi";

const api = useApi();
const toast = useToast();
const totalSteps = 3;
const currentStep = ref(1);
const isSubmitting = ref(false);
const isLoading = ref(true);

const stepOneRef = ref<{ isValid: boolean } | null>(null);
const stepTwoRef = ref<{ isValid: boolean } | null>(null);

const cfToken = ref<string>("");

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

onMounted(() => {
    const username = localStorage.getItem("username");
    const role = localStorage.getItem("role");
    const isRegistered = localStorage.getItem("isRegistered");

    if (role === "viewer" && isRegistered === "true") {
        toast.add({
            title: "Already Registered",
            description: "Redirecting to jobs page...",
            color: "success",
        });
        navigateTo("/jobs", { replace: true });
        return;
    }

    if (role !== "viewer" || !username) {
        navigateTo("/", { replace: true });
        return;
    }
    form.fullName = username || "";
    isLoading.value = false;
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
        const token = localStorage.getItem("token");
        if (!token) {
            toast.add({
                title: "Authentication Error",
                description: "Please log in to submit your registration",
                color: "error",
            });
            navigateTo("/", { replace: true });
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

        await api.postFormData("/auth/student/register", formData, {
            headers: {
                "X-Turnstile-Token": cfToken.value,
            },
        });
        localStorage.setItem("isRegistered", "true"); // Set isRegistered to True.
        currentStep.value++;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
        console.error("Submission error:", error);
        api.showErrorToast(error, "Submission Failed");
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
    navigateTo("/jobs", { replace: true });
};
</script>
