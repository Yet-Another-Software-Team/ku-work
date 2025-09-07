<template>
    <div class="block lg:flex light:bg-[#F7F8F4]">
        <SidebarComponent :is-admin="true" />
        <div class="m-10 lg:w-full h-full">
            <slot />
        </div>
    </div>
</template>
