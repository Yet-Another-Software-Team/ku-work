<template>
    <UModal
        v-model:open="open"
        title="Job Application Form"
        :ui="{ content: 'w-[380px]', body: 'p-6', title: 'text-xl font-bold' }"
    >
        <!-- Button Trigger -->
        <UButton
            class="w-full justify-center my-5 p-2 text-xl"
            label="Apply"
            :disabled="appliedLocal"
        />

        <template #body>
            <!-- Progress -->
            <div class="flex flex-1 justify-center items-center gap-x-2 mb-4">
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
                    class="mt-4 block rounded-md border-2 border-dashed border-neutral-500/50 bg-neutral-50 p-6 text-center cursor-pointer transition light:hover:bg-neutral-100 dark:bg-neutral-800 dark:hover:bg-neutral-700"
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
                        <div
                            class="h-10 w-10 rounded-full bg-neutral-400 dark:bg-neutral-900 grid place-items-center"
                        >
                            <Icon name="basil:file-upload-outline" class="text-xl text-white" />
                        </div>
                        <div class="text-sm">Drop your resume here</div>
                        <p class="text-[10px]">PDF, PNG, JPG, JPEG (max 10MB each, 2 files)</p>
                        <p type="button" class="text-xs text-gray" @click="triggerPick"></p>
                    </div>
                </label>

                <!-- Selected files -->
                <div
                    v-for="(f, i) in files"
                    :key="i"
                    class="mt-2 flex items-center justify-between p-2 rounded-lg bg-gray-100 dark:bg-neutral-800 text-sm text-black"
                >
                    <div class="flex items-center">
                        <Icon name="material-symbols:description" class="text-primary-600 mr-2" />
                        <span class="text-neutral-900 dark:text-neutral-50">{{
                            truncateName(f.name, 25)
                        }}</span>
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
                        <div class="text-[13px] text-primary-800 font-semibold mb-1">
                            Phone <span class="text-gray-500 font-normal">(optional)</span>
                        </div>
                        <UInput
                            v-model="contactPhone"
                            type="tel"
                            placeholder="+66919999999"
                            :error="!!errors.contactPhone"
                            :ui="{ base: 'rounded-lg bg-white dark:bg-neutral-800 text-black' }"
                            @blur="validateField('contactPhone', contactPhone)"
                        />
                        <span v-if="errors.contactPhone" class="text-error text-sm">{{
                            errors.contactPhone
                        }}</span>
                    </div>
                    <div>
                        <div class="text-[13px] text-primary-800 font-semibold mb-1">
                            Contact Mail <span class="text-gray-500 font-normal">(optional)</span>
                        </div>
                        <UInput
                            v-model="contactMail"
                            type="email"
                            placeholder="sample@mail.com"
                            :error="!!errors.contactMail"
                            :ui="{ base: 'rounded-lg bg-white dark:bg-neutral-800 text-black' }"
                            @blur="validateField('contactMail', contactMail)"
                        />
                        <span v-if="errors.contactMail" class="text-error text-sm">{{
                            errors.contactMail
                        }}</span>
                    </div>
                    <TurnstileWidget @callback="(tk) => (cfToken = tk)" />
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
            </div>
        </template>

        <template v-if="currentStep === 1" #footer="{ close }">
            <div class="flex items-center gap-3 w-full">
                <UButton
                    label="Cancel"
                    color="neutral"
                    variant="outline"
                    class="flex-1 rounded-md text-neutral-900 bg-white justify-center hover:bg-gray-200 hover:cursor-pointer border-1 border-gray-200 px-4 py-2 font-md transition"
                    @click="close"
                />
                <UButton
                    :disabled="!isValid"
                    :label="isSubmitting ? 'Submitting...' : 'Next'"
                    :loading="isSubmitting"
                    class="flex-1 rounded-md justify-center px-4 py-2 font-semibold text-white disabled:opacity-50 disabled:cursor-not-allowed bg-primary-500 hover:bg-primary-700 transition"
                    @click="onNext"
                />
            </div>
        </template>

        <template v-else-if="currentStep === 2" #footer="{ close }">
            <UButton
                label="Done"
                class="w-full text-xl justify-center px-4 py-2 font-semibold text-white bg-primary-500 hover:bg-primary-700 transition"
                @click="handleDone(close)"
            />
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from "vue";
import * as z from "zod";

const api = useApi();

interface Props {
    jobId: number;
    applied?: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{ (e: "update:applied", value: boolean): void }>();
const toast = useToast();

// Local mutable applied state synced with prop for v-model support
const appliedLocal = ref<boolean>(props.applied);
watch(
    () => props.applied,
    (v) => {
        appliedLocal.value = v;
    }
);

const totalSteps = 2;
const currentStep = ref(1);

type FileLike = File & { previewUrl?: string };

const open = ref(false);
const files = ref<FileLike[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);
const isSubmitting = ref(false);

const contactPhone = ref("");
const contactMail = ref("");
const cfToken = ref("");

const errors = reactive({
    contactPhone: "",
    contactMail: "",
    resumeFile: "",
});

const schema = z.object({
    contactPhone: z
        .string()
        .refine(
            (v) => (v ? /^\+(?:[1-9]\d{0,2})\d{4,14}$/.test(v) : true),
            "Invalid phone number format"
        ),
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
    const hasFiles = files.value.length > 0;
    const noErrors = !Object.values(errors).some(Boolean);
    return hasFiles && noErrors && !isSubmitting.value;
});

const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

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
}

function truncateName(name: string, limit = 25): string {
    if (!name) return "";
    return name.length > limit ? name.slice(0, limit) + "..." : name;
}

async function onNext() {
    const okResume = validateField("resumeFile", null);

    if (!okResume) return;

    isSubmitting.value = true;

    try {
        const formData = new FormData();

        if (contactPhone.value.trim()) {
            formData.append("phone", contactPhone.value.trim());
        }

        if (contactMail.value.trim()) {
            formData.append("email", contactMail.value.trim());
        }

        files.value.forEach((file) => {
            console.log(file);
            formData.append("files", file);
        });

        await api.postFormData(`/jobs/${props.jobId}/apply`, formData, {
            withCredentials: true,
            headers: {
                "X-Turnstile-Token": cfToken.value,
            },
        });

        currentStep.value = 2;
        // update local state and emit update for v-model:applied
        appliedLocal.value = true;
        emit("update:applied", true);
        toast.add({
            title: "Apply Successfully",
            description: "Your application has been submitted successfully",
            color: "success",
        });
    } catch (error) {
        console.error("Application submission failed:", error);
        toast.add({
            title: "Application Failed",
            description: "Failed to submit application. Please try again.",
            color: "error",
        });
    } finally {
        isSubmitting.value = false;
    }
}

function handleDone(close: () => void) {
    contactPhone.value = "";
    contactMail.value = "";
    files.value = [];
    errors.contactPhone = "";
    errors.contactMail = "";
    errors.resumeFile = "";
    currentStep.value = 1;
    close();
}
</script>
