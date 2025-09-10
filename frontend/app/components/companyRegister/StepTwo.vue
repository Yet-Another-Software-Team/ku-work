<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="col-span-2 space-y-4">
                <div class="flex flex-col space-y-1">
                    <label class="text-primary-800 font-semibold">About Company</label>
                    <UTextarea
                        :model-value="about"
                        placeholder="Optional: Tell us about your company"
                        :error="!!errors.about"
                        :ui="{ base: 'h-[15em] rounded-lg bg-white text-black resize-none' }"
                        @update:model-value="updateAbout"
                    />
                    <div class="flex justify-between items-center">
                        <span v-if="errors.about" class="text-error text-sm">
                            {{ errors.about }}
                        </span>
                        <span class="text-gray-500 text-sm ml-auto">
                            {{ about.length }}/16,384 characters
                        </span>
                    </div>
                </div>
            </div>
            <div class="flex flex-col items-center justify-center">
                <div class="flex flex-col items-center gap-3">
                    <label class="text-primary-800 font-semibold"
                        >Company Logo * (JPEG, PNG - Max 5MB)</label
                    >
                    <button
                        class="size-[5em] rounded-full bg-gray-200 flex items-center justify-center text-4xl text-gray-500 outline-1 outline-primary overflow-hidden hover:cursor-pointer"
                        :class="{ 'border-2 border-error': errors.companyLogo }"
                        @click="logoFileInputRef?.click()"
                    >
                        <span v-if="!previewLogoUrl">+</span>
                        <img
                            v-else
                            :src="previewLogoUrl"
                            alt="Company Logo"
                            class="w-full h-full object-cover"
                        />
                    </button>
                    <span v-if="errors.companyLogo" class="text-error text-sm text-center">
                        {{ errors.companyLogo }}
                    </span>

                    <input
                        ref="logoFileInputRef"
                        type="file"
                        accept="image/jpeg,image/jpg,image/png"
                        class="hidden"
                        @change="onLogoFileChange"
                    />
                </div>
            </div>
        </div>

        <div>
            <div class="flex flex-col items-center justify-center">
                <div class="flex flex-col items-center gap-3 w-full">
                    <label class="text-primary-800 font-semibold w-full">
                        Banner Picture * (JPEG, PNG - Max 5MB)</label
                    >
                    <button
                        class="w-full h-[5em] rounded-lg bg-gray-200 flex items-center justify-center text-4xl text-gray-500 outline-1 outline-primary overflow-hidden hover:cursor-pointer"
                        :class="{ 'border-2 border-error': errors.banner }"
                        @click="bannerFileInputRef?.click()"
                    >
                        <span v-if="!previewBannerUrl">+</span>
                        <img
                            v-else
                            :src="previewBannerUrl"
                            alt="Company Banner"
                            class="w-full h-full object-cover"
                        />
                    </button>
                    <span v-if="errors.banner" class="text-error text-sm text-center">
                        {{ errors.banner }}
                    </span>

                    <input
                        ref="bannerFileInputRef"
                        type="file"
                        accept="image/jpeg,image/jpg,image/png"
                        class="hidden"
                        @change="onBannerFileChange"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onUnmounted } from "vue";
import * as z from "zod";

const FIVE_MB = 5 * 1024 * 1024;

const props = defineProps({
    about: {
        type: String,
        default: "",
    },
    companyLogo: {
        type: [File, null],
        default: null,
    },
    banner: {
        type: [File, null],
        default: null,
    },
});

const emit = defineEmits(["update:about", "update:companyLogo", "update:banner"]);

const previewLogoUrl = ref(null);
const previewBannerUrl = ref(null);
const logoFileInputRef = ref(null);
const bannerFileInputRef = ref(null);

const errors = reactive({
    about: "",
    companyLogo: "",
    banner: "",
});

const schema = z.object({
    about: z.string().max(16384, "About must be 16,384 characters or less").optional(),
});

const validateField = (fieldName, value) => {
    try {
        if (fieldName === "companyLogo") {
            if (!value) {
                errors.companyLogo = "Company logo is required";
                return false;
            }
            if (value.size > FIVE_MB) {
                errors.companyLogo = "File size must be less than 5MB";
                return false;
            }
            const allowedTypes = ["image/jpeg", "image/png", "image/jpg"];
            if (!allowedTypes.includes(value.type)) {
                errors.companyLogo = "Only JPEG, JPG, and PNG files are allowed";
                return false;
            }
            errors.companyLogo = "";
            return true;
        }
        if (fieldName === "banner") {
            if (!value) {
                errors.banner = "Banner picture is required";
                return false;
            }
            if (value.size > FIVE_MB) {
                errors.banner = "File size must be less than 5MB";
                return false;
            }
            const allowedTypes = ["image/jpeg", "image/png", "image/jpg"];
            if (!allowedTypes.includes(value.type)) {
                errors.banner = "Only JPEG, JPG, and PNG files are allowed";
                return false;
            }
            errors.banner = "";
            return true;
        }

        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });
        errors[fieldName] = "";
        return true;
    } catch (error) {
        if (error.errors?.[0]?.message) {
            errors[fieldName] = error.errors[0].message;
        } else {
            errors[fieldName] = "Invalid value";
        }
        return false;
    }
};

const isValid = computed(() => {
    const hasRequiredValues = props.companyLogo && props.banner;
    const hasNoErrors = !Object.values(errors).some((error) => error);
    return hasRequiredValues && hasNoErrors;
});

const updateAbout = (value) => {
    emit("update:about", value);
    validateField("about", value);
};

const onLogoFileChange = (event) => {
    const file = event.target.files[0];
    if (file) {
        if (previewLogoUrl.value) {
            URL.revokeObjectURL(previewLogoUrl.value);
        }
        previewLogoUrl.value = URL.createObjectURL(file);
        emit("update:companyLogo", file);
        validateField("companyLogo", file);
    }
};

const onBannerFileChange = (event) => {
    const file = event.target.files[0];
    if (file) {
        if (previewBannerUrl.value) {
            URL.revokeObjectURL(previewBannerUrl.value);
        }
        previewBannerUrl.value = URL.createObjectURL(file);
        emit("update:banner", file);
        validateField("banner", file);
    }
};

watch(
    () => props.companyLogo,
    (newLogo) => {
        if (newLogo) {
            if (previewLogoUrl.value) URL.revokeObjectURL(previewLogoUrl.value);
            previewLogoUrl.value = URL.createObjectURL(newLogo);
        } else {
            if (previewLogoUrl.value) URL.revokeObjectURL(previewLogoUrl.value);
            previewLogoUrl.value = null;
        }
    },
    { immediate: true }
);

watch(
    () => props.banner,
    (newBanner) => {
        if (newBanner) {
            if (previewBannerUrl.value) URL.revokeObjectURL(previewBannerUrl.value);
            previewBannerUrl.value = URL.createObjectURL(newBanner);
        } else {
            if (previewBannerUrl.value) URL.revokeObjectURL(previewBannerUrl.value);
            previewBannerUrl.value = null;
        }
    },
    { immediate: true }
);

onUnmounted(() => {
    if (previewLogoUrl.value) {
        URL.revokeObjectURL(previewLogoUrl.value);
    }
    if (previewBannerUrl.value) {
        URL.revokeObjectURL(previewBannerUrl.value);
    }
});

defineExpose({
    isValid,
});
</script>
