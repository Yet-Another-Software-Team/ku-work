<template>
  <section class="w-64">
    <!-- Sidebar toggle with button (mobile, ipad, small screens) -->
    <USlideover
      side="left"
      :ui="{
        content:
          'bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A] w-64 block md:hidden',
      }"
    >
      <UButton label="Open" color="neutral" variant="subtle" />
      <template #content>
        <header class="border-none">
          <h1 class="text-2xl font-bold text-white mb-4">Icon here</h1>
        </header>
        <UButton
          v-for="item in getSidebarItems(props.isViewer)"
          :key="item.label"
          :icon="item.icon"
          :to="item.to"
          :label="item.label"
          :disabled="item.disable"
          variant="ghost"
          size="xl"
          class="border-none"
          block
        />
      </template>
    </USlideover>
    <!-- Sidebar expanded (desktop) -->
    <div
      class=" fixed top-0 left-0 w-64 h-screen hidden md:block bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A]/90 shadow-md p-4 space-y-2"
    >
      <header>
        <h1 class="text-2xl font-bold text-white mb-4">Icon here</h1>
      </header>
      <UButton
        v-for="item in getSidebarItems(props.isViewer)"
        :key="item.label"
        :icon="item.icon"
        :to="item.to"
        :label="item.label"
        :disabled="item.disable"
        variant="ghost"
        size="xl"
        block
      />
    </div>
  </section>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    isViewer: boolean;
  }>(),
  {
    isViewer: false,
  }
);

function getSidebarItems(
  isViewer: boolean
): Array<{ label: string; icon: string; to: string; disable: boolean }> {
  let items = [
    {
      label: "Job Board",
      icon: "ic:baseline-work",
      to: "/student/jobboard",
      disable: false,
    },
  ];
  if (!isViewer) {
    items.unshift({
      label: "Profile",
      icon: "ic:baseline-person",
      to: "/student/profile",
      disable: false,
    });
    items.push({
      label: "Dashboard",
      icon: "ic:baseline-dashboard",
      to: "/student/dashboard",
      disable: false,
    });
  } else {
    items.unshift({
      label: "Profile",
      icon: "ic:baseline-person",
      to: "/profile",
      disable: true,
    });
  }
  return items;
}
</script>
