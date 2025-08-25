<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="col-span-2 space-y-4">
                <div class="flex flex-col space-y-1">
                    <label class="text-green-800 font-semibold">Full Name</label>
                    <UInput
                        :model-value="fullName"
                        placeholder="John Doe"
                        size="xl"
                        disabled
                        :ui="{ base: 'rounded-lg bg-white text-black' }"
                    />
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div class="flex flex-col space-y-1">
                        <label class="text-green-800 font-semibold">Date of Birth</label>
                        <UInput
                            :model-value="dateOfBirth"
                            placeholder="MM/DD/YYYY"
                            icon="i-heroicons-calendar"
                            size="xl"
                            type="date"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                        />
                    </div>
                    <div class="flex flex-col space-y-1">
                        <label class="text-green-800 font-semibold">Phone</label>
                        <UInput
                            :model-value="phone"
                            placeholder="Phone"
                            size="xl"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                        />
                    </div>
                </div>
            </div>

            <div class="flex flex-col items-center justify-center">
                <div class="flex items-center gap-3">
                    <button
                        class="size-[5em] rounded-full bg-gray-200 flex items-center justify-center text-4xl text-gray-500 outline-1 outline-primary overflow-hidden hover:cursor-pointer"
                        @click="$refs.fileInput.click()"
                    >
                        <span v-if="!previewUrl">+</span>
                        <img
                            v-else
                            :src="previewUrl"
                            alt="Avatar"
                            class="w-full h-full object-cover"
                        />
                    </button>

                    <input
                        ref="fileInput"
                        type="file"
                        accept="image/*"
                        class="hidden"
                        @change="onFileChange"
                    />
                </div>
            </div>
        </div>

        <div class="flex flex-col space-y-1">
            <label class="text-green-800 font-semibold">About me</label>
            <UTextarea
                :model-value="aboutMe"
                placeholder="About me"
                :ui="{ base: 'h-[15em] rounded-lg bg-white text-black resize-none' }"
            />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">Github</label>
                <UInput
                    type="url"
                    :model-value="githubURL"
                    placeholder="Github URL"
                    size="xl"
                    :ui="{ base: 'rounded-lg bg-white text-black' }"
                />
            </div>
            <div class="flex flex-col space-y-1">
                <label class="text-green-800 font-semibold">LinkedIn</label>
                <UInput
                    type="url"
                    :model-value="linkedinURL"
                    size="xl"
                    placeholder="LinkedIn URL"
                    :ui="{ base: 'rounded-lg bg-white text-black' }"
                />
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from "vue";

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

const onFileChange = (event) => {
    const file = event.target.files[0];
    if (file) {
        if (previewUrl.value) {
            URL.revokeObjectURL(previewUrl.value);
        }
        previewUrl.value = URL.createObjectURL(file);
        emit("update:avatar", file);
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
</script>
