<template>
    <div class="w-screen flex-col items-center justify-center mt-5 overflow-hidden">
        <img
            src="~/assets/images/admin.png"
            class="h-[4em] lg:h-[5em] w-full object-contain mb-10"
            alt="KU Work Admin"
        />
        <div class="bg-white rounded-lg shadow-md p-6 w-[95vw] lg:w-[28em] mx-auto">
            <h1 class="text-primary text-3xl font-bold mb-4 text-center">Admin Login</h1>
            <AdminLoginField />
        </div>
    </div>
</template>

<script setup lang="ts"></script>
