<template>
    <div>
        <h1 class="text-8xl w-full text-center font-bold">API Test</h1>
        <p class="text-2xl w-full text-center mx-auto">
            This is page for testing API endpoints and it is a subject to be
            changed
        </p>

        <!-- Add User Card -->
        <UCard class="m-4">
            <template #header>
                <h2>Add User</h2>
            </template>
            <div class="flex flex-col gap-2">
                <UInput
                    v-model="user"
                    class="w-[50%] h-[2em]"
                    placeholder="Username"
                />
                <UInput
                    v-model="password"
                    class="w-[50%] h-[2em]"
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
                        label="Add User"
                        :loading="isAdding"
                        :disabled="!user || !password"
                        @click="addUser"
                    />
                    <UButton
                        label="Refresh Users"
                        variant="outline"
                        :loading="isLoading"
                        @click="fetchUsers"
                    />
                </div>
            </template>
        </UCard>

        <!-- Users List Card -->
        <UCard class="m-4">
            <template #header>
                <h2>Users List</h2>
            </template>

            <div v-if="isLoading" class="flex justify-center p-4">
                <div
                    class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"
                />
            </div>

            <div
                v-else-if="users.length === 0"
                class="text-center p-4 text-gray-500"
            >
                No users found
            </div>

            <div v-else class="space-y-2">
                <div
                    v-for="(userData, index) in users"
                    :key="index"
                    class="p-3 border border-gray-200 rounded-lg flex justify-between items-center"
                >
                    <div>
                        <span class="text-gray-600 pr-4">ID: {{ userData.ID || userData.id }}</span>
                        <span class="font-medium">{{
                            userData.Username ||
                            userData.username ||
                            `User ${index + 1}`
                        }}</span>
                    </div>
                </div>
            </div>
        </UCard>
    </div>
</template>

<script setup>
const toast = useToast();
const config = useRuntimeConfig();

const user = ref("");
const password = ref("");
const show = ref(false);
const users = ref([]);
const isLoading = ref(false);
const isAdding = ref(false);

// Fetch users on component mount
onMounted(() => {
    fetchUsers();
});

// Function to add a new user
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
        const _ = await $fetch("/create_user", {
            method: "POST",
            baseURL: config.public.apiBaseUrl,
            headers: {
                "Content-Type": "application/json",
            },
            body: {
                user: user.value,
                password: password.value,
            },
        });

        toast.add({
            title: "Success",
            description: "User created successfully!",
            color: "success",
        });

        // Clear form
        user.value = "";
        password.value = "";
        show.value = false;

        // Refresh users list
        await fetchUsers();
    } catch (error) {
        console.error("Error creating user:", error);

        let errorMessage = "Failed to create user";
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

// Function to fetch all users
const fetchUsers = async () => {
    isLoading.value = true;

    try {
        const response = await $fetch("/users", {
            method: "GET",
            baseURL: config.public.apiBaseUrl,
        });

        // Handle backend response format: {"users": [...]}
        if (response.users && Array.isArray(response.users)) {
            users.value = response.users;
        } else if (Array.isArray(response)) {
            users.value = response;
        } else if (response.data && Array.isArray(response.data)) {
            users.value = response.data;
        } else {
            users.value = [];
        }
    } catch (error) {
        console.error("Error fetching users:", error);

        let errorMessage = "Failed to fetch users";
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

        users.value = [];
    } finally {
        isLoading.value = false;
    }
};
</script>

<style>
/* Hide the password reveal button in Edge */
::-ms-reveal {
    display: none;
}
</style>
