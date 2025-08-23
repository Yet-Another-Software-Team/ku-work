<template>

  <div class="main-container px-[12vw] h-fit">
    <div class="h-[8em] pt-2">
      <Icon name="meteocons:umbrella-wind-alt" class="size-[8em] text-white" />
    </div>
    <div class="w-full flex flex-col md:flex-row justify-center items-center gap-[5em] pb-5">
      <div class="text-white">
        <h1 class="text-5xl font-bold">Here in <span class="text-primary">KU-Work</span>,</h1>
        <h2 class="text-3xl font-semibold mt-3">Great job awaits for you</h2>
        <p class="text-base font-light mt-5 hidden lg:block">A platform connecting KU's SKE/CPE students with<br>companies, offering direct access to relevant tech-<br>jobs and internships.</p>
      </div>
      <LoginCard/>
    </div>
  </div>
</template>

<script setup>
</script>

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
  background-image: url('/images/background.png');
  background-size: cover;
  background-position: center;
  z-index: -1;
}
</style>
