<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="col-span-2 space-y-4">
                <div class="flex flex-col space-y-1">
                    <label class="text-primary-800 font-semibold">Full Name *</label>
                    <UInput
                        :model-value="fullName"
                        placeholder="John Doe"
                        size="xl"
                        disabled
                        :ui="{ base: 'rounded-lg bg-white text-black' }"
                        @update:model-value="updateFullName"
                    />
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div class="flex flex-col space-y-1">
                        <label class="text-primary-800 font-semibold">Date of Birth *</label>
                        <UInput
                            :model-value="dateOfBirth"
                            placeholder="MM/DD/YYYY"
                            icon="i-heroicons-calendar"
                            size="xl"
                            type="date"
                            :error="!!errors.dateOfBirth"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                            @update:model-value="updateDateOfBirth"
                        />
                        <span v-if="errors.dateOfBirth" class="text-error text-sm">
                            {{ errors.dateOfBirth }}
                        </span>
                    </div>
                    <div class="flex flex-col space-y-1">
                        <label class="text-primary-800 font-semibold">Phone</label>
                        <UInput
                            :model-value="phone"
                            placeholder="Optional: +66919999999"
                            size="xl"
                            :error="!!errors.phone"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                            @update:model-value="updatePhone"
                        />
                        <span v-if="errors.phone" class="text-error text-sm">
                            {{ errors.phone }}
                        </span>
                    </div>
                </div>
            </div>

            <div class="flex flex-col items-center justify-center">
                <div class="flex flex-col items-center gap-3">
                    <label class="text-primary-800 font-semibold"
                        >Profile Picture * <br />
                        <span class="font-normal text-sm">(JPEG, PNG, WEBP - Max 5MB)</span>
                    </label>
                    <button
                        class="size-[5em] rounded-full bg-gray-200 flex items-center justify-center text-4xl text-gray-500 outline-1 outline-primary overflow-hidden hover:cursor-pointer"
                        :class="{ 'border-2 border-error': errors.avatar }"
                        @click="fileInputRef?.click()"
                    >
                        <span v-if="!previewUrl">+</span>
                        <img
                            v-else
                            :src="previewUrl"
                            alt="Avatar"
                            class="w-full h-full object-cover"
                        />
                    </button>
                    <span v-if="errors.avatar" class="text-error text-sm text-center">
                        {{ errors.avatar }}
                    </span>

                    <input
                        ref="fileInputRef"
                        type="file"
                        accept="image/jpeg,image/jpg,image/png,image/webp"
                        class="hidden"
                        @change="onFileChange"
                    />
                </div>
            </div>
        </div>

        <div class="flex flex-col space-y-1">
            <label class="text-primary-800 font-semibold">About me</label>
            <UTextarea
                :model-value="aboutMe"
                placeholder="Optional: Tell us about yourself, your interests, skills, and career goals"
                :error="!!errors.aboutMe"
                :ui="{ base: 'h-[15em] rounded-lg bg-white text-black resize-none' }"
                @update:model-value="updateAboutMe"
            />
            <div class="flex justify-between items-center">
                <span v-if="errors.aboutMe" class="text-error text-sm">
                    {{ errors.aboutMe }}
                </span>
                <span class="text-gray-500 text-sm ml-auto">
                    {{ aboutMe.length }}/16,384 characters
                </span>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">GitHub URL</label>
                <UInput
                    type="url"
                    :model-value="githubURL"
                    placeholder="Optional: https://github.com/yourusername"
                    size="xl"
                    :error="!!errors.githubURL"
                    :ui="{ base: 'rounded-lg bg-white text-black' }"
                    @update:model-value="updateGithubURL"
                />
                <span v-if="errors.githubURL" class="text-error text-sm">
                    {{ errors.githubURL }}
                </span>
            </div>
            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">LinkedIn URL</label>
                <UInput
                    type="url"
                    :model-value="linkedinURL"
                    size="xl"
                    placeholder="Optional: https://linkedin.com/in/yourusername"
                    :error="!!errors.linkedinURL"
                    :ui="{ base: 'rounded-lg bg-white text-black' }"
                    @update:model-value="updateLinkedinURL"
                />
                <span v-if="errors.linkedinURL" class="text-error text-sm">
                    {{ errors.linkedinURL }}
                </span>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onUnmounted } from "vue";
import * as z from "zod";

const FIVE_MB = 5 * 1024 * 1024;

const props = defineProps({
    fullName: {
        type: String,
        default: "",
    },
    dateOfBirth: {
        type: String,
        default: "",
    },
    phone: {
        type: String,
        default: "",
    },
    aboutMe: {
        type: String,
        default: "",
    },
    githubURL: {
        type: String,
        default: "",
    },
    linkedinURL: {
        type: String,
        default: "",
    },
    avatar: {
        type: [File, null],
        default: null,
    },
});

const emit = defineEmits([
    "update:fullName",
    "update:dateOfBirth",
    "update:phone",
    "update:aboutMe",
    "update:githubURL",
    "update:linkedinURL",
    "update:avatar",
]);

const previewUrl = ref(null);
const fileInputRef = ref(null);

const errors = reactive({
    dateOfBirth: "",
    phone: "",
    aboutMe: "",
    githubURL: "",
    linkedinURL: "",
    avatar: "",
});

const schema = z.object({
    dateOfBirth: z.string().refine((date) => {
        if (!date) return true;
        const birthDate = new Date(date);
        const today = new Date();
        const age = today.getFullYear() - birthDate.getFullYear();
        return age >= 16 && age <= 80;
    }, "Age must be between 16 and 80 years"),
    phone: z
        .string()
        .regex(/^\+(?:[1-9]\d{0,2}) \d{4,14}$/, "Please enter a valid phone number")
        .optional(),
    aboutMe: z.string().max(16384, "About me must be 16,384 characters or less").optional(),
    githubURL: z
        .string()
        .max(256, "GitHub URL must be 256 characters or less")
        .optional()
        .refine((url) => {
            if (!url) return true;
            try {
                const validUrl = new URL(url);
                return validUrl.hostname.includes("github.com");
            } catch {
                return false;
            }
        }, "Must be a valid GitHub URL"),
    linkedinURL: z
        .string()
        .max(256, "LinkedIn URL must be 256 characters or less")
        .optional()
        .refine((url) => {
            if (!url) return true;
            try {
                const validUrl = new URL(url);
                return validUrl.hostname.includes("linkedin.com");
            } catch {
                return false;
            }
        }, "Must be a valid LinkedIn URL"),
});

const validateField = (fieldName, value) => {
    try {
        if (fieldName === "avatar") {
            if (!value) {
                errors.avatar = "Profile picture is required";
                return false;
            }
            if (value.size > FIVE_MB) {
                errors.avatar = "File size must be less than 5MB";
                return false;
            }
            const allowedTypes = ["image/jpeg", "image/png", "image/jpg", "image/webp"];
            if (!allowedTypes.includes(value.type)) {
                errors.avatar = "Only JPEG, JPG, WEBP and PNG files are allowed";
                return false;
            }
            errors.avatar = "";
            return true;
        }

        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });
        errors[fieldName] = "";
        return true;
    } catch (error) {
        if (error instanceof z.ZodError) {
            errors[fieldName] = error.issues[0]?.message || "Invalid value";
        } else {
            errors[fieldName] = "Unexpected Error Occurered";
        }
        return false;
    }
};

const isValid = computed(() => {
    const hasRequiredValues = props.avatar && props.dateOfBirth;
    const hasNoErrors = !Object.values(errors).some((error) => error);
    return hasRequiredValues && hasNoErrors;
});

const updateDateOfBirth = (value) => {
    emit("update:dateOfBirth", value);
    validateField("dateOfBirth", value);
};

const updatePhone = (value) => {
    emit("update:phone", value);
    validateField("phone", value);
};

const updateAboutMe = (value) => {
    emit("update:aboutMe", value);
    validateField("aboutMe", value);
};

const updateGithubURL = (value) => {
    emit("update:githubURL", value);
    validateField("githubURL", value);
};

const updateLinkedinURL = (value) => {
    emit("update:linkedinURL", value);
    validateField("linkedinURL", value);
};

const onFileChange = (event) => {
    const file = event.target.files[0];
    if (file && file.size <= FIVE_MB) {
        if (previewUrl.value) {
            URL.revokeObjectURL(previewUrl.value);
        }
        previewUrl.value = URL.createObjectURL(file);
        emit("update:avatar", file);
        validateField("avatar", file);
    }
};

watch(
    () => props.avatar,
    (newAvatar, oldAvatar) => {
        if (oldAvatar && previewUrl.value) {
            URL.revokeObjectURL(previewUrl.value);
        }
        previewUrl.value = newAvatar ? URL.createObjectURL(newAvatar) : null;
    },
    { immediate: true }
);

onUnmounted(() => {
    if (previewUrl.value) {
        URL.revokeObjectURL(previewUrl.value);
    }
});

defineExpose({
    isValid,
});
</script>
