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
                        class="cursor-pointer"
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
        <TurnstileWidget @callback="(tk) => (token = tk)" />
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
import { useApi, type LoginResponse } from "~/composables/useApi";

const toast = useToast();
const api = useApi();

const schema = z.object({
    username: z.string().min(1, "Username is required."),
    password: z.string().min(1, "Password is required."),
});

type Schema = z.output<typeof schema>;

const show = ref(false);
const isLoggingIn = ref(false);
const state = reactive<Partial<Schema>>({
    username: "",
    password: "",
});
const token = ref<string>("");

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
        const response = await api.post<LoginResponse>(
            "/auth/admin/login",
            {
                username: state.username,
                password: state.password,
            },
            {
                headers: {
                    "X-Turnstile-Token": token.value,
                },
                withCredentials: true,
            }
        );

        localStorage.setItem("token", response.data.token);
        localStorage.setItem("username", response.data.username);
        localStorage.setItem("userId", response.data.userId);
        localStorage.setItem("role", response.data.role);

        navigateTo("/admin/dashboard", { replace: true });

        state.username = "";
        state.password = "";
        show.value = false;

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
        let description = "Incorrect username or password. Please try again.";

        if (error.status === 401) {
            description = "Incorrect username or password. Please try again.";
        } else if (error.status === 500) {
            description = "Server error. Please try again later.";
        } else if (error.message) {
            description = error.message;
        }

        toast.add({
            title: "Login Failed",
            description,
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
