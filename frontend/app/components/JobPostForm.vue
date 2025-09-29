<template>
    <div class="flex w-full p-6 mt-4 bg-white rounded-xl mx-auto max-w-6xl overflow-visible">
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
                    <USelect
                        v-model="form.type"
                        placeholder="Select Job Type"
                        class="w-full p-2 text-black bg-white border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                        :items="jobTypes"
                    />
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
                    <USelect
                        v-model="form.experience"
                        placeholder="Select Required Experience"
                        class="w-full p-2 text-black bg-white border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 hover:cursor-pointer appearance-none pr-8"
                        :items="experiences"
                    />
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
                        <UInputNumber
                            v-model="form.minSalary"
                            placeholder="Minimum Salary"
                            orientation="vertical"
                            class="w-full"
                            :min="0"
                        />
                        <UInputNumber
                            v-model="form.maxSalary"
                            placeholder="Maximum Salary"
                            orientation="vertical"
                            class="w-full"
                            :min="0"
                        />
                    </div>
                    <span class="text-error text-sm">
                        {{ errors.salary || errors.minSalary || errors.maxSalary }}
                    </span>
                </div>
            </div>

            <!-- Duration -->
            <div class="grid grid-cols-12 gap-4 items-center w-full">
                <label
                    class="col-span-12 md:col-span-4 text-left md:text-right text-primary-800 font-semibold"
                >
                    Duration
                </label>
                <div class="col-span-12 md:col-span-8">
                    <UInput
                        v-model="form.duration"
                        placeholder="e.g., 6 months, 1 year, Permanent"
                        class="w-full bg-white"
                    />
                    <span class="text-error text-sm">{{ errors.duration }}</span>
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

            <!-- Post & Cancel -->
            <div class="grid grid-cols-12 w-full">
                <div class="col-span-12 md:col-start-9 md:col-span-4 flex justify-end gap-x-3">
                    <!-- Cancel -->
                    <UButton
                        class="size-fit text-xl rounded-md px-15 font-medium hover:bg-gray-200 hover:cursor-pointer"
                        color="neutral"
                        variant="outline"
                        label="Cancel"
                        @click="cancel"
                    />

                    <!-- Post -->
                    <UButton
                        class="size-fit text-xl text-white rounded-md px-15 font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                        type="submit"
                        label="Post"
                    />
                </div>
            </div>
        </form>
    </div>
</template>

<script setup>
import { ref, reactive, watch } from "vue";
import * as z from "zod";

const emit = defineEmits(["close"]);

const { add: addToast } = useToast();

const showDiscardConfirm = ref(false);

const api = useApi();

const form = ref({
    title: "",
    location: "",
    type: "",
    experience: "",
    minSalary: "",
    maxSalary: "",
    description: "",
    duration: "",
});

const jobTypes = [
    { label: "Full-time", value: "fulltime" },
    { label: "Part-time", value: "parttime" },
    { label: "Internship", value: "internship" },
    { label: "Contract", value: "contract" },
    { label: "Casual", value: "casual" },
];

const experiences = [
    { label: "Senior", value: "senior" },
    { label: "Junior", value: "junior" },
    { label: "New Grad", value: "newgrad" },
    { label: "Manager", value: "manager" },
    { label: "Internship", value: "internship" },
];

const errors = reactive({
    title: "",
    location: "",
    type: "",
    experience: "",
    minSalary: "",
    maxSalary: "",
    description: "",
    duration: "",
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
        duration: z.string().min(1, "Duration is required"),
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

function cancel() {
    showDiscardConfirm.value = false;
    emit("close");
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
    () => form.value.duration,
    (v) => validateField("duration", v)
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

async function onSubmit() {
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
    console.log("Form data is valid:", result.data);
    const response = await api.post("/job", result.data, {
        withCredentials: true,
    });
    if (response.status < 200 || response.status >= 300) {
        addToast({
            title: "Form submission failed",
            description: response.data?.message || "An error occurred. Please try again.",
            color: "error",
        });
        return;
    }

    // Map frontend field names to backend expected names
    const backendData = {
        name: result.data.title, // title -> name
        position: result.data.title, // use title as position too
        location: result.data.location,
        jobtype: result.data.type, // type -> jobtype
        experience: result.data.experience,
        minsalary: result.data.minSalary, // minSalary -> minsalary
        maxsalary: result.data.maxSalary, // maxSalary -> maxsalary
        description: result.data.description,
        duration: result.data.duration,
        open: true, // default to open
    };

    console.log("Sending data to backend:", backendData);

    try {
        await api.post("/job", backendData, {
            withCredentials: true,
        });

        addToast({
            title: "Form submitted",
            description: "Your job post has been saved successfully.",
            color: "success",
        });
        emit("close");
    } catch (error) {
        console.error("Job submission error:", error);
        addToast({
            title: "Form submission failed",
            description: error.message || "An error occurred. Please try again.",
            color: "error",
        });
    }
}
</script>
