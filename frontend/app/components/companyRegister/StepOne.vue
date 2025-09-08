<template>
    <div class="space-y-6 my-5">
        <div class="grid grid-cols-2 gap-6">
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-primary-800 font-semibold">Company Name *</label>
                <div class="relative">
                    <input
                        :value="companyName"
                        type="text"
                        placeholder="Enter your Company Name"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.companyName }"
                        @input="updateCompanyName"
                    />
                </div>
                <span v-if="errors.companyName" class="text-error text-sm">
                    {{ errors.companyName }}
                </span>
            </div>
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-primary-800 font-semibold">Company Email *</label>
                <div class="relative">
                    <input
                        :value="companyEmail"
                        type="email"
                        placeholder="Enter your Company Email"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.companyEmail }"
                        @input="updateCompanyEmail"
                    />
                </div>
                <span v-if="errors.companyEmail" class="text-error text-sm">
                    {{ errors.companyEmail }}
                </span>
            </div>
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-primary-800 font-semibold">Password *</label>
                <div class="relative">
                    <input
                        :value="password"
                        type="password"
                        placeholder="Enter your Password"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.password }"
                        @input="updatePassword"
                    />
                </div>
                <span v-if="errors.password" class="text-error text-sm">
                    {{ errors.password }}
                </span>
            </div>
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-primary-800 font-semibold">Phone *</label>
                <div class="relative">
                    <input
                        :value="phone"
                        type="text"
                        placeholder="Enter your Phone Number"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.phone }"
                        @input="updatePhone"
                    />
                </div>
                <span v-if="errors.phone" class="text-error text-sm">
                    {{ errors.phone }}
                </span>
            </div>
            <div class="flex flex-col space-y-1 col-span-2">
                <label class="text-primary-800 font-semibold">Address *</label>
                <div class="relative">
                    <input
                        :value="address"
                        type="text"
                        placeholder="Enter your Address"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.address }"
                        @input="updateAddress"
                    />
                </div>
                <span v-if="errors.address" class="text-error text-sm">
                    {{ errors.address }}
                </span>
            </div>
            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">City *</label>
                <div class="relative">
                    <input
                        :value="city"
                        type="text"
                        placeholder="Enter your City"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.city }"
                        @input="updateCity"
                    />
                </div>
                <span v-if="errors.city" class="text-error text-sm">
                    {{ errors.city }}
                </span>
            </div>
            <div class="flex flex-col space-y-1">
                <label class="text-primary-800 font-semibold">Country *</label>
                <div class="relative">
                    <input
                        :value="country"
                        type="text"
                        placeholder="Enter your Country"
                        class="w-full px-4 py-3 text-black bg-white border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                        :class="{ 'border-error': errors.country }"
                        @input="updateCountry"
                    />
                </div>
                <span v-if="errors.country" class="text-error text-sm">
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

const schema = z.object({
    companyName: z.string().min(2).max(100),
    companyEmail: z.string().email(),
    password: z.string().min(8).max(100),
    phone: z.string().min(10).max(20),
    address: z.string().min(5).max(200),
    city: z.string().min(2).max(100),
    country: z.string().min(2).max(100),
});

const validateField = (fieldName: keyof typeof errors, value: unknown) => {
    try {
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
    if (value) validateField("companyName", value);
};

const updateCompanyEmail = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:companyEmail", value);
    if (value) validateField("companyEmail", value);
};

const updatePassword = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:password", value);
    if (value) validateField("password", value);
};

const updatePhone = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:phone", value);
    if (value) validateField("phone", value);
};

const updateAddress = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:address", value);
    if (value) validateField("address", value);
};

const updateCity = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:city", value);
    if (value) validateField("city", value);
};

const updateCountry = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    emit("update:country", value);
    if (value) validateField("country", value);
};

defineExpose({
    isValid,
});
</script>
