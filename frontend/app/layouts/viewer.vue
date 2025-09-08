<template>
    <div class="block lg:flex light:bg-[#F7F8F4]">
        <SidebarComponent :is-viewer="false" />
        <div class="m-10 lg:w-full h-full min-h-dvh">
            <slot />
        </div>
    </div>
</template>
