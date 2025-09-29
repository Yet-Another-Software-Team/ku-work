<template>
    <div class="main-container px-[12vw] h-fit">
        <div
            class="w-full flex flex-col md:flex-row justify-center items-center gap-[5em] pb-5 pt-[3em]"
        >
            <div class="text-white">
                <h1 class="text-5xl font-bold">
                    <img src="~/assets/images/base.png" class="h-[2em]" alt="KU Work" />
                </h1>
                <h2 class="text-3xl font-semibold text-shadow-lg mt-3">Great job awaits for you</h2>
                <p class="text-base font-light mt-5 hidden text-shadow-lg lg:block">
                    A platform connecting KU's SKE/CPE students with<br />companies, offering direct
                    access to relevant tech-<br />jobs and internships.
                </p>
            </div>
            <LoginCard class="mt-[5em] shadow-lg" />
        </div>
    </div>
</template>

<script setup></script>

<style>
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
