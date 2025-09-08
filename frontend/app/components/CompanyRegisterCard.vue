<template>
    <div class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-between">
        <div>
            <div class="flex justify-between items-center">
                <h1 class="text-xl md:text-3xl font-bold text-left text-black py-4">
                    Company Registration
                </h1>
            </div>
            <div class="flex flex-1 justify-center items-center gap-x-2">
                <UProgress
                    v-for="i in totalSteps"
                    :key="i"
                    :model-value="currentStep >= i ? 100 : 0"
                    size="lg"
                />
            </div>
            <CompanyRegisterStepOne
                v-if="currentStep === 1"
                ref="stepOneRef"
                v-model:company-name="form.companyName"
                v-model:company-email="form.companyEmail"
                v-model:password="form.password"
                v-model:phone="form.phone"
                v-model:address="form.address"
                v-model:city="form.city"
                v-model:country="form.country"
            />
            <CompanyRegisterStepTwo
                v-if="currentStep === 2"
                ref="stepTwoRef"
                v-model:about="form.about"
                v-model:company-logo="form.companyLogo"
                v-model:banner="form.banner"
            />
            <div v-if="currentStep === 3" class="text-center py-8">
                <icon name="hugeicons:checkmark-circle-03" class="text-[15em] text-primary mb-4" />
                <h2 class="text-4xl font-bold text-primary mb-2">You’re All Set</h2>
                <p class="text-gray-600">You have successfully registered your account.</p>
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
    companyName: "",
    companyEmail: "",
    password: "",
    phone: "",
    address: "",
    city: "",
    country: "",
    about: "",
    companyLogo: null as File | null,
    banner: null as File | null,
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
        const formData = new FormData();
        formData.append("name", form.companyName);
        formData.append("email", form.companyEmail);
        formData.append("password", form.password);
        formData.append("phone", form.phone);
        formData.append("address", form.address);
        formData.append("city", form.city);
        formData.append("country", form.country);
        formData.append("about", form.about);
        if (form.companyLogo) {
            formData.append("logo", form.companyLogo);
        }
        if (form.banner) {
            formData.append("banner", form.banner);
        }

        const response = await $fetch("/company/register", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            body: formData,
        });

        if (response.token) {
            localStorage.setItem("jwt_token", response.token);
            localStorage.setItem("username", response.username);
            if (response.isCompany) {
                localStorage.setItem("role", "company");
            } else if (response.isStudent) {
                localStorage.setItem("role", "student");
            }
        }

        currentStep.value++;
    } catch (error: unknown) {
        console.error("Submission error:", error);

        let errorMessage = "There was an error submitting your registration. Please try again.";
        const apiError = error as { status?: number; data?: { error?: string }; message?: string };

        if (apiError.data?.error) {
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
        description: "Registration complete. Redirecting to dashboard...",
        color: "success",
    });
    navigateTo("/dashboard");
};
</script>
