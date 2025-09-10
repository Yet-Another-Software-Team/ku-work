<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">Student ID *</label>
                <div class="relative">
                    <input
                        :value="studentId"
                        type="text"
                        placeholder="Enter your Student ID"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.studentId }"
                        @input="updateStudentId"
                    />
                </div>
                <span v-if="errors.studentId" class="text-error text-sm">
                    {{ errors.studentId }}
                </span>
            </div>

            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">Major *</label>
                <div class="relative">
                    <select
                        :value="major"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                        :class="{ 'border-error': errors.major }"
                        @change="updateMajor"
                    >
                        <option value="" disabled>Select your Major</option>
                        <option value="Computer Engineering">Computer Engineering</option>
                        <option value="Software and Knowledge Engineering">
                            Software and Knowledge Engineering
                        </option>
                    </select>
                    <div
                        class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none"
                    >
                        <svg
                            class="h-5 w-5 text-gray-400"
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 20 20"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
                                clip-rule="evenodd"
                            />
                        </svg>
                    </div>
                </div>
                <span v-if="errors.major" class="text-error text-sm">
                    {{ errors.major }}
                </span>
            </div>

            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">Student Status *</label>
                <div class="relative">
                    <select
                        :value="studentStatus"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                        :class="{ 'border-red-500': errors.studentStatus }"
                        @change="updateStudentStatus"
                    >
                        <option value="" disabled>Select Student Status</option>
                        <option value="Current Student">Current Student</option>
                        <option value="Graduated">Graduated</option>
                    </select>
                    <div
                        class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none"
                    >
                        <svg
                            class="h-5 w-5 text-gray-400"
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 20 20"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
                                clip-rule="evenodd"
                            />
                        </svg>
                    </div>
                </div>
                <span v-if="errors.studentStatus" class="text-error text-sm">
                    {{ errors.studentStatus }}
                </span>
            </div>

            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">Verify Student Status *</label>
                <div>
                    <button
                        type="button"
                        class="w-full flex items-center justify-center px-4 py-3 text-white bg-green-500 border border-green-500 rounded-lg shadow-sm font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
                        :class="{
                            'border-red-500 bg-red-500 hover:bg-red-600': errors.verificationFile,
                        }"
                        @click="fileInputRef?.click()"
                    >
                        <svg
                            class="h-6 w-6 mr-2"
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 20 20"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M10 1a1 1 0 011 1v6h6a1 1 0 110 2h-6v6a1 1 0 11-2 0v-6H3a1 1 0 110-2h6V2a1 1 0 011-1z"
                                clip-rule="evenodd"
                            />
                        </svg>
                        {{ verificationFileName ? "Change File" : "Upload Document" }}
                    </button>
                    <input
                        ref="fileInputRef"
                        type="file"
                        accept="application/pdf,image/jpeg,image/jpg,image/png"
                        class="hidden"
                        @change="onFileChange"
                    />
                </div>
                <div class="text-sm text-gray-600">
                    Upload transcript, student ID card, or graduation certificate (PDF, JPEG, PNG
                     - Max 10MB)
                </div>
                <span v-if="errors.verificationFile" class="text-error text-sm">
                    {{ errors.verificationFile }}
                </span>
                <div
                    v-if="verificationFileName"
                    class="mt-2 flex items-center justify-between p-2 rounded-lg bg-gray-100 text-sm text-black"
                >
                    <div class="flex items-center">
                        <icon name="material-symbols:description" class="text-primary-600 mr-2" />
                        <span>{{ verificationFileName }}</span>
                        <span class="ml-2 text-gray-500">({{ formatFileSize(fileSize) }})</span>
                    </div>
                    <UIcon
                        name="material-symbols:close"
                        class="text-gray-500 hover:text-red-700 cursor-pointer"
                        @click="removeFile"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from "vue";
import * as z from "zod";

const toast = useToast();

const props = defineProps({
    studentId: {
        type: String,
        default: "",
    },
    major: {
        type: String,
        default: "",
    },
    studentStatus: {
        type: String,
        default: "",
    },
    verificationFile: {
        type: [File, null],
        default: null,
    },
});

const emit = defineEmits([
    "update:studentId",
    "update:major",
    "update:studentStatus",
    "update:verificationFile",
]);

const verificationFileName = ref<string | null>(null);
const fileSize = ref<number>(0);
const fileInputRef = ref<HTMLInputElement | null>(null);

const errors = reactive({
    studentId: "",
    major: "",
    studentStatus: "",
    verificationFile: "",
});

const schema = z.object({
    studentId: z
        .string()
        .length(10, "Student ID must be exactly 10 characters")
        .regex(/^[A-Za-z0-9-]+$/, "Student ID can only contain letters, numbers, and hyphens"),
    major: z
        .string()
        .refine(
            (value) =>
                value === "Software and Knowledge Engineering" || value === "Computer Engineering",
            "Please select a valid major"
        ),
    studentStatus: z
        .string()
        .refine(
            (value) => value === "Graduated" || value === "Current Student",
            "Please select a valid student status"
        ),
});

const validateField = (fieldName: keyof typeof errors, value: unknown) => {
    try {
        if (fieldName === "verificationFile") {
            if (!value) {
                errors.verificationFile = "Verification file is required";
                return false;
            }
            const file = value as File;
            if (file.size > 10 * 1024 * 1024) {
                errors.verificationFile = "File size must be less than 10MB";
                return false;
            }
            const allowedTypes = [
                "application/pdf",
                "image/jpeg",
                "image/png",
                "image/jpg",
            ];
            if (!allowedTypes.includes(file.type)) {
                errors.verificationFile = "Only PDF, JPEG, JPG, and PNG files are allowed";
                return false;
            }
            errors.verificationFile = "";
            return true;
        }

        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });
        errors[fieldName] = "";
        return true;
    } catch (error: unknown) {
        if (error && typeof error === "object" && "errors" in error) {
            const zodError = error as { errors: Array<{ message: string }> };
            errors[fieldName] = zodError.errors[0]?.message || "Invalid value";
        } else {
            errors[fieldName] = "Invalid value";
        }
        return false;
    }
};

const isValid = computed(() => {
    const hasValues =
        props.studentId && props.major && props.studentStatus && props.verificationFile;
    const hasNoErrors = !Object.values(errors).some((error) => error);
    return hasValues && hasNoErrors;
});

const updateStudentId = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:studentId", value);
    if (value) validateField("studentId", value);
};

const updateMajor = (event: Event) => {
    const value = (event.target as HTMLSelectElement).value;
    emit("update:major", value);
    if (value) validateField("major", value);
};

const updateStudentStatus = (event: Event) => {
    const value = (event.target as HTMLSelectElement).value;
    emit("update:studentStatus", value);
    if (value) validateField("studentStatus", value);
};

const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const onFileChange = (event: Event) => {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) {
        verificationFileName.value = file.name;
        fileSize.value = file.size;
        emit("update:verificationFile", file);
        validateField("verificationFile", file);
    }
};

watch(
    () => props.verificationFile,
    (newFile) => {
        if (newFile) {
            verificationFileName.value = newFile.name;
            fileSize.value = newFile.size;
        } else {
            verificationFileName.value = null;
            fileSize.value = 0;
        }
    },
    { immediate: true }
);

const removeFile = () => {
    verificationFileName.value = null;
    fileSize.value = 0;
    emit("update:verificationFile", null);
    errors.verificationFile = "Verification file is required";

    if (fileInputRef.value) {
        fileInputRef.value.value = "";
    }

    toast.add({
        title: "File removed",
        description: "Verification document has been removed",
        color: "warning",
    });
};

defineExpose({
    isValid,
});
</script>
