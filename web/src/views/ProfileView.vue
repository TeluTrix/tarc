<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { me, type User } from '@/auth'

const user = ref<User | null>(null)
const error = ref('')

onMounted(async () => {
  user.value = await me()
  if (!user.value) {
    error.value = 'Unable to load profile'
  }
})
</script>

<template>
  <h1>Profile</h1>
  <p v-if="error" role="alert">{{ error }}</p>
  <dl v-else-if="user">
    <dt>Email</dt>
    <dd>{{ user.email }}</dd>
    <dt>Role</dt>
    <dd>{{ user.role }}</dd>
    <dt>ID</dt>
    <dd>{{ user.id }}</dd>
  </dl>
</template>