<template>
<div>
    <h1 class="text-3xl font-bold w-full text-center">Welcome to KU-Work</h1>
    <p class="text-lg w-full text-center">This is the homepage of KU-Work.</p>
    
    <div v-if="!login">
        <LoginCard @login-success="handleLoginSuccess" />
        <RegisterCard @login-success="handleLoginSuccess" />
        <GoogleOauthButton @login-success="handleLoginSuccess" />
    </div>
    <div v-else>
        <h2>Welcome back!</h2>
        <p>You are logged in.</p>
        <p v-if="isAdmin">You are an admin.</p>
        <LogoutButton @logout="handleLogoutSuccess" />
    </div>
</div>
</template>

<script setup>
const login = ref(false);
const isAdmin = ref(false);
const isRefreshing = ref(false);
const config = useRuntimeConfig();

onMounted(() => {
  // Try to refresh the token
  refreshToken();
});

const refreshToken = async () => {
  if (isRefreshing.value) return;
  isRefreshing.value = true;
  try {
    const response = await $fetch("/refresh", {
        method: "POST",
        baseURL: config.public.apiBaseUrl,
        credentials: 'include'
    });
    
    localStorage.setItem("jwt_token", response.token);
    login.value = true;
    
    const admin_resp = await $fetch("/admin", {
        baseURL: config.public.apiBaseUrl,
        headers: {
            Authorization: `Bearer ${localStorage.getItem("jwt_token")}`,
        },
        credentials: 'include'
    });
    isAdmin.value = true;
  } catch (error) {
    isAdmin.value = false;
    console.error(error);
  } finally {
    isRefreshing.value = false;
  }
};

function handleLoginSuccess() {
    login.value = true;
}

function handleLogoutSuccess() {
    login.value = false;
}

</script>
