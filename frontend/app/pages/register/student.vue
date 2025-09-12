<template>
    <div class="mx-[12vw] main-container mt-[5vw]">
        <StudentRegisterCard />
    </div>
</template>

<style>
.main-container::before {
    content: "";
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-image: url("/images/background.png");
    background-size: cover;
    background-position: center;
    z-index: -1;
}
</style>
