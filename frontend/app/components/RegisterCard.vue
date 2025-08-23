<template>
  <UCard class="m-4">
    <template #header>
      <h2>Register User</h2>
    </template>
    <div class="flex flex-col gap-2">
      <UInput
        v-model="user"
        class="w-full h-[2em]"
        placeholder="Username"
      />
      <UInput
        v-model="password"
        class="w-full h-[2em]"
        placeholder="Password"
        :type="show ? 'text' : 'password'"
        :ui="{ trailing: 'pe-1' }"
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
    </div>
    <template #footer>
      <div class="flex gap-2">
        <UButton
          label="Register User"
          :loading="isAdding"
          :disabled="!user || !password"
          @click="addUser"
        />
      </div>
    </template>
  </UCard>
</template>

<script setup>
const toast = useToast();
const config = useRuntimeConfig();

const user = ref("");
const password = ref("");
const show = ref(false);
const isAdding = ref(false);
const emit = defineEmits(["login-success"]);

const addUser = async () => {
  if (!user.value || !password.value) {
    toast.add({
      title: "Validation Error",
      description: "Please provide both username and password",
      color: "error",
    });
    return;
  }

  isAdding.value = true;

  try {
    const response = await $fetch("/register", {
      method: "POST",
      baseURL: config.public.apiBaseUrl,
      headers: {
        "Content-Type": "application/json",
      },
      body: {
        username: user.value,
        password: password.value,
      },
      credentials: 'include'
    });

    localStorage.setItem("jwt_token", response.token);
    // Emit event to notify parent of token change
    emit("login-success");
    
    toast.add({
      title: "Success",
      description: "User registered and logged in successfully!",
      color: "success",
    });

    user.value = "";
    password.value = "";
    show.value = false;
  } catch (error) {
    console.error("Error creating user:", error);

    let errorMessage = "Failed to register user";
    if (error.data?.detail) {
      errorMessage = error.data.detail;
    } else if (error.data?.message) {
      errorMessage = error.data.message;
    } else if (error.message) {
      errorMessage = error.message;
    }

    toast.add({
      title: "Error",
      description: errorMessage,
      color: "error",
    });
  } finally {
    isAdding.value = false;
  }
};
</script>

<style scoped>
/* Scoped styles ensure this only applies to this component */
::-ms-reveal {
  display: none;
}
</style>