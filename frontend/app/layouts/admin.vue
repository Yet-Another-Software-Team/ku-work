<template>
    <div class="block lg:flex bg-[#F7F8F4] dark:bg-neutral-900">
        <SidebarComponent :is-admin="true" />
        <div class="lg:w-full h-full min-h-dvh flex flex-col">
            <div class="flex-grow mx-10">
                <slot />
            </div>
            <LayoutFooter />
        </div>
    </div>
</template>
