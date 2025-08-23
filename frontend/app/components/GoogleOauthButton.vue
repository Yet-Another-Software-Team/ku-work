<template>
  <UButton
    class="size-fit text-xl rounded-md px-10 gap-2 font-medium py-3 bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
    icon="cib:google"
    label="Sign In with Google"
    :loading="isLoggingIn"
    @click="login"
  />
</template>

<script setup lang="ts">
import { googleAuthCodeLogin } from "vue3-google-login";

const isLoggingIn = ref(false);

const config = useRuntimeConfig();
const toast = useToast();

type loginResponse = {
  token: string
}

const login = async () => {
  if (isLoggingIn.value) {
    return;
  }
  
  isLoggingIn.value = true

  try {
    const oauth_response = await googleAuthCodeLogin();
    
    try {
      const response: loginResponse = await $fetch("/google/login", {
        method: "POST",
        baseURL: config.public.apiBaseUrl,
        body: {
          code: oauth_response.code,
        },
        credentials: 'include'
      });

      localStorage.setItem("jwt_token", response.token);

      toast.add({
        title: "Login Successful",
        description: "You have successfully logged in with Google.",
        color: "success",
      });
    } catch (apiError: any) {

      let errorMessage = "Failed to log in with Google.";
      if (apiError.message) {
        errorMessage = apiError.message;
      }
      
      // Show an error message.
      toast.add({
        title: "Login Failed",
        description: errorMessage,
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
