<template>
    <div class="rounded-xl dark:bg-[#001F26] max-h-[90vh] overflow-y-auto p-5 w-full max-w-2xl">
        <div class="w-full flex justify-center mb-10">
            <div class="relative">
                <div
                    class="w-24 h-24 sm:w-28 sm:h-28 rounded-full bg-gray-200 overflow-hidden shadow"
                >
                    <img
                        v-if="avatarPreview"
                        :src="avatarPreview"
                        alt="Avatar"
                        class="w-full h-full object-cover"
                    />
                    <Icon
                        v-else
                        name="ic:baseline-account-circle"
                        class="w-full h-full text-gray-400"
                    />
                </div>
                <button
                    type="button"
                    class="absolute -right-1 bottom-0 translate-y-1 inline-flex items-center justify-center w-9 h-9 rounded-full bg-primary-500 text-white shadow hover:bg- primary-600 ring-4 ring-white/60"
                    aria-label="Change avatar"
                    @click="triggerAvatarPicker"
                >
                    <Icon name="material-symbols:edit-square-outline-rounded" class="w-5 h-5" />
                </button>
                <input
                    ref="avatarInput"
                    type="file"
                    accept="image/jpeg,image/jpg,image/png"
                    class="hidden"
                    @change="onAvatarSelected"
                />
            </div>
        </div>

        <form class="grid grid-cols-1 md:grid-cols-2 gap-4" @submit.prevent="handleSubmit">
            <div class="md:col-span-2">
                <label class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >Name</label
                >
                <div
                    class="rounded-lg border border- primary-700/50 bg-gray-100 px-4 py-2 text-gray-900 dark:border- primary-700/40 dark:bg-[#013B49] dark:text-white"
                >
                    {{ `${profile.firstName} ${profile.lastName}` }}
                </div>
            </div>

            <div class="col-span-1">
                <label for="dob" class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >Date of Birth</label
                >
                <UInput
                    id="dob"
                    v-model="form.dob"
                    type="date"
                    icon="i-heroicons-calendar-20-solid"
                    placeholder="Birth date"
                    class="w-full rounded-lg bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
                />
                <p v-if="errors.dob" class="mt-1 text-sm text-red-500">{{ errors.dob }}</p>
            </div>

            <div class="col-span-1">
                <label
                    for="phone"
                    class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >Phone</label
                >
                <UInput
                    id="phone"
                    v-model="form.phone"
                    placeholder="Optional: +66919999999"
                    class="w-full rounded-lg bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
                />
                <p v-if="errors.phone" class="mt-1 text-sm text-red-500">{{ errors.phone }}</p>
            </div>

            <div class="col-span-1">
                <label
                    for="github"
                    class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >GitHub</label
                >
                <UInput
                    id="github"
                    v-model="form.github"
                    placeholder="Optional: https://github.com/username"
                    class="w-full rounded-lg bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
                />
                <p v-if="errors.github" class="mt-1 text-sm text-red-500">{{ errors.github }}</p>
            </div>

            <div class="col-span-1">
                <label
                    for="linkedin"
                    class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >LinkedIn</label
                >
                <UInput
                    id="linkedin"
                    v-model="form.linkedin"
                    placeholder="Optional: https://linkedin.com/in/username"
                    class="w-full rounded-lg bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
                />
                <p v-if="errors.linkedin" class="mt-1 text-sm text-red-500">
                    {{ errors.linkedin }}
                </p>
            </div>

            <div class="md:col-span-2">
                <label
                    for="aboutMe"
                    class="block text-primary-800 dark:text-primary font-semibold mb-1"
                    >About me</label
                >
                <div class="rounded-lg bg-white dark:bg-[#013B49]">
                    <UTextarea
                        id="aboutMe"
                        v-model="form.aboutMe"
                        placeholder="Optional: Tell us about yourself"
                        :rows="6"
                        class="w-full bg-transparent border-0 focus:outline-none resize-none text-gray-900 dark:text-white"
                    />
                </div>
            </div>

            <div class="md:col-span-2 flex flex-wrap justify-end gap-3 pt-2 w-full">
                <UButton
                    type="button"
                    variant="outline"
                    color="neutral"
                    class="rounded-md px-4"
                    @click="tryDiscard"
                >
                    Discard
                </UButton>
                <UButton type="submit" color="primary" class="rounded-md px-5"> Save </UButton>
            </div>
        </form>

        <UModal
            v-model:open="showDiscardConfirm"
            title="Discard changes?"
            :dismissible="false"
            :ui="{
                title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                overlay: 'fixed inset-0 bg-black/50',
            }"
        >
            <template #body>
                <p class="dark:text-white">This will discard your current inputs. Are you sure?</p>
            </template>
            <template #footer>
                <div class="flex justify-end gap-2">
                    <UButton variant="outline" color="neutral" @click="showDiscardConfirm = false"
                        >Cancel</UButton
                    >
                    <UButton color="primary" @click="confirmDiscard">Discard</UButton>
                </div>
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from "vue";
import * as z from "zod";

const { add: addToast } = useToast();

interface StudentProfile {
    name?: string;
    birthDate?: string;
    phone?: string;
    github?: string;
    linkedIn?: string;
    aboutMe?: string;
    photo?: string;
}

interface FormState {
    dob: string;
    phone: string;
    github: string;
    linkedin: string;
    aboutMe: string;
}

type FormKey = keyof FormState;

type SavedPayload = StudentProfile & { _avatarFile?: File | null };

const props = defineProps<{
    profile: {
        photo: string;
        birthDate: string;
        phone: string;
        major: string;
        linkedIn: string;
        github: string;
        aboutMe: string;
        firstName: string;
        lastName: string;
        email: string;
    };
}>();

const emit = defineEmits<{
    (e: "close"): void;
    (e: "saved", payload: SavedPayload): SavedPayload;
}>();

const form = reactive<FormState>({
    dob: "",
    phone: "",
    github: "",
    linkedin: "",
    aboutMe: "",
});

const errors = reactive<Record<FormKey, string>>({
    dob: "",
    phone: "",
    github: "",
    linkedin: "",
    aboutMe: "",
});

const schema = z.object({
    dob: z.string().min(1, "Date of birth is required"),
    phone: z
        .string()
        .optional()
        .refine(
            (value) => !value || /^\+(?:[1-9]\d{0,2}) \d{4,14}$/.test(value),
            "Please enter a valid phone number"
        ),
    github: z
        .string()
        .max(256, "GitHub URL must be 256 characters or less")
        .optional()
        .refine((url) => !url || isHost(url, "github.com"), "Must be a valid GitHub URL"),
    linkedin: z
        .string()
        .max(256, "LinkedIn URL must be 256 characters or less")
        .optional()
        .refine((url) => !url || isHost(url, "linkedin.com"), "Must be a valid LinkedIn URL"),
    aboutMe: z.string().max(16384, `About me must be 16384 characters or less`).optional(),
});

const avatarInput = ref<HTMLInputElement | null>(null);
const avatarPreview = ref<string>("");
const avatarFile = ref<File | null>(null);
const showDiscardConfirm = ref(false);

watch(
    () => props.profile,
    (profile) => resetForm(profile),
    { immediate: true, deep: true }
);

function resetForm(profile: StudentProfile) {
    console.log(profile.birthDate!);
    form.dob = (profile.birthDate ?? new Date(Date.now()).toISOString()).split("T")[0]!;
    form.phone = profile.phone ?? "";
    form.github = profile.github ?? "";
    form.linkedin = profile.linkedIn ?? "";
    form.aboutMe = profile.aboutMe ?? "";

    updateAvatarPreview(profile.photo ?? "");
    avatarFile.value = null;
    clearErrors();
    runAllValidations();
}

function runAllValidations() {
    (Object.keys(form) as FormKey[]).forEach((key) => validateField(key, form[key]));
}

function clearErrors() {
    (Object.keys(errors) as FormKey[]).forEach((key) => (errors[key] = ""));
}

function validateField(field: FormKey, value: string) {
    const result = (schema.shape[field] as z.ZodTypeAny).safeParse(value);
    errors[field] = result.success ? "" : (result.error.issues[0]?.message ?? "Invalid value");
    return result.success;
}

(Object.keys(form) as FormKey[]).forEach((key) => {
    watch(
        () => form[key],
        (value) => validateField(key, value)
    );
});

function triggerAvatarPicker() {
    avatarInput.value?.click();
}

function onAvatarSelected(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const isGif = file.type === "image/gif" || file.name.toLowerCase().endsWith(".gif");
    if (isGif) {
        addToast({
            title: "Unsupported image",
            description: "GIF images are not supported",
            color: "warning",
        });
        (event.target as HTMLInputElement).value = "";
        return;
    }
    updateAvatarPreview(URL.createObjectURL(file));
    avatarFile.value = file;
}

function updateAvatarPreview(source: string) {
    if (avatarPreview.value?.startsWith("blob:")) {
        URL.revokeObjectURL(avatarPreview.value);
    }
    avatarPreview.value = source;
}

function tryDiscard() {
    if (
        form.dob !==
            (props.profile.birthDate ?? new Date(Date.now()).toISOString()).split("T")[0]! ||
        form.github !== props.profile.github ||
        form.phone !== props.profile.phone ||
        form.linkedin !== props.profile.linkedIn ||
        form.aboutMe !== props.profile.aboutMe ||
        avatarFile.value != null
    ) {
        showDiscardConfirm.value = true;
        return;
    }
    confirmDiscard();
}

function confirmDiscard() {
    showDiscardConfirm.value = false;
    resetForm(props.profile);
    addToast({ title: "Changes discarded", description: "Old data reloaded.", color: "success" });
    emit("close");
}

function handleSubmit() {
    const result = schema.safeParse({ ...form });
    if (!result.success) {
        result.error.issues.forEach((issue) => {
            const key = issue.path[0];
            if (typeof key === "string" && key in errors) {
                errors[key as FormKey] = issue.message;
            }
        });
        addToast({
            title: "Form submission failed",
            description: "Please check the highlighted errors and try again.",
            color: "warning",
        });
        return;
    }

    addToast({ title: "Saved", description: "Profile updated successfully.", color: "success" });

    emit("saved", {
        ...props.profile,
        birthDate: result.data.dob,
        phone: result.data.phone ?? "",
        github: result.data.github ?? "",
        linkedIn: result.data.linkedin ?? "",
        aboutMe: result.data.aboutMe ?? "",
        photo: avatarPreview.value || props.profile.photo || "",
        _avatarFile: avatarFile.value,
    });

    emit("close");
}

function isHost(url: string, expectedHost: string) {
    try {
        return new URL(url).hostname.includes(expectedHost);
    } catch {
        return false;
    }
}

onBeforeUnmount(() => {
    if (avatarPreview.value?.startsWith("blob:")) {
        URL.revokeObjectURL(avatarPreview.value);
    }
});
</script>
