<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-2 gap-6">
            <!-- Company Name Input -->
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-gray-800 font-semibold">Company Name *</label>
                <div class="relative">
                    <input
                        :value="companyName"
                        type="text"
                        placeholder="Enter your Company Name (username)"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.companyName }"
                        @input="updateCompanyName"
                    />
                </div>
                <span v-if="errors.companyName" class="text-red-500 text-sm">
                    {{ errors.companyName }}
                </span>
            </div>

            <!-- Company Email Input -->
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-gray-800 font-semibold">Company Email *</label>
                <div class="relative">
                    <input
                        :value="companyEmail"
                        type="email"
                        placeholder="Enter your Company Email"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.companyEmail }"
                        @input="updateCompanyEmail"
                    />
                </div>
                <span v-if="errors.companyEmail" class="text-red-500 text-sm">
                    {{ errors.companyEmail }}
                </span>
            </div>

            <!-- Password Input -->
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-gray-800 font-semibold">Password *</label>
                <div class="relative">
                    <input
                        :value="password"
                        type="password"
                        placeholder="Enter your Password"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.password }"
                        @input="updatePassword"
                    />
                </div>
                <span v-if="errors.password" class="text-red-500 text-sm">
                    {{ errors.password }}
                </span>
            </div>

            <!-- Phone Input -->
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-gray-800 font-semibold">Phone *</label>
                <div class="relative">
                    <input
                        :value="phone"
                        type="tel"
                        placeholder="Enter International Phone Number"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.phone }"
                        @input="updatePhone"
                    />
                </div>
                <span v-if="errors.phone" class="text-red-500 text-sm">
                    {{ errors.phone }}
                </span>
            </div>

            <!-- Address Input -->
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-gray-800 font-semibold">Address *</label>
                <div class="relative">
                    <input
                        :value="address"
                        type="text"
                        placeholder="Enter your Address"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.address }"
                        @input="updateAddress"
                    />
                </div>
                <span v-if="errors.address" class="text-red-500 text-sm">
                    {{ errors.address }}
                </span>
            </div>

            <!-- City Input -->
            <div class="flex flex-col space-y-1">
                <label class="text-gray-800 font-semibold">City *</label>
                <div class="relative">
                    <input
                        :value="city"
                        type="text"
                        placeholder="Enter your City"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.city }"
                        @input="updateCity"
                    />
                </div>
                <span v-if="errors.city" class="text-red-500 text-sm">
                    {{ errors.city }}
                </span>
            </div>

            <!-- Country Input -->
            <div class="flex flex-col space-y-1">
                <label class="text-gray-800 font-semibold">Country *</label>
                <div class="relative">
                    <input
                        :value="country"
                        type="text"
                        placeholder="Enter your Country"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-red-500': errors.country }"
                        @input="updateCountry"
                    />
                </div>
                <span v-if="errors.country" class="text-red-500 text-sm">
                    {{ errors.country }}
                </span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { reactive, computed } from "vue";
import * as z from "zod";

const props = defineProps({
    companyName: {
        type: String,
        default: "",
    },
    companyEmail: {
        type: String,
        default: "",
    },
    password: {
        type: String,
        default: "",
    },
    phone: {
        type: String,
        default: "",
    },
    address: {
        type: String,
        default: "",
    },
    city: {
        type: String,
        default: "",
    },
    country: {
        type: String,
        default: "",
    },
});

const emit = defineEmits([
    "update:companyName",
    "update:companyEmail",
    "update:password",
    "update:phone",
    "update:address",
    "update:city",
    "update:country",
]);

const errors = reactive({
    companyName: "",
    companyEmail: "",
    password: "",
    phone: "",
    address: "",
    city: "",
    country: "",
});

// Zod schema with improved, user-friendly error messages
const schema = z.object({
    companyName: z.string().min(2, "Name must be at least 2 characters"),
    companyEmail: z.email("Please enter a valid email address"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    phone: z.string().regex(/^\+(?:[1-9]\d{0,2})\d{4,14}$/, "Please enter a valid phone number"),
    address: z.string().min(5, "Address must be at least 5 characters"),
    city: z.string().min(2, "City must be at least 2 characters"),
    country: z.string().min(2, "Country must be at least 2 characters"),
});

const validateField = (fieldName: keyof typeof errors, value: unknown) => {
    try {
        schema.pick({ [fieldName]: true }).parse({ [fieldName]: value });
        errors[fieldName] = ""; // Clear error if validation is successful
        return true;
    } catch (error) {
        if (error instanceof z.ZodError) {
            errors[fieldName] = error.issues[0]?.message || "Invalid value";
        } else {
            errors[fieldName] = "An unexpected error occurred.";
        }
        return false;
    }
};

const isValid = computed(() => {
    const hasValues =
        props.companyName &&
        props.companyEmail &&
        props.password &&
        props.phone &&
        props.address &&
        props.city &&
        props.country;
    const hasNoErrors = !Object.values(errors).some((error) => error);
    return hasValues && hasNoErrors;
});

const updateCompanyName = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:companyName", value);
    validateField("companyName", value);
};

const updateCompanyEmail = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:companyEmail", value);
    validateField("companyEmail", value);
};

const updatePassword = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:password", value);
    validateField("password", value);
};

const updatePhone = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:phone", value);
    validateField("phone", value);
};

const updateAddress = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:address", value);
    validateField("address", value);
};

const updateCity = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:city", value);
    validateField("city", value);
};

const updateCountry = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:country", value);
    validateField("country", value);
};

defineExpose({
    isValid,
});
</script>
