<template>
    <UButton
        class="size-fit text-xl rounded-md px-10 gap-2 font-medium py-3 bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
        icon="cib:google"
        label="Continue with Google"
        :loading="isLoggingIn"
        @click="login"
    />
</template>

<script setup lang="ts">
import { googleAuthCodeLogin } from "vue3-google-login";
import { useApi, type LoginResponse } from "~/composables/useApi";

const isLoggingIn = ref(false);

const toast = useToast();

const login = async () => {
    if (isLoggingIn.value) {
        return;
    }

    isLoggingIn.value = true;

    interface oauthLoginResponse extends LoginResponse {
        isRegistered: boolean;
    }

    try {
        const oauth_response = await googleAuthCodeLogin();

        try {
            const api = useApi();
            const response = await api.post<oauthLoginResponse>(
                "/google/login",
                {
                    code: oauth_response.code,
                },
                {
                    withCredentials: true,
                }
            );

            localStorage.setItem("token", response.data.token);
            localStorage.setItem("username", response.data.username);
            localStorage.setItem("isRegistered", response.data.isRegistered.toString());

            if (response.data.isCompany) {
                localStorage.setItem("role", "company");
            } else if (response.data.isStudent) {
                localStorage.setItem("role", "student");
            } else {
                localStorage.setItem("role", "viewer");
            }

            if (response.status == 201) {
                // Account is new, navigate to registration page
                navigateTo("/register/student", { replace: true });
            } else {
                navigateTo("/jobs", { replace: true });
            }

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: any) {
            let description = "Failed to log in with Google.";

            if (error.status === 401) {
                description = "Google account not authorized. Please use a valid account.";
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
        }
    } catch (error) {
        // Handle errors from the Google login library itself.
        console.error("Google OAuth error:", error);
        toast.add({
            title: "OAuth Error",
            description: "There was an issue with the Google login process.",
            color: "error",
        });
    } finally {
        isLoggingIn.value = false;
    }
};
</script>
