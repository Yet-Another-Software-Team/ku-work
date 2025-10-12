<template>
    <div class="main-container w-screen flex-col items-center justify-center pt-5 overflow-hidden">
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

<style scoped>
.main-container {
    position: relative;
}

.main-container::before {
    content: "";
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-image: url("~/assets/images/background.png");
    background-size: cover;
    background-position: center;
    z-index: -1;
}
</style>