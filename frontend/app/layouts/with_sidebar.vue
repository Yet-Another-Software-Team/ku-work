<template>
    <div class="block lg:flex">
        <SidebarComponent :is-viewer="false" />
        <slot />
    </div>
</template>
