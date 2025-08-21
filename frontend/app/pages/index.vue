<template>
    <div>
        <h1 class="text-8xl w-full text-center font-bold">API Test</h1>
        <p class="text-2xl w-full text-center mx-auto">
            This is a page for testing API endpoints and is subject to change.
        </p>

        <!-- User Registration Card -->
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
                            :aria-label="
                                show ? 'Hide password' : 'Show password'
                            "
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

        <!-- Login Card -->
        <UCard class="m-4">
            <template #header>
                <h2>Login</h2>
            </template>
            <div class="flex flex-col gap-2">
                <UInput
                    v-model="loginUser.username"
                    class="w-full h-[2em]"
                    placeholder="Username"
                />
                <UInput
                    v-model="loginUser.password"
                    class="w-full h-[2em]"
                    placeholder="Password"
                    type="password"
                />
            </div>
            <template #footer>
                <div class="flex flex-col gap-2">
                    <UButton
                        label="Login"
                        :loading="isLoggingIn"
                        :disabled="!loginUser.username || !loginUser.password"
                        @click="login"
                    />
                    <div
                        v-if="localToken"
                        class="text-wrap break-all text-sm mt-2"
                    >
                        <h4 class="font-bold">JWT Token in Local Storage:</h4>
                        <p class="text-gray-500">{{ localToken }}</p>
                    </div>
                </div>
            </template>
        </UCard>
        
        <!-- Logout Card -->
        <UCard class="m-4">
            <template #header>
                <h2>Logout</h2>
            </template>
            <p>
                Click this button to log out. This will remove your JWT token and clear the protected route response.
            </p>
            <template #footer>
                <UButton
                    label="Logout"
                    color="primary"
                    @click="logout"
                />
            </template>
        </UCard>

        <!-- Protected Route Card -->
        <UCard class="m-4">
            <template #header>
                <h2>Test Protected Route</h2>
            </template>
            <div class="flex flex-col gap-2">
                <p>
                    Click the button below to test the protected `/protected`
                    route. You must be logged in and have a JWT token stored.
                </p>
            </div>
            <template #footer>
                <div class="flex flex-col gap-2">
                    <UButton
                        label="Test API"
                        :loading="isTestingProtected"
                        @click="testProtected"
                    />
                    <div v-if="profileData" class="text-sm mt-2">
                        <h4 class="font-bold">API Response:</h4>
                        <pre
                            class="bg-gray-100 p-2 rounded-lg text-gray-800 break-all"
                            >{{ JSON.stringify(profileData, null, 2) }}</pre>
                        >
                    </div>
                </div>
            </template>
        </UCard>
    </div>
</template>

<script setup>
import { onMounted } from 'vue';

const toast = useToast();
const config = useRuntimeConfig();

// Registration
const user = ref("");
const password = ref("");
const show = ref(false);
const isAdding = ref(false);

// Login
const loginUser = ref({
    username: "",
    password: "",
});
const localToken = ref(null);
const isLoggingIn = ref(false);

// Protected Route
const profileData = ref(null);
const isTestingProtected = ref(false);

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
        });

        localStorage.setItem("jwt_token", response.token);
        localToken.value = response.token;
        
        toast.add({
            title: "Success",
            description: "User registered and logged in successfully!",
            color: "success",
        });

        // Clear form
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

// Function to log in and get tokens
const login = async () => {
    if (!loginUser.value.username || !loginUser.value.password) {
        toast.add({
            title: "Validation Error",
            description: "Please provide both username and password for login",
            color: "error",
        });
        return;
    }

    isLoggingIn.value = true;
    localToken.value = null;
    profileData.value = null;

    try {
        const response = await $fetch("/login", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            headers: {
                "Content-Type": "application/json",
            },
            body: {
                username: loginUser.value.username,
                password: loginUser.value.password,
            },
            credentials: 'include'
        });

        localStorage.setItem("jwt_token", response.token);
        localToken.value = response.token;

        toast.add({
            title: "Login Successful",
            description:
                "You have successfully logged in and received a token.",
            color: "success",
        });
    } catch (error) {
        console.error("Error during login:", error);

        let errorMessage = "Failed to log in";
        if (error.data?.message) {
            errorMessage = error.data.message;
        } else if (error.data?.error) {
            errorMessage = error.data.error;
        } else if (error.message) {
            errorMessage = error.message;
        }

        toast.add({
            title: "Login Failed",
            description: errorMessage,
            color: "error",
        });
    } finally {
        isLoggingIn.value = false;
    }
};

// Function to handle logout
const logout = async () => {
    try {
        await $fetch("/logout", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            credentials: 'include'
        });
    } catch (error) {
        console.error("Error during logout:", error);
    } finally {
        localStorage.removeItem("jwt_token");
        localToken.value = null;
        profileData.value = null;

        toast.add({
            title: "Logged Out",
            description: "You have been successfully logged out.",
            color: "warning",
        });
    }
};

// Helper function to refresh the JWT token
const refreshToken = async () => {
    try {
        const response = await $fetch("/refresh-token", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            credentials: 'include'
        });

        localStorage.setItem("jwt_token", response.token);
        localToken.value = response.token;
        return response.token;
    } catch (error) {
        console.error("Failed to refresh token:", error);
        throw error;
    }
};

// Function to fetch the user profile from a protected route
const testProtected = async () => {
    const token = localStorage.getItem("jwt_token");

    isTestingProtected.value = true;
    profileData.value = null;

    try {
        const response = await $fetch("/protected", {
            method: "GET",
            baseURL: config.public.apiBaseUrl,
            headers: {
                Authorization: `Bearer ${token}`,
            },
            credentials: 'include'
        });

        profileData.value = response;
        toast.add({
            title: "Success",
            description: "Successfully fetched protected profile data.",
            color: "success",
        });
    } catch (error) {
        console.error("Error accessing protected route:", error);

        if (error.response?.status === 401) {
            console.log("Attempting to refresh token...");
            try {
                const newToken = await refreshToken();

                const response = await $fetch("/protected", {
                    method: "GET",
                    baseURL: config.public.apiBaseUrl,
                    headers: {
                        Authorization: `Bearer ${newToken}`,
                    },
                    credentials: 'include'
                });

                profileData.value = response;
                toast.add({
                    title: "Success (Refreshed)",
                    description: "Token refreshed and profile data fetched.",
                    color: "success",
                });

            } catch {
                toast.add({
                    title: "Session Expired",
                    description: "Your session has expired. Please log in again.",
                    color: "error",
                });
            }
        } else {
            // Handle other types of errors
            let errorMessage = "Failed to access protected route";
            if (error.data?.error) {
                errorMessage = error.data.error;
            } else if (error.data?.message) {
                errorMessage = error.data.message;
            } else if (error.message) {
                errorMessage = error.message;
            }

            toast.add({
                title: "Authorization Error",
                description: errorMessage,
                color: "error",
            });
        }
    } finally {
        isTestingProtected.value = false;
    }
};

// This lifecycle hook runs when the component is mounted.
onMounted(() => {
  const storedToken = localStorage.getItem("jwt_token");
  if (storedToken) {
    localToken.value = storedToken;
  }
});
</script>

<style>
/* Hide the password reveal button in Edge */
::-ms-reveal {
    display: none;
}
</style>
