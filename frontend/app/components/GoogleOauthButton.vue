<template>
  <UButton
    icon="simple-icons:google"
    label="Sign In with Google"
    :loading="isLoggingIn"
    @click="oauthLogin"
  />
</template>

<script setup>
import { googleAuthCodeLogin } from "vue3-google-login";

const props = defineProps({
  isLoggingIn: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["login-success"]);
const config = useRuntimeConfig();
const toast = useToast();

const oauthLogin = async () => {
  if (props.isLoggingIn) {
    return;
  }
  
  try {
    const oauth_response = await googleAuthCodeLogin();
    
    try {
      const _ = await $fetch("/google/login", {
        method: "POST",
        baseURL: config.public.apiBaseUrl,
        body: {
          code: oauth_response.code,
        },
        credentials: 'include'
      });

      emit("login-success");

      toast.add({
        title: "Login Successful",
        description: "You have successfully logged in with Google.",
        color: "success",
      });
    } catch (apiError) {
      console.error("API login error:", apiError);
      let errorMessage = "Failed to log in with Google.";
      if (apiError.data?.detail) {
        errorMessage = apiError.data.detail;
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
  }
};
</script>
