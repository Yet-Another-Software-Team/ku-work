<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Student ID field -->
            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">Student ID</label>
                <div class="relative">
                    <input
                        :value="props.studentId"
                        type="text"
                        placeholder="Student ID"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                    />
                </div>
            </div>

            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">Major</label>
                <div class="relative">
                    <select
                        :value="props.major"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                    >
                        <option value="" disabled selected>Major</option>
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
            </div>

            <!-- Student Status dropdown -->
            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">Student Status</label>
                <div class="relative">
                    <select
                        :value="studentStatus"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                    >
                        <option value="" disabled selected>Student Status</option>
                        <option value="Alumni">Graduated</option>
                        <option value="Current Student">Current Student</option>
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
            </div>

            <!-- Verify Student Status file upload -->
            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">Verify Student Status</label>
                <div>
                    <button
                        type="button"
                        class="w-full flex items-center justify-center px-4 py-3 text-white bg-green-500 border border-green-500 rounded-lg shadow-sm font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
                        @click="$refs.fileInput.click()"
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
                        Upload File
                    </button>
                    <input
                        ref="fileInput"
                        type="file"
                        accept="application/pdf,image/*"
                        class="hidden"
                        @change="onFileChange"
                    />
                </div>
                <div
                    v-if="verificationFileName"
                    class="mt-2 flex items-center justify-between p-2 rounded-lg bg-gray-100 text-sm text-black"
                >
                    <span>{{ verificationFileName }}</span>
                    <UIcon
                        name="material-symbols:close"
                        class="text-gray-500 hover:text-red-700"
                        @click="removeFile"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

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

const emit = defineEmits(["update:verificationFile"]);

const verificationFileName = ref<string | null>(null);

const onFileChange = (event: Event) => {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) {
        verificationFileName.value = file.name;
        emit("update:verificationFile", file);
    }
};

watch(
    () => props.verificationFile,
    (newFile) => {
        if (newFile) {
            verificationFileName.value = newFile.name;
        } else {
            verificationFileName.value = null;
        }
    },
    { immediate: true }
);

const removeFile = () => {
    verificationFileName.value = null;
    emit("update:verificationFile", null);
};
</script>
