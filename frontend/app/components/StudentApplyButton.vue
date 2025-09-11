<template>
    <!-- Button -->
    <!-- Need change when apply to job board -->
    <UButton
        class="size-fit text-xl rounded-md px-10 gap-2 font-small py-3 bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
        label="Apply"
        @click="openForm"
    />

    <!-- Overlay -->
    <div
        v-if="open"
        class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
        @keydown.esc="closeForm"
    >
        <!-- Modal -->
        <div class="w-[380px] rounded-2xl p-6 shadow-lg bg-white text-gray-900 dark:text-white">
            <div class="flex justify-between items-center mb-4">
                <h1 class="text-xl md:text-3xl font-bold text-left text-black">
                    Job Application Form
                </h1>
            </div>

            <!-- Progress -->
            <div class="flex flex-1 justify-center items-center gap-x-2">
                <UProgress
                    v-for="i in totalSteps"
                    :key="i"
                    :model-value="currentStep >= i ? 100 : 0"
                    size="lg"
                />
            </div>

            <!-- Step 1 -->
            <div v-if="currentStep === 1">
                <!-- Dropzone -->
                <label
                    class="mt-4 block rounded-md border-2 border-dashed border-neutral-500/50 bg-neutral-50 p-6 text-center cursor-pointer transition hover:bg-neutral-100"
                    @dragover.prevent
                    @drop="onDrop"
                >
                    <input
                        ref="fileInput"
                        type="file"
                        class="hidden"
                        accept="application/pdf,image/png,image/jpeg,image/jpg"
                        multiple
                        @change="onPick"
                    />
                    <div class="mx-auto grid place-items-center gap-2">
                        <div class="h-10 w-10 rounded-full bg-neutral-400 grid place-items-center">
                            <Icon name="basil:file-upload-outline" class="text-xl text-white" />
                        </div>
                        <div class="text-sm">Drop your resume here</div>
                        <p class="text-[10px] text-neutral-400">
                            PDF, PNG, JPG, JPEG (max 10MB each, 2 files)
                        </p>
                        <p type="button" class="text-xs text-gray" @click="triggerPick"></p>
                    </div>
                </label>

                <!-- Selected files -->
                <div
                    v-for="(f, i) in files"
                    :key="i"
                    class="mt-2 flex items-center justify-between p-2 rounded-lg bg-gray-100 text-sm text-black"
                >
                    <div class="flex items-center">
                        <Icon name="material-symbols:description" class="text-primary-600 mr-2" />
                        <span>{{ truncateName(f.name, 25) }}</span>
                        <span class="ml-2 text-gray-500">({{ formatFileSize(f.size) }})</span>
                    </div>
                    <UIcon
                        name="material-symbols:close"
                        class="text-gray-500 hover:text-red-700 cursor-pointer"
                        @click="removeFile(i)"
                    />
                </div>
                <span v-if="errors.resumeFile" class="text-error text-sm mt-1 block">{{
                    errors.resumeFile
                }}</span>

                <!-- Inputs -->
                <div class="mt-4 grid grid-cols-2 gap-4">
                    <div>
                        <div class="text-[13px] text-primary-800 font-semibold mb-1">Phone</div>
                        <UInput
                            v-model="contactPhone"
                            type="tel"
                            placeholder="0919999999"
                            :error="!!errors.contactPhone"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                            @blur="validateField('contactPhone', contactPhone)"
                        />
                        <span v-if="errors.contactPhone" class="text-error text-sm">{{
                            errors.contactPhone
                        }}</span>
                    </div>
                    <div>
                        <div class="text-[13px] text-primary-800 font-semibold mb-1">
                            Contact Mail
                        </div>
                        <UInput
                            v-model="contactMail"
                            type="email"
                            placeholder="sample@mail.com"
                            :error="!!errors.contactMail"
                            :ui="{ base: 'rounded-lg bg-white text-black' }"
                            @blur="validateField('contactMail', contactMail)"
                        />
                        <span v-if="errors.contactMail" class="text-error text-sm">{{
                            errors.contactMail
                        }}</span>
                    </div>
                </div>

                <!-- Actions -->
                <div class="mt-5 flex items-center gap-3">
                    <UButton
                        label="Cancel"
                        color="neutral"
                        variant="outline"
                        class="flex-1 rounded-md bg-white justify-center hover:bg-gray-200 hover:cursor-pointer border-1 border-gray-200 px-4 py-2 font-md transition"
                        @click="closeForm"
                    />
                    <UButton
                        :disabled="!isValid"
                        label="Next"
                        class="flex-1 rounded-md justify-center px-4 py-2 font-semibold text-white disabled:opacity-50 disabled:cursor-not-allowed bg-primary-500 hover:bg-primary-700 transition"
                        @click="onNext"
                    />
                </div>
            </div>

            <!-- Step 2 -->
            <div v-else-if="currentStep === 2" class="text-center">
                <Icon
                    name="hugeicons:checkmark-circle-03"
                    class="text-[12em] text-primary mb-4 mt-8"
                />
                <h2 class="text-3xl font-bold text-gray-600 mb-2">Apply Success</h2>
                <p class="text-gray-600">You have applied to this job application.</p>

                <UButton
                    label="Done"
                    class="mt-5 w-full text-xl justify-center px-4 py-2 font-semibold text-white bg-primary-500 hover:bg-primary-700 transition"
                    @click="handleDone"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from "vue";
import * as z from "zod";

const toast = useToast();

const totalSteps = 2;
const currentStep = ref(1);

type FileLike = File & { previewUrl?: string };

const open = ref(false);
const files = ref<FileLike[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);

const contactPhone = ref("");
const contactMail = ref("");

const errors = reactive({
    contactPhone: "",
    contactMail: "",
    resumeFile: "",
});

const schema = z.object({
    contactPhone: z
        .string()
        .max(20, "Phone number must be 20 characters or less")
        .refine((v) => (v ? /^[+]?[0-9\-()\s]+$/.test(v) : true), "Invalid phone number format"),
    contactMail: z
        .string()
        .refine((v) => (v ? /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) : true), "Invalid email format"),
});

function validateField(fieldName: keyof typeof errors, value: unknown) {
    if (fieldName === "resumeFile") {
        if (files.value.length === 0) {
            errors.resumeFile = "Resume file is required";
            return false;
        }
        errors.resumeFile = "";
        return true;
    }

    try {
        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });
        if (typeof value === "string" && value.trim() === "") {
            errors[fieldName] = "This field is required";
            return false;
        }
        errors[fieldName] = "";
        return true;
    } catch (error: unknown) {
        if (error instanceof z.ZodError) {
            errors[fieldName] = error.issues[0]?.message ?? "Invalid value";
        } else {
            errors[fieldName] = "Invalid value";
        }
        return false;
    }
}

watch(contactPhone, (v) => validateField("contactPhone", v));
watch(contactMail, (v) => validateField("contactMail", v));

const isValid = computed(() => {
    const filled = contactPhone.value.trim() && contactMail.value.trim() && files.value.length > 0;
    const noErrors = !Object.values(errors).some(Boolean);
    return !!filled && noErrors;
});

const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

function openForm() {
    open.value = true;
}
function closeForm() {
    open.value = false;
}

function triggerPick() {
    fileInput.value?.click();
}

function onPick(e: Event) {
    const target = e.target as HTMLInputElement;
    addFiles(Array.from(target.files ?? []));
    if (target) target.value = "";
}

function onDrop(e: DragEvent) {
    e.preventDefault();
    addFiles(Array.from(e.dataTransfer?.files ?? []));
}

function addFiles(incoming: File[]) {
    let blockedReason = "";
    for (const f of incoming) {
        const okType = /\/(pdf|jpeg|jpg|png)$/i.test(f.type);
        const okSize = f.size <= 10 * 1024 * 1024;
        if (!okType) blockedReason = "Only PDF, PNG, JPG, and JPEG files are allowed";
        if (!okSize) blockedReason = "File size must be less than 10MB";
        if (okType && okSize && files.value.length < 2) {
            files.value.push(f as FileLike);
        }
    }
    if (files.value.length === 0) {
        errors.resumeFile = blockedReason || "Resume file is required";
    } else {
        errors.resumeFile = "";
    }
}

function removeFile(idx: number) {
    files.value.splice(idx, 1);
    if (files.value.length === 0) errors.resumeFile = "Resume file is required";
    toast.add({
        title: "File removed",
        description: "Resume has been removed",
        color: "warning",
    });
}

function truncateName(name: string, limit = 25): string {
    if (!name) return "";
    return name.length > limit ? name.slice(0, limit) + "..." : name;
}

function onNext() {
    const okPhone = validateField("contactPhone", contactPhone.value);
    const okMail = validateField("contactMail", contactMail.value);
    const okResume = validateField("resumeFile", null);

    if (!(okPhone && okMail && okResume)) return;

    currentStep.value = 2;
    toast.add({
        title: "Apply Successfully",
        description: "Your application response will be sent to your contact email",
        color: "success",
    });
}

function handleDone() {
    contactPhone.value = "";
    contactMail.value = "";
    files.value = [];
    errors.contactPhone = "";
    errors.contactMail = "";
    errors.resumeFile = "";
    currentStep.value = 1;
    closeForm();
}
</script>
