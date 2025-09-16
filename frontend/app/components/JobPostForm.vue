<template>
    <a href="/dashboard">
        <h1
            class="flex items-center text-5xl text-primary-800 dark:text-primary font-bold mb-6 gap-2 cursor-pointer"
        >
            <Icon name="iconoir:nav-arrow-left" class="items-center" />
            <span>Back</span>
        </h1>
    </a>

    <div class="flex w-full p-6 bg-white rounded-xl shadow-md mx-auto max-w-4xl overflow-visible">
        <form class="space-y-4 w-full flex-1" @submit.prevent="onSubmit">
            <!-- Job Title -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Job Title
                </label>
                <div class="col-span-12 md:col-span-8">
                    <UInput
                        v-model="form.title"
                        placeholder="Enter job title"
                        class="w-full bg-white"
                    />
                    <span class="text-error text-sm">{{ errors.title }}</span>
                </div>
            </div>

            <!-- Job Location -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Job Location
                </label>
                <div class="col-span-12 md:col-span-8 bg-white">
                    <UInput
                        v-model="form.location"
                        placeholder="Location"
                        icon="material-symbols:location-on-outline-rounded"
                        class="w-full bg-white"
                    />
                    <span class="text-error text-sm">{{ errors.location }}</span>
                </div>
            </div>

            <!-- Job Type -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Job Type
                </label>
                <div class="col-span-12 md:col-span-8 relative z-50">
                    <select
                        v-model="form.type"
                        class="w-full px-2 py-1 text-black bg-white border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                    >
                        <option value="" disabled>Select Job Type</option>
                        <option v-for="opt in jobTypes" :key="opt.value" :value="opt.value">
                            {{ opt.label }}
                        </option>
                    </select>
                    <span class="text-error text-sm">{{ errors.type }}</span>
                </div>
            </div>

            <!-- Required Experience -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Required Experience
                </label>
                <div class="col-span-12 md:col-span-8 relative z-50">
                    <select
                        v-model="form.experience"
                        class="w-full px-2 py-1 text-black bg-white border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                    >
                        <option value="" disabled>Select Required Experience</option>
                        <option v-for="opt in experiences" :key="opt.value" :value="opt.value">
                            {{ opt.label }}
                        </option>
                    </select>
                    <span class="text-error text-sm">{{ errors.experience }}</span>
                </div>
            </div>

            <!-- Salary -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Salary
                </label>
                <div class="col-span-12 md:col-span-8">
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <UInput
                            v-model="form.minSalary"
                            type="number"
                            placeholder="Minimum Salary"
                            class="w-full"
                        >
                            <template #trailing>Baht</template>
                        </UInput>
                        <UInput
                            v-model="form.maxSalary"
                            type="number"
                            placeholder="Maximum Salary"
                            class="w-full"
                        >
                            <template #trailing>Baht</template>
                        </UInput>
                    </div>
                    <span class="text-error text-sm">
                        {{ errors.salary || errors.minSalary || errors.maxSalary }}
                    </span>
                </div>
            </div>

            <!-- Job Description -->
            <div class="grid grid-cols-12 gap-4 items-start w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Job Description
                </label>
                <div class="col-span-12 md:col-span-8">
                    <UTextarea
                        v-model="form.description"
                        rows="6"
                        placeholder="Enter job description"
                        class="w-full"
                    />
                    <span class="text-error text-sm">{{ errors.description }}</span>
                </div>
            </div>

            <!-- Actions -->
            <div class="grid grid-cols-12 w-full">
                <div class="col-span-12 md:col-start-9 md:col-span-4 flex justify-end">
                    <UButton
                        class="size-fit text-xl text-white rounded-md px-15 font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                        type="submit"
                        label="Submit"
                    />
                </div>
            </div>
        </form>
    </div>
</template>

<script setup>
import { ref, reactive, watch } from "vue";
import * as z from "zod";

const { add: addToast } = useToast();

const form = ref({
    title: "",
    location: "",
    type: "",
    experience: "",
    minSalary: "",
    maxSalary: "",
    description: "",
});

const jobTypes = [
    { label: "Full-time", value: "Full-time" },
    { label: "Part-time", value: "Part-time" },
    { label: "Internship", value: "Internship" },
    { label: "Contract", value: "Contract" },
];

const experiences = [
    { label: "Senior", value: "Senior" },
    { label: "Junior", value: "Junior" },
    { label: "New Grad", value: "New Grad" },
    { label: "Manager", value: "Manager" },
];

const errors = reactive({
    title: "",
    location: "",
    type: "",
    experience: "",
    minSalary: "",
    maxSalary: "",
    description: "",
    salary: "",
});

const schema = z
    .object({
        title: z.string().min(1, "Job Title is required"),
        location: z.string().min(1, "Job Location is required"),
        type: z.string().min(1, "Job Type is required"),
        experience: z.string().min(1, "Experience is required"),
        minSalary: z.coerce.number().min(0, "Minimum salary cannot be negative"),
        maxSalary: z.coerce.number().min(0, "Maximum salary cannot be negative"),
        description: z.string().min(1, "Job Description is required"),
    })
    .refine((d) => d.minSalary <= d.maxSalary, {
        message: "Minimum salary must be less than or equal to maximum salary",
        path: ["salary"],
    });

function validateField(fieldName, value) {
    try {
        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });

        if (typeof value === "string" && value.trim() === "") {
            errors[fieldName] = "This field is required";
            return false;
        }

        errors[fieldName] = "";
        return true;
    } catch (error) {
        if (error instanceof z.ZodError) {
            errors[fieldName] = error.issues[0]?.message ?? "Invalid value";
        } else {
            errors[fieldName] = "Invalid value";
        }
        return false;
    }
}

function validateSalaryCross() {
    errors.salary = "";
    const min = Number(form.value.minSalary);
    const max = Number(form.value.maxSalary);

    if (!Number.isFinite(min) || !Number.isFinite(max)) return;

    if (min > max) {
        errors.salary = "Minimum salary must be less than or equal to maximum salary";
    }
}

watch(
    () => form.value.title,
    (v) => validateField("title", v)
);
watch(
    () => form.value.location,
    (v) => validateField("location", v)
);
watch(
    () => form.value.type,
    (v) => validateField("type", v)
);
watch(
    () => form.value.experience,
    (v) => validateField("experience", v)
);
watch(
    () => form.value.description,
    (v) => validateField("description", v)
);

watch(
    () => form.value.minSalary,
    (v) => {
        validateField("minSalary", v);
        validateSalaryCross();
    }
);
watch(
    () => form.value.maxSalary,
    (v) => {
        validateField("maxSalary", v);
        validateSalaryCross();
    }
);

function onSubmit() {
    const result = schema.safeParse(form.value);

    if (!result.success) {
        for (const issue of result.error.issues) {
            const key = issue.path?.[0];
            if (typeof key === "string" && key in errors) {
                errors[key] = issue.message;
            } else if (key === "salary" || key === undefined) {
                errors.salary = issue.message;
            }
        }
        addToast({
            title: "Form submission failed",
            description: "Please check the highlighted errors and try again.",
            color: "warning",
        });
        return;
    }

    addToast({
        title: "Form submitted",
        description: "Your job post has been saved successfully.",
        color: "success",
    });
}
</script>
