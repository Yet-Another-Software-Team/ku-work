<template>
    <div class="mx-[12vw] main-container">
        <img src="~/assets/images/company.png" class="my-5 h-[5em] mx-auto" alt="KU Work Company" />
        <CompanyRegisterCard />
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
    background-image: url("~/assets/images/background.png");
    background-size: cover;
    background-position: center;
    z-index: -1;
}
</style>
