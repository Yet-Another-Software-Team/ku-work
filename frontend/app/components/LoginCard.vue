<template>
  <UCard class="m-4">
    <template #header>
      <h2>Login User</h2>
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
          label="Login"
          :loading="isProcessing"
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
const isProcessing = ref(false);
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

  isProcessing.value = true;

  try {
    const response = await $fetch("/login", {
      method: "POST",
      baseURL: config.public.apiBaseUrl,
      headers: {
        "Content-Type": "application/json",
      },
      body: {
        username: user.value,
        password: password.value,
      },
      credentials: "include"
    });

    localStorage.setItem("jwt_token", response.token);
    emit("login-success")
    
    toast.add({
      title: "Success",
      description: "User logged in successfully!",
      color: "success",
    });

    user.value = "";
    password.value = "";
    show.value = false;
  } catch {
    toast.add({
      title: "Error",
      description: "Invalid User or Password",
      color: "error",
    });
  } finally {
    isProcessing.value = false;
  }
};
</script>

<style scoped>
/* Scoped styles ensure this only applies to this component */
::-ms-reveal {
  display: none;
}
</style>