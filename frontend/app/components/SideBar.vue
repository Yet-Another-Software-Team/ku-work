<template>
  <div class="relative">
    <!-- Sidebar opens with button (mobile, ipad, small screens) -->
    <div class="block md:hidden">
      <USlideover side="left">
        <UButton label="Open" color="neutral" variant="subtle" />
        <template #header>
            <div>Icon here</div>
        </template>
        <template #body>
          <div class="h-full p-4 flex flex-col gap-2  bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A]">
            <UButton to="/student/dashboard" label="Dashboard" block />
            <UButton to="/student/profile" label="Profile" block />
            <UButton to="/student/jobboard" label="Job Board" block />
          </div>
        </template>
      </USlideover>
    </div>
    <!-- Sidebar opened (desktop) -->
    <div class="hidden md:block">
      <div
        class="fixed top-0 left-0 w-64 h-screen bg-linear-to-bl from-[#013B49] from-50% to-[#40DC7A]/90 shadow-md p-4 space-y-2"
      >
        <UButton to="/student/dashboard" label="Dashboard" block />
        <UButton to="/student/profile" label="Profile" block />
        <UButton to="/student/jobboard" label="Job Board" block />
      </div>
    </div>
  </div>
</template>
