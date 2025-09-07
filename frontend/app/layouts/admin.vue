<template>
    <div class="block lg:flex">
        <SidebarComponent :is-admin="true" />
        <div class="m-10 lg:w-full">
            <slot />
        </div>
    </div>
</template>
