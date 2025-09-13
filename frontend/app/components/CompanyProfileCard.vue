<template>
    <div class="flex">
        <!-- Main Content -->
        <main class="flex-1 p-8 overflow-y-auto">
            <h2 class="text-3xl font-bold text-green-700 mb-6">Profile</h2>

            <!-- Banner -->
            <div class="bg-gray-300 h-32 rounded-lg relative overflow-hidden -z-1">
                <!-- Banner image placeholder -->
            </div>

            <!-- Company Info -->
            <div class="flex items-center gap-4 -mt-10 mb-6">
                <div class="w-24 h-24 rounded-full bg-gray-400 border-4 border-white"></div>
                <div>
                    <h3 class="text-xl font-bold">Roblox Corp.</h3>
                    <p class="text-gray-600">San Mateo, CA, United States</p>
                </div>
                <div class="ml-auto">
                    <UButton label="Edit Profile" variant="outline" />
                </div>
            </div>

            <hr class="my-6" />

            <!-- Contact + About -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                <!-- Contact -->
                <div>
                    <h4 class="font-bold mb-2">Contact</h4>
                    <p class="flex items-center gap-2 text-gray-700">
                        <Icon name="ic:baseline-link" />
                        <a href="https://www.roblox.com" class="text-blue-600 hover:underline"
                            >https://www.roblox.com</a
                        >
                    </p>
                    <p class="flex items-center gap-2 text-gray-700 mt-2">
                        <Icon name="ic:baseline-email" /> roblox.kid@co.com
                    </p>
                </div>

                <!-- About -->
                <div>
                    <h4 class="font-bold mb-2">About us</h4>
                    <p class="text-gray-600 text-sm leading-relaxed">
                        Lorem ipsum dolor sit amet consectetur adipiscing elit. Amet consectetur
                        adipiscing elit quisque faucibus ex sapien. Quisque faucibus ex sapien vitae
                        pellentesque sem placerat. Vitae pellentesque sem placerat in id cursus mi.
                    </p>
                </div>
            </div>
        </main>
    </div>
</template>

<script setup lang="ts"></script>
