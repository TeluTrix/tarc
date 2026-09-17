<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '@/auth'

const router = useRouter()
const route = useRoute()
const email = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await login(email.value, password.value)
    await router.push(typeof route.query.redirect === 'string' ? route.query.redirect : '/profile')
  } catch (reason) {
    error.value = reason instanceof Error ? reason.message : 'Unable to log in'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <h1>Log in</h1>
  <form @submit.prevent="submit">
    <label>
      Email
      <input v-model="email" type="email" autocomplete="email" required />
    </label>
    <label>
      Password
      <input v-model="password" type="password" autocomplete="current-password" required />
    </label>
    <p v-if="error" role="alert">{{ error }}</p>
    <button type="submit" :disabled="submitting">Log in</button>
  </form>
  <p>No account? <RouterLink to="/register">Register</RouterLink></p>
</template>