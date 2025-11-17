<template>
    <div class="block lg:flex bg-[#F7F8F4] dark:bg-neutral-900">
        <SidebarComponent :is-admin="true" />
        <div class="mx-10 lg:w-full h-full min-h-dvh flex flex-col">
            <div class="flex-grow">
                <slot />
            </div>
            <LayoutFooter />
        </div>
    </div>
</template>
