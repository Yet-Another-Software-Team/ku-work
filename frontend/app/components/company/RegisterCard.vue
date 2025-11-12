<template>
    <div
        v-if="isLoading"
        class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-center items-center"
    >
        <icon name="svg-spinners:180-ring-with-bg" class="text-[10em] text-primary mb-4" />
        <h2 class="text-4xl font-bold text-primary mb-2">Loading...</h2>
    </div>
    <div v-else class="bg-white w-full min-h-[50em] rounded-lg p-5 flex flex-col justify-between">
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
                v-model:company-website="form.companyWebsite"
                v-model:password="form.password"
                v-model:r-password="form.rPassword"
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
            <div v-if="currentStep === 2" class="my-4">
                <TermsConsent
                    v-model:accepted="termsAccepted"
                    :docs="termsDocs"
                    :disabled="isSubmitting"
                />
            </div>
            <div v-if="currentStep === 3" class="text-center py-8">
                <icon name="hugeicons:checkmark-circle-03" class="text-[15em] text-primary mb-4" />
                <h2 class="text-4xl font-bold text-primary mb-2">You’re All Set</h2>
                <p class="text-gray-600">You have successfully registered your account.</p>
            </div>
        </div>
        <div v-if="currentStep === 2" class="flex justify-center">
            <TurnstileWidget @callback="(tk) => (token = tk)" />
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
import { useApi, type AuthResponse } from "~/composables/useApi";

const toast = useToast();
const api = useApi();

const isLoading = ref(true);

const totalSteps = 3;
const currentStep = ref(1);
const isSubmitting = ref(false);

const stepOneRef = ref<{ isValid: boolean } | null>(null);
const stepTwoRef = ref<{ isValid: boolean } | null>(null);

const token = ref<string>("");
const termsAccepted = ref(false);
const termsDocs = [
    {
        key: "ku-work-core-terms",
        title: "KU Work Core Terms of Use and Privacy Notice",
        src: "/terms/ku_work_core_terms.txt",
    },
    {
        key: "privacy-policy",
        title: "KU Work Privacy Policy (Summary)",
        src: "/terms/privacy_policy.txt",
    },
    {
        key: "company-terms",
        title: "Company Terms of Use and Privacy Notice",
        src: "/terms/company_terms.txt",
    },
];

const form = reactive({
    companyName: "",
    companyEmail: "",
    companyWebsite: "",
    password: "",
    rPassword: "",
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
        return stepTwoRef.value.isValid && termsAccepted.value;
    }
    return false;
});

onMounted(() => {
    if (import.meta.client) {
        const token = localStorage.getItem("token");
        const role = localStorage.getItem("role");
        if (token) {
            if (role === "company" || role === "student") {
                navigateTo("/dashboard", { replace: true });
            } else if (role === "admin") {
                navigateTo("/admin/dashboard", { replace: true });
            } else if (role === "viewer") {
                navigateTo("/jobs", { replace: true });
            } else {
                navigateTo("/", { replace: true });
            }
        }
        isLoading.value = false;
    }
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
        if (!form.companyLogo || !form.banner) {
            return;
        }
        const formData = new FormData();
        formData.append("username", form.companyName);
        formData.append("email", form.companyEmail);
        formData.append("website", form.companyWebsite);
        formData.append("password", form.password);
        formData.append("phone", form.phone);
        formData.append("address", form.address);
        formData.append("city", form.city);
        formData.append("country", form.country);
        formData.append("about", form.about);
        formData.append("photo", form.companyLogo);
        formData.append("banner", form.banner);

        const response = await api.postFormData<AuthResponse>("/auth/company/register", formData, {
            headers: {
                "X-Turnstile-Token": token.value,
            },
        });

        if (response.data.token) {
            localStorage.setItem("token", response.data.token);
            localStorage.setItem("username", response.data.username as string);
            if (response.data.userId) {
                localStorage.setItem("userId", response.data.userId);
            }
            if (response.data.role) {
                localStorage.setItem("role", response.data.role);
            } else {
                localStorage.setItem("role", "company");
            }
        }

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
        description: "Registration complete. Redirecting to dashboard...",
        color: "success",
    });
    navigateTo("/dashboard", { replace: true });
};
</script>
