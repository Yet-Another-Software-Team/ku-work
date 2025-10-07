<template>
  <div class="rounded-xl max-h-[90vh] overflow-y-auto">
    <div class="p-4 sm:p-5 md:p-6">
      <div class="relative">
        <div class="relative h-28 sm:h-32 md:h-36 w-full">
          <img :src="bannerPreview" alt="Company banner" class="h-full w-full object-cover" />
          <button
            type="button"
            class="absolute right-3 top-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-white shadow hover:bg-primary-600 hover:cursor-pointer"
            @click="triggerBannerPicker"
            aria-label="Change banner"
          >
            <Icon name="material-symbols:edit-square-outline-rounded" class="h-5 w-5" />
          </button>
          <input ref="bannerInput" type="file" accept="image/*" class="hidden" @change="onBannerSelected" />
        </div>

        <div class="absolute -bottom-10 left-4 z-20">
          <div class="relative">
            <div class="h-20 w-20 overflow-hidden rounded-full bg-gray-200 shadow sm:h-24 sm:w-24">
              <img v-if="logoPreview" :src="logoPreview" alt="Company logo" class="h-full w-full object-cover" />
              <Icon v-else name="ic:baseline-account-circle" class="h-full w-full text-gray-400" />
            </div>
            <button
              type="button"
              class="absolute -right-1 bottom-0 translate-y-1 inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-white shadow hover:bg-primary-600 hover:cursor-pointer sm:right-0"
              @click="triggerLogoPicker"
              aria-label="Change logo"
            >
              <Icon name="material-symbols:edit-square-outline-rounded" class="h-5 w-5" />
            </button>
            <input ref="logoInput" type="file" accept="image/*" class="hidden" @change="onLogoSelected" />
          </div>
        </div>
      </div>

      <div class="pt-12" />

      <form class="mt-2 grid grid-cols-1 gap-2 md:grid-cols-2" @submit.prevent="handleSubmit">
        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">Company Name *</label>
          <UInput
            v-model="form.companyName"
            placeholder="Company Name"
            class="mt-1 w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
          />
          <p v-if="errors.companyName" class="mt-1 text-sm text-red-500">{{ errors.companyName }}</p>
        </div>

        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">Company Email *</label>
          <UInput
            v-model="form.mail"
            placeholder="Company Email"
            type="email"
            class="mt-1 w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
          />
          <p v-if="errors.mail" class="mt-1 text-sm text-red-500">{{ errors.mail }}</p>
        </div>

        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">Phone *</label>
          <UInput
            v-model="form.phone"
            placeholder="Phone"
            type="tel"
            class="mt-1 w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
          />
          <p v-if="errors.phone" class="mt-1 text-sm text-red-500">{{ errors.phone }}</p>
        </div>

        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">Address *</label>
          <UInput
            v-model="form.address"
            placeholder="Address"
            class="mt-1 w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
          />
          <p v-if="errors.address" class="mt-1 text-sm text-red-500">{{ errors.address }}</p>
        </div>

        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">Location *</label>
          <div class="mt-1 grid grid-cols-1 gap-2 md:grid-cols-2">
            <div>
              <UInput
                v-model="form.city"
                placeholder="City"
                class="w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
              />
              <p v-if="errors.city" class="mt-1 text-sm text-red-500">{{ errors.city }}</p>
            </div>
            <div>
              <UInput
                v-model="form.country"
                placeholder="Country"
                class="w-full rounded-md border border-gray-500 bg-white text-gray-900 dark:bg-[#013B49] dark:text-white"
              />
              <p v-if="errors.country" class="mt-1 text-sm text-red-500">{{ errors.country }}</p>
            </div>
          </div>
        </div>

        <div class="md:col-span-2">
          <label class="block font-semibold text-primary-800 dark:text-primary">About us</label>
          <div class="mt-1 rounded-md border border-gray-500 bg-white dark:bg-[#013B49]">
            <UTextarea
              v-model="form.aboutUs"
              placeholder="About us"
              :rows="6"
              class="w-full bg-transparent text-gray-900 dark:text-white"
            />
          </div>
          <p v-if="errors.aboutUs" class="mt-1 text-sm text-red-500">{{ errors.aboutUs }}</p>
        </div>

        <div class="md:col-span-2 flex flex-wrap justify-end gap-3 pt-4">
          <UButton
            type="button"
            variant="outline"
            color="neutral"
            class="rounded-md px-4"
            @click="openDiscardModal"
          >
            Discard
          </UButton>
          <UButton type="submit" color="primary" class="rounded-md px-5">
            Save
          </UButton>
        </div>
      </form>
    </div>
  </div>

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
        <UButton variant="outline" color="neutral" @click="hideDiscardModal">Cancel</UButton>
        <UButton color="primary" @click="confirmDiscard">Discard</UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import * as z from 'zod'

const { add: addToast } = useToast()

interface Profile {
  name?: string
  address?: string
  website?: string
  banner?: string
  logo?: string
  aboutUs?: string
  city?: string
  country?: string
  mail?: string
  phone?: string
}

interface FormState {
  companyName: string
  mail: string
  phone: string
  address: string
  city: string
  country: string
  aboutUs: string
}

type FormKey = keyof FormState

type SavedPayload = Profile & {
  _logoFile?: File | null
  _bannerFile?: File | null
}

const props = defineProps<{ profile: Profile }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved', val: SavedPayload): void
}>()

const PHONE_REGEX = /^\+(?:[1-9]\d{0,2})\d{4,14}$/
const showDiscardConfirm = ref(false)
const form = reactive<FormState>({
  companyName: '',
  mail: '',
  phone: '',
  address: '',
  city: '',
  country: '',
  aboutUs: ''
})

const errors = reactive<Record<FormKey, string>>({
  companyName: '',
  mail: '',
  phone: '',
  address: '',
  city: '',
  country: '',
  aboutUs: ''
})

const schema = z.object({
  companyName: z.string().trim().min(2, 'Name must be at least 2 characters'),
  mail: z.string().trim().email('Please enter a valid email address'),
  phone: z.string().trim().regex(PHONE_REGEX, 'Please enter a valid phone number'),
  address: z.string().trim().min(5, 'Address must be at least 5 characters'),
  city: z.string().trim().min(2, 'City must be at least 2 characters'),
  country: z.string().trim().min(2, 'Country must be at least 2 characters'),
  aboutUs: z.string().max(2000, 'About us must be 2,000 characters or less').optional()
})

const logoInput = ref<HTMLInputElement | null>(null)
const logoFile = ref<File | null>(null)
const logoPreview = ref<string>('')

const bannerInput = ref<HTMLInputElement | null>(null)
const bannerFile = ref<File | null>(null)
const bannerPreview = ref<string>('')

watch(
  () => props.profile,
  (profile) => resetForm(profile),
  { deep: true, immediate: true }
)

function resetForm(profile: Profile) {
  form.companyName = profile.name ?? ''
  form.mail = profile.mail ?? ''
  form.phone = profile.phone ?? ''
  form.address = profile.address ?? ''
  form.city = profile.city ?? ''
  form.country = profile.country ?? ''
  form.aboutUs = profile.aboutUs ?? ''

  updateLogoPreview(profile.logo ?? '')
  updateBannerPreview(profile.banner ?? '')
  logoFile.value = null
  bannerFile.value = null
  clearErrors()
  runAllValidations()
}

function runAllValidations() {
  ;(Object.keys(form) as FormKey[]).forEach((key) => validateField(key, form[key]))
}

function clearErrors() {
  ;(Object.keys(errors) as FormKey[]).forEach((key) => (errors[key] = ''))
}

function validateField(field: FormKey, value: string) {
  const result = (schema.shape[field] as z.ZodTypeAny).safeParse(value)
  errors[field] = result.success ? '' : result.error.issues[0]?.message ?? 'Invalid value'
  return result.success
}

;(Object.keys(form) as FormKey[]).forEach((key) => {
  watch(
    () => form[key],
    (value) => validateField(key, value)
  )
})

function triggerLogoPicker() {
  logoInput.value?.click()
}

function onLogoSelected(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (!validateLogo(file)) {
    target.value = ''
    return
  }

  updateLogoPreview(URL.createObjectURL(file))
  logoFile.value = file
}

function validateLogo(file: File) {
  const valid = ['image/png', 'image/jpeg', 'image/jpg', 'image/webp', 'image/gif']
  if (!valid.includes(file.type)) {
    addToast({ title: 'Unsupported file type', color: 'warning' })
    return false
  }
  if (file.size > 2 * 1024 * 1024) {
    addToast({ title: 'File too large (max 2MB)', color: 'warning' })
    return false
  }
  return true
}

function triggerBannerPicker() {
  bannerInput.value?.click()
}

function onBannerSelected(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (!validateBanner(file)) {
    target.value = ''
    return
  }

  updateBannerPreview(URL.createObjectURL(file))
  bannerFile.value = file
}

function validateBanner(file: File) {
  const valid = ['image/png', 'image/jpeg', 'image/jpg', 'image/webp']
  if (!valid.includes(file.type)) {
    addToast({ title: 'Unsupported banner type', color: 'warning' })
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    addToast({ title: 'Banner too large (max 5MB)', color: 'warning' })
    return false
  }
  return true
}

function updateLogoPreview(source: string) {
  if (logoPreview.value?.startsWith('blob:')) URL.revokeObjectURL(logoPreview.value)
  logoPreview.value = source
}

function updateBannerPreview(source: string) {
  if (bannerPreview.value?.startsWith('blob:')) URL.revokeObjectURL(bannerPreview.value)
  bannerPreview.value = source
}

function openDiscardModal() {
  showDiscardConfirm.value = true
}

function hideDiscardModal() {
  showDiscardConfirm.value = false
}

function confirmDiscard() {
  hideDiscardModal()
  resetForm(props.profile)
  addToast({ title: 'Changes discarded', description: 'Old data reloaded.', color: 'success' })
  emit('close')
}

function handleSubmit() {
  const result = schema.safeParse({ ...form })
  if (!result.success) {
    result.error.issues.forEach((issue) => {
      const key = issue.path[0]
      if (typeof key === 'string' && key in errors) {
        errors[key as FormKey] = issue.message
      }
    })
    addToast({ title: 'Form submission failed', description: 'Please check the errors and try again.', color: 'warning' })
    return
  }

  addToast({ title: 'Saved', description: 'Company profile has been updated.', color: 'success' })

  emit('saved', {
    ...props.profile,
    name: result.data.companyName,
    mail: result.data.mail,
    phone: result.data.phone,
    address: result.data.address,
    city: result.data.city,
    country: result.data.country,
    aboutUs: result.data.aboutUs ?? '',
    banner: bannerPreview.value || props.profile.banner || '',
    logo: logoPreview.value || props.profile.logo || '',
    _logoFile: logoFile.value,
    _bannerFile: bannerFile.value
  })

  emit('close')
}

onBeforeUnmount(() => {
  if (logoPreview.value?.startsWith('blob:')) URL.revokeObjectURL(logoPreview.value)
  if (bannerPreview.value?.startsWith('blob:')) URL.revokeObjectURL(bannerPreview.value)
})
</script>
