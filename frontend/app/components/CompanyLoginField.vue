<template>
  <div class="flex flex-col items-center">
    <UInput
      v-model="username"
      type="text"
      placeholder="Username"
      class="mx-auto mb-4 w-[80%]"
      :ui="{
        base: 'bg-white text-black',
      }"
    />
    <UInput
      v-model="password"
      placeholder="Password"
      class="mx-auto mb-4 w-[80%]"
      :type="show ? 'text' : 'password'"
      :ui="{
        base: 'bg-white text-black',
        trailing: 'pe-1'
      }"
    >
      <template #trailing>
        <UButton
          color="neutral"
          variant="link"
          size="sm"
          :icon="show ? 'i-lucide-eye-off' : 'i-lucide-eye'"
          :aria-label="show ? 'Hide password' : 'Show password'"
          :aria-pressed="show"
          aria-controls="password"
          @click="show = !show"
        />
      </template>
    </UInput>
    <UButton
      class="size-fit text-xl text-white rounded-md px-15 font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
      label="Log In"
    />
  </div>
</template>

<script setup>
const show = ref(false);
const password = ref('');
const username = ref('');
</script>

<style>
::-ms-reveal {
  display: none;
}
</style>