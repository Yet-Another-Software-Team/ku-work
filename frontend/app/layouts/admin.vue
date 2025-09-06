<template>
    <div class="block lg:flex">
        <SidebarComponent :is-admin="true" />
        <div class="m-10">
            <slot />
        </div>
    </div>
</template>
