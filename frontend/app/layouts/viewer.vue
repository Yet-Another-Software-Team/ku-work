<template>
    <div class="block lg:flex">
        <SidebarComponent :is-viewer="false" />
        <div class="m-10 lg:w-full">
            <slot />
        </div>
    </div>
</template>
