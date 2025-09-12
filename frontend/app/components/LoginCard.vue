<template>
    <div class="relative translate-y-[-2em] mb-10">
        <div class="flex flex-row gap-2 h-[7em] w-[28em] absolute left-1/2 -translate-x-1/2 top-0">
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
            class="h-[30em] w-[28em] rounded-xl bg-white pt-10 text-black relative top-[3.5em] mx-auto z-10"
        >
            <div v-if="!isCompany" class="flex flex-col h-full w-full">
                <h2 class="text-xl font-semibold mx-auto mb-5">KU Students/Staffs Login</h2>
                <GoogleOauthButton class="mx-auto text-white" />
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
</script>
