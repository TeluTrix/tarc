<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '@/auth'

const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await register(email.value, password.value)
    await router.push('/profile')
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Unable to register'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <h1>Register</h1>
  <form @submit.prevent="submit">
    <label>
      Email
      <input v-model="email" type="email" autocomplete="email" required />
    </label>
    <label>
      Password
      <input v-model="password" type="password" autocomplete="new-password" minlength="8" required />
    </label>
    <p v-if="error" role="alert">{{ error }}</p>
    <button type="submit" :disabled="submitting">Register</button>
  </form>
  <p>Already registered? <RouterLink to="/login">Log in</RouterLink></p>
</template>