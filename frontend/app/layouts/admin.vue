<template>
    <div class="block lg:flex light:bg-[#F7F8F4]">
        <SidebarComponent :is-admin="true" />
        <div class="mx-10 lg:w-full h-full min-h-dvh">
            <slot />
        </div>
    </div>
</template>
