<template>
    <div :id="'Component-' + requestId" class="flex justify-center items-center w-[20em] h-full">
        <!-- card container is clickable -->
        <div
            class="flex w-full h-[7em] m-2 shadow-md/20 bg-white dark:bg-gray-500 rounded-md cursor-pointer"
            @click="() => navigateToProfile(requestId % 3)"
        >
            <!-- profile pic -->
            <div v-if="data.profile.photo" class="w-20 h-20 flex-shrink-0">
                <img
                    :src="data.profile.photo"
                    alt="Profile photo"
                    class="w-17 h-17 object-cover rounded-full justify-center items-center m-2"
                />
            </div>
            <div v-else class="flex items-center justify-center w-20 h-20 flex-shrink-0">
                <Icon name="ic:baseline-account-circle" class="size-full" />
            </div>

            <!-- user data -->
            <div class="flex flex-col justify-center flex-1 p-2">
                <!-- name -->
                <p class="overflow-hidden truncate max-w-[12em]">{{ data.profile.name }}</p>
                <p class="text-xs">{{ data.profile.major }}</p>

                <!-- buttons -->
                <div class="flex gap-2 mt-2">
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="error"
                        label="Decline"
                        :icon="'iconoir:xmark'"
                        @click.stop="declineRequest"
                    />
                    <UButton
                        class="font-bold p-1 rounded flex items-center gap-1 w-fit px-2"
                        variant="outline"
                        color="primary"
                        label="Accept"
                        :icon="'iconoir:check'"
                        @click.stop="acceptRequest"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { mockData } from "~/data/mockData";

const props = defineProps<{
    requestId: number;
    data: typeof mockData;
}>();

const toast = useToast();
const router = useRouter();

function navigateToProfile(id: number) {
    console.log("Navigating to profile of request:", props.requestId);
    router.push(`/admin/dashboard/profile?id=${id}`);
}

// add function later
function acceptRequest() {
    console.log("Accepted request:", props.requestId);
    toast.add({
        title: "Request accepted!",
        description: props.data.profile.name,
        color: "success",
    });
}

function declineRequest() {
    console.log("Declined request:", props.requestId);
    toast.add({
        title: "Request accepted!",
        description: props.data.profile.name,
        color: "error",
    });
}
</script>
