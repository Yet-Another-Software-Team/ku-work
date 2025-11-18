<template>
    <div class="relative translate-y-[-2em] mb-10">
        <div
            class="flex flex-row gap-2 h-[7em] w-[95vw] lg:w-[28em] absolute left-1/2 -translate-x-1/2 top-0"
        >
            <div
                class="hover:cursor-pointer transition-all duraion-150"
                :class="setTailwindClasses(false)"
                @click="selectRecruit"
            >
                <p class="font-bold px-5 py-5 text-2xl">KU Recruits</p>
            </div>
            <div
                class="hover:cursor-pointer transition-all duration-150"
                :class="setTailwindClasses(true)"
                @click="selectCompany"
            >
                <p class="font-bold px-5 py-5 text-2xl">Company</p>
            </div>
        </div>

        <div
            class="h-[30em] w-[95vw] lg:w-[28em] rounded-xl bg-white pt-10 text-black relative top-[3.5em] mx-auto z-10"
        >
            <div v-if="!isCompany" class="flex flex-col h-full w-full">
                <h2 class="text-xl font-semibold mx-auto mb-5">KU Students/Staffs Login</h2>
                <GoogleOauthButton class="mx-auto" />
                <div class="flex flex-col items-center pt-1 text-xs px-[10%]">
                    <p>By continuing, you agree to</p>
                    <p>
                        <UModal
                            v-model="termsModalOpen"
                            title="Terms of Service"
                            :ui="{ content: 'min-w-[72vw]' }"
                        >
                            <a
                                class="text-primary-600 font-semibold underline hover:text-primary-800 hover:cursor-pointer"
                                @click="openTerms"
                                >Terms of Service</a
                            >
                            <template #body>
                                <AgreementTermsOfService />
                            </template>
                        </UModal>
                        ,
                        <UModal
                            v-model="openOAuthPrivacy"
                            title="Google OAuth Privacy"
                            :ui="{ content: 'min-w-[72vw]' }"
                        >
                            <a
                                class="text-primary-600 font-semibold underline hover:text-primary-800 hover:cursor-pointer"
                                @click="openOAuthPrivacy"
                                >Google OAuth Privacy</a
                            >
                            <template #body>
                                <AgreementGoogleOAuth />
                            </template>
                        </UModal>
                        and
                        <UModal
                            v-model="privacyModalOpen"
                            title="Privacy Policy"
                            :ui="{ content: 'min-w-[76vw]' }"
                        >
                            <a
                                class="text-primary-600 font-semibold underline hover:text-primary-800 hover:cursor-pointer"
                                @click="openPrivacy"
                                >Privacy Policy</a
                            >
                            <template #body>
                                <AgreementPrivacy />
                            </template>
                        </UModal>
                    </p>
                </div>
            </div>
            <div v-else class="flex flex-col h-full w-full">
                <h2 class="text-xl font-semibold mx-auto mb-5">Company Login</h2>
                <CompanyLoginField />
                <p class="text-center w-full mt-auto mb-2">
                    Don't have an account?
                    <a
                        class="text-primary-600 font-semibold underline hover:text-primary-800"
                        href="/register/company"
                    >
                        Sign Up
                    </a>
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
const isCompany = ref(false);
const termsModalOpen = ref(false);
const privacyModalOpen = ref(false);
const oauthPrivacyModalOpen = ref(false);
const anyModalOpen = computed(
    () => termsModalOpen.value || privacyModalOpen.value || oauthPrivacyModalOpen.value
);

function setTailwindClasses(activeCondition) {
    if (isCompany.value == activeCondition) {
        return "bg-primary-200 flex flex-col rounded-3xl w-1/2 text-primary-800 hover:bg-primary-300";
    } else {
        return "bg-gray-200 flex flex-col rounded-3xl w-1/2 text-gray-500 hover:bg-gray-300";
    }
}

function selectCompany() {
    isCompany.value = true;
}

function selectRecruit() {
    isCompany.value = false;
}

const openPrivacy = () => {
    if (anyModalOpen.value) {
        return; // There is already a modal open
    }
    privacyModalOpen.value = true;
};

const openTerms = () => {
    if (anyModalOpen.value) {
        return; // There is already a modal open
    }
    termsModalOpen.value = true;
};

const openOAuthPrivacy = () => {
    if (anyModalOpen.value) {
        return; // There is already a modal open
    }
    oauthPrivacyModalOpen.value = true;
};
</script>
