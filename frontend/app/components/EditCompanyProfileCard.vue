<template>
  <div class="rounded-xl max-h-[90vh] overflow-y-auto">
    <div class="p-4 sm:p-5 md:p-6">
      <!-- Header: banner + avatar -->
      <div class="relative">
        <!-- Banner -->
        <div class="relative h-28 sm:h-32 md:h-36 w-full">
          <img
            :src="bannerPreview"
            alt="Company banner"
            class="h-full w-full object-cover"
          />

          <!-- Edit banner -->
          <button
            type="button"
            class="absolute right-3 top-3 inline-flex items-center justify-center w-8 h-8 rounded-full bg-primary-500 text-white shadow hover:bg-primary-600"
            @click="triggerBannerPicker"
            aria-label="Change banner"
          >
            <Icon name="material-symbols:edit-square-outline-rounded" class="w-5 h-5" />
          </button>
          <input
            ref="bannerInput"
            type="file"
            accept="image/*"
            class="hidden"
            @change="onBannerSelected"
          />
        </div>

        <!-- Logo -->
        <div class="absolute -bottom-10 left-4 z-20">
          <div class="relative">
            <div class="w-20 h-20 sm:w-24 sm:h-24 rounded-full bg-gray-200 overflow-hidden shadow">
              <img
                v-if="logoPreview"
                :src="logoPreview"
                alt="Company logo"
                class="w-full h-full object-cover"
              />
              <Icon v-else name="ic:baseline-account-circle" class="w-full h-full text-gray-400" />
            </div>
            <button
              type="button"
              class="absolute -right-1 sm:right-0 bottom-0 translate-y-1 inline-flex items-center justify-center w-8 h-8 rounded-full bg-primary-500 text-white shadow hover:bg-primary-600"
              @click="triggerLogoPicker"
              aria-label="Change logo"
            >
              <Icon name="material-symbols:edit-square-outline-rounded" class="w-5 h-5" />
            </button>
            <input
              ref="logoInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="onLogoSelected"
            />
          </div>
        </div>
      </div>
      <div class="pt-12"></div>

      <!-- Form -->
      <form class="mt-2 grid grid-cols-1 md:grid-cols-2 gap-2" @submit.prevent="save">
        <!-- Company Name -->
        <div class="md:col-span-2 text-primary-800 dark:text-primary font-semibold ">
          Company Name
        </div>
        <UInput
          v-model="form.companyName"
          placeholder="Company Name"
          class="md:col-span-2 rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />

        <!-- Address -->
        <div class="md:col-span-2 text-primary-800 dark:text-primary font-semibold mt-1">
          Address
        </div>
        <UInput
          v-model="form.address"
          placeholder="Address"
          class="md:col-span-2 rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />

        <!-- Location -->
        <div class="md:col-span-2 text-primary-800 dark:text-primary font-semibold mt-1">
          Location
        </div>
        <UInput
          v-model="form.city"
          placeholder="City"
          class="rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
        <UInput
          v-model="form.country"
          placeholder="Country"
           
          class="rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />

        <!-- Connections -->
        <div class="md:col-span-2 text-primary-800 dark:text-primary font-semibold  mt-1">
          Connections
        </div>
        <UInput
          v-model="form.phone"
          placeholder="Phone"
          type="tel"
          class="rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
        <UInput
          v-model="form.mail"
          placeholder="Mail"
          type="email"
          class="rounded-md border border-gray-500 bg-white dark:bg-[#013B49] text-gray-900 dark:text-white"
        />
        <!-- About us -->
        <div class="md:col-span-2 text-primary-800 dark:text-primary font-semibold  mt-1">
          About us
        </div>
        <div
          class="md:col-span-2 max-h-56 md:max-h-64 overflow-y-auto rounded-md border border-gray-500  bg-white dark:bg-[#013B49]"
        >
          <UTextarea
            v-model="form.aboutUs"
            placeholder="About us"
            :rows="6"
            class="w-full bg-transparent border border-gray-500 focus:outline-none resize-none text-gray-900 dark:text-white"
          />
        </div>
        <div></div>
        <!-- Save & Discard -->
            <div class="grid grid-cols-6 w-full">
                <div class="col-span-12 md:col-start-9 md:col-span-4 flex justify-end gap-x-3 ml-auto">
                    <!-- Discard (opens confirm modal) -->
                    <UButton
                        class="size-fit text-xl rounded-md px-15 font-medium hover:bg-gray-800 hover:cursor-pointer"
                        variant="outline" 
                        color="neutral"
                        label="Discard"
                        @click="showDiscardConfirm = true"
                    />

                    <!-- Confirm Discard Modal -->
                    <UModal
                        v-model:open="showDiscardConfirm"
                        title="Discard Change?"
                        :dismissible="false"
                        :ui="{
                            title: 'text-xl font-semibold text-primary-800 dark:text-primary',
                            container: 'fixed inset-0 z-[100] flex items-center justify-center p-4',
                            overlay: 'fixed inset-0 bg-black/50',
                        }"
                    >
                        <template #body>
                            <div class="space-y-2">
                                <p class="dark:text-white">
                                    This will discard your current inputs. Are you sure?
                                </p>
                            </div>
                        </template>
                        <template #footer>
                            <div class="w-full flex justify-end gap-2">
                                <UButton variant="outline" color="neutral" @click="cancelDiscard">
                                    Cancel
                                </UButton>
                                <UButton color="primary" @click="confirmDiscard"> Discard </UButton>
                            </div>
                        </template>
                    </UModal>

                    <!-- Save -->
                    <UButton
                        class="size-fit text-xl text-white rounded-md px-15 font-medium bg-primary-500 hover:bg-primary-700 hover:cursor-pointer active:bg-primary-800"
                        type="submit"
                        label="Save"
                    />
                </div>
                </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onBeforeUnmount } from 'vue'
import * as z from 'zod'

type Profile = {
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

const props = defineProps<{ profile: Profile }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved', val: Profile & { _logoFile?: File | null; _bannerFile?: File | null }): void
}>()

const { add: addToast } = useToast()
const showDiscardConfirm = ref(false)
const form = ref({
  companyName: props.profile.name ?? '',
  address: props.profile.address ?? '',
  city: props.profile.city ?? '',
  country: props.profile.country ?? '',
  mail: props.profile.mail ?? '',
  phone: props.profile.phone ?? '',
  aboutUs: props.profile.aboutUs ?? ''
})

const errors = reactive<Record<string, string>>({
  companyName: '',
  address: '',
  city: '',
  country: '',
  mail: '',
  phone: '',
  aboutUs: ''
})

const schema = z.object({
  companyName: z.string().min(1, 'Company name is required'),
  address: z.string().min(1, 'Address is required'),
  city: z.string().min(1, 'City is required'),
  country: z.string().min(1, 'Country is required'),
  mail: z
    .string()
    .min(1, 'Email is required')
    .email('Email is invalid'),
  phone: z
    .string()
    .optional()
    .or(z.literal('')),
  aboutUs: z.string()
})

function validateField(fieldName: keyof typeof form.value, value: unknown) {
  try {
    schema.pick({ [fieldName]: true } as any).parse({ [fieldName]: value })
    if (typeof value === 'string' && value.trim() === '') {
      errors[fieldName as string] = 'This field is required'
      return false
    }
    errors[fieldName as string] = ''
    return true
  } catch (error) {
    if (error instanceof z.ZodError) {
      errors[fieldName as string] = error.issues[0]?.message ?? 'Invalid value'
    } else {
      errors[fieldName as string] = 'Invalid value'
    }
    return false
  }
}

watch(() => form.value.companyName, v => validateField('companyName', v))
watch(() => form.value.address,     v => validateField('address', v))
watch(() => form.value.city,        v => validateField('city', v))
watch(() => form.value.country,     v => validateField('country', v))
watch(() => form.value.mail,        v => validateField('mail', v))
watch(() => form.value.phone,       v => validateField('phone', v))
watch(() => form.value.aboutUs,     v => validateField('aboutUs', v))

function cancelDiscard() {
  showDiscardConfirm.value = false
}
function confirmDiscard() {
  showDiscardConfirm.value = false
  addToast({
    title: 'Changes discarded',
    description: 'Old data reloaded.',
    color: 'success'
  })
  emit('close')
}

const logoInput = ref<HTMLInputElement | null>(null)
const logoFile = ref<File | null>(null)
const logoPreview = ref<string>(props.profile.logo || '')

function triggerLogoPicker() {
  logoInput.value?.click()
}
function onLogoSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  const valid = ['image/png', 'image/jpeg', 'image/jpg', 'image/webp', 'image/gif'].includes(file.type)
  if (!valid) { addToast({ title: 'Unsupported file type', color: 'warning' }); target.value = ''; return }
  const maxBytes = 2 * 1024 * 1024
  if (file.size > maxBytes) { addToast({ title: 'File too large (max 2MB)', color: 'warning' }); target.value = ''; return }
  if (logoPreview.value?.startsWith('blob:')) URL.revokeObjectURL(logoPreview.value)
  logoFile.value = file
  logoPreview.value = URL.createObjectURL(file)
}

const bannerInput = ref<HTMLInputElement | null>(null)
const bannerFile = ref<File | null>(null)
const bannerPreview = ref<string>(props.profile.banner || '')

function triggerBannerPicker() {
  bannerInput.value?.click()
}
function onBannerSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  const valid = ['image/png', 'image/jpeg', 'image/jpg', 'image/webp'].includes(file.type)
  if (!valid) { addToast({ title: 'Unsupported banner type', color: 'warning' }); target.value = ''; return }
  const maxBytes = 5 * 1024 * 1024
  if (file.size > maxBytes) { addToast({ title: 'Banner too large (max 5MB)', color: 'warning' }); target.value = ''; return }
  if (bannerPreview.value?.startsWith('blob:')) URL.revokeObjectURL(bannerPreview.value)
  bannerFile.value = file
  bannerPreview.value = URL.createObjectURL(file)
}

onBeforeUnmount(() => {
  if (logoPreview.value?.startsWith('blob:')) URL.revokeObjectURL(logoPreview.value)
  if (bannerPreview.value?.startsWith('blob:')) URL.revokeObjectURL(bannerPreview.value)
})

function onSubmit() {
  const result = schema.safeParse(form.value)
  if (!result.success) {
    for (const issue of result.error.issues) {
      const key = issue.path?.[0]
      if (typeof key === 'string' && key in errors) {
        errors[key] = issue.message
      }
    }
    addToast({
      title: 'Form submission failed',
      description: 'Please check the highlighted errors and try again.',
      color: 'warning'
    })
    return
  }

  // success: emit payload + files, toast, and close
  addToast({
    title: 'Saved',
    description: 'Company profile has been updated.',
    color: 'success'
  })

  emit('saved', {
    ...props.profile,
    name: form.value.companyName,
    address: form.value.address,
    city: form.value.city,
    country: form.value.country,
    mail: form.value.mail,
    phone: form.value.phone,
    aboutUs: form.value.aboutUs,
    banner: bannerPreview.value || props.profile.banner || '',
    logo: logoPreview.value || props.profile.logo || '',
    _logoFile: logoFile.value,
    _bannerFile: bannerFile.value
  })

  emit('close')
}

const save = onSubmit
</script>

