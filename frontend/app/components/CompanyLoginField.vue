<template>
    <UForm :schema="schema" :state="state" class="flex flex-col items-center" @submit="onSubmit">
        <UFormField name="username" class="mx-auto mb-4 w-[80%]">
            <UInput
                v-model="state.username"
                type="text"
                placeholder="Username"
                class="w-full"
                :ui="{
                    base: 'bg-white text-black text-xl',
                }"
            />
        </UFormField>
        <UFormField name="password" class="mx-auto mb-4 w-[80%]">
            <UInput
                v-model="state.password"
                placeholder="Password"
                class="w-full"
                :type="show ? 'text' : 'password'"
                :ui="{
                    base: 'bg-white text-black text-xl',
                    trailing: 'pe-1',
                }"
            >
                <template #trailing>
                    <UButton
                        color="neutral"
                        variant="link"
                        size="sm"
                        :icon="show ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                        :aria-label="show ? 'Hide password' : 'Show password'"
                        :aria-pressed="show"
                        aria-controls="password"
                        @click="show = !show"
                    />
                </template>
            </UInput>
        </UFormField>

        <UButton
            class="size-fit text-xl text-white rounded-md px-15 font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
            type="submit"
            label="Log In"
        />
    </UForm>
</template>

<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";

const toast = useToast();
const config = useRuntimeConfig();

const schema = z.object({
    username: z.string().min(1, "Username is required."),
    password: z.string().min(1, "Password is required."),
});

type Schema = z.output<typeof schema>;

type loginResponse = {
    token: string;
};

const show = ref(false);
const isLoggingIn = ref(false);
const state = reactive<Partial<Schema>>({
    username: undefined,
    password: undefined,
});

async function onSubmit(_: FormSubmitEvent<Schema>) {
    if (!state.username || !state.password) {
        toast.add({
            title: "Validation Error",
            description: "Please provide both username and password",
            color: "error",
        });
        return;
    }

    isLoggingIn.value = true;

    try {
        const response: loginResponse = await $fetch("/login", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            headers: {
                "Content-Type": "application/json",
            },
            body: {
                username: state.username,
                password: state.password,
            },
            credentials: "include",
        });

        localStorage.setItem("jwt_token", response.token);

        toast.add({
            title: "Success",
            description: "User logged in successfully!",
            color: "success",
        });

        state.username = "";
        state.password = "";
        show.value = false;
    } catch {
        toast.add({
            title: "Error",
            description: "Invalid User or Password",
            color: "error",
        });
    } finally {
        isLoggingIn.value = false;
    }
}
</script>

<style>
::-ms-reveal {
    display: none;
}
</style>
