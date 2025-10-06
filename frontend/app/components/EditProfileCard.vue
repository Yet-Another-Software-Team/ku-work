<template>
  <div class="rounded-xl dark:bg-[#001F26] max-h-[90vh] overflow-y-auto p-5 w-full max-w-2xl">
    <!-- Avatar -->
    <div class="w-full flex justify-center mb-10">
      <div class="relative">
        <div class="w-24 h-24 sm:w-28 sm:h-28 rounded-full bg-gray-200 overflow-hidden shadow">
          <img v-if="avatarPreview" :src="avatarPreview" alt="Avatar" class="w-full h-full object-cover" />
          <Icon v-else name="ic:baseline-account-circle" class="w-full h-full text-gray-400" />
        </div>
        <button
          type="button"
          class="absolute -right-1 bottom-0 translate-y-1 inline-flex items-center justify-center w-9 h-9 rounded-full bg-emerald-500 text-white shadow hover:bg-emerald-600 ring-4 ring-white/60"
          @click="pickAvatar"
          aria-label="Change avatar"
        >
          <Icon name="material-symbols:edit-square-outline-rounded" class="w-5 h-5" />
        </button>
        <input ref="avatarInput" type="file" accept="image/*" class="hidden" @change="onAvatarSelected" />
      </div>
    </div>

    <form class="grid grid-cols-1 md:grid-cols-2 gap-4" @submit.prevent="onSubmit">
      <div class="md:col-span-2">
        <label class="block text-primary-800 dark:text-primary font-semibold mb-1">Name</label>
        <div class="rounded-lg border border-emerald-700/50 bg-gray-100 px-4 py-2 text-gray-900 dark:border-emerald-700/40 dark:bg-[#013B49] dark:text-white">
          {{ props.profile.name || '�' }}
        </div>
      </div>

      <div class="col-span-1">
        <label for="dob" class="block text-primary-800 dark:text-primary font-semibold mb-1">Date of Birth</label>
        <UInput
          id="dob"
          v-model="form.dob"
          type="date"
          icon="i-heroicons-calendar-20-solid"
          placeholder="Birth date"
          class="w-full rounded-lg border border-emerald-700/70 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
      </div>

      <div class="col-span-1">
        <label for="phone" class="block text-primary-800 dark:text-primary font-semibold mb-1">Phone</label>
        <UInput
          id="phone"
          v-model="form.phone"
          placeholder="Phone"
          class="w-full rounded-lg border border-emerald-700/70 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
      </div>

      <div class="col-span-1">
        <label for="github" class="block text-primary-800 dark:text-primary font-semibold mb-1">Github</label>
        <UInput
          id="github"
          v-model="form.github"
          placeholder="Github"
          class="w-full rounded-lg border border-emerald-700/70 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
      </div>

      <div class="col-span-1">
        <label for="linkedin" class="block text-primary-800 dark:text-primary font-semibold mb-1">LinkedIn</label>
        <UInput
          id="linkedin"
          v-model="form.linkedin"
          placeholder="LinkedIn"
          class="w-full rounded-lg border border-emerald-700/70 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
      </div>

      <div class="md:col-span-2">
        <label for="aboutMe" class="block text-primary-800 dark:text-primary font-semibold mb-1">About me</label>
        <div class="rounded-lg border border-emerald-700/70 bg-white dark:bg-[#013B49]">
          <UTextarea
            id="aboutMe"
            v-model="form.aboutMe"
            placeholder="About me"
            :rows="6"
            class="w-full bg-transparent border-0 focus:outline-none resize-none text-gray-900 dark:text-white"
          />
        </div>
      </div>

      <!-- Save & Discard -->
      <div class="md:col-span-2 flex flex-wrap justify-end gap-3 pt-2 w-full">
        <!-- Discard -->
        <UButton
          type="button"
          variant="outline"
          color="neutral"
          class="rounded-md px-4"
          @click="showDiscardConfirm = true"
        >
          Discard
        </UButton>
          <!-- Confirm Discard Modal -->
    <UModal
      v-model:open="showDiscardConfirm"
      title="Discard changes?"
      :dismissible="false"
      :ui="{
        title: 'text-xl font-semibold text-primary-800 dark:text-primary',
        container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
        overlay: 'fixed inset-0 bg-black/50'
      }"
    >
      <template #body>
        <p class="dark:text-white">This will discard your current inputs. Are you sure?</p>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton variant="outline" color="neutral" @click="cancelDiscard">
            Cancel
          </UButton>
          <UButton color="primary" @click="confirmDiscard">
            Discard
          </UButton>
        </div>
      </template>
    </UModal>
        <!-- Save -->
        <UButton type="submit" color="primary" class="rounded-md px-5">
          Save
        </UButton>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'

const { add: addToast } = useToast()
const showDiscardConfirm = ref(false)

type StudentProfile = {
  name?: string
  birthDate?: string
  phone?: string
  github?: string
  linkedIn?: string
  aboutMe?: string
  photo?: string
}

const props = defineProps<{ profile: StudentProfile }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved', payload: StudentProfile & { name?: string; _avatarFile?: File | null }): void
}>()

const form = reactive({
  dob: '',
  phone: '',
  github: '',
  linkedin: '',
  aboutMe: ''
})

const avatarInput = ref<HTMLInputElement | null>(null)
const avatarPreview = ref<string>('')
const avatarFile = ref<File | null>(null)

function resetForm(profile: StudentProfile) {
  form.dob = profile.birthDate ? profile.birthDate.slice(0, 10) : ''
  form.phone = profile.phone ?? ''
  form.github = profile.github ?? ''
  form.linkedin = profile.linkedIn ?? ''
  form.aboutMe = profile.aboutMe ?? ''

  if (avatarPreview.value?.startsWith('blob:')) {
    URL.revokeObjectURL(avatarPreview.value)
  }
  avatarPreview.value = profile.photo ?? ''
  avatarFile.value = null
}

watch(
  () => props.profile,
  (profile) => {
    resetForm(profile)
  },
  { immediate: true, deep: true }
)

function pickAvatar() {
  avatarInput.value?.click()
}

function onAvatarSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  if (avatarPreview.value?.startsWith('blob:')) {
    URL.revokeObjectURL(avatarPreview.value)
  }

  avatarFile.value = file
  avatarPreview.value = URL.createObjectURL(file)
}

function cancelDiscard() {
  showDiscardConfirm.value = false
}

function confirmDiscard() {
  showDiscardConfirm.value = false
  resetForm(props.profile)
  addToast({
    title: 'Changes discarded',
    description: 'Old data reloaded.',
    color: 'success'
  })
  emit('close')
}

function onSubmit() {
  emit('saved', {
    ...props.profile,
    birthDate: form.dob || props.profile.birthDate,
    phone: form.phone,
    github: form.github,
    linkedIn: form.linkedin,
    aboutMe: form.aboutMe,
    photo: avatarPreview.value || props.profile.photo || '',
    _avatarFile: avatarFile.value
  })
}

onBeforeUnmount(() => {
  if (avatarPreview.value?.startsWith('blob:')) {
    URL.revokeObjectURL(avatarPreview.value)
  }
})
</script>
