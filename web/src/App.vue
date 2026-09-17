<template>
  <nav>
    <RouterLink to="/profile">Profile</RouterLink>
    <span> | </span>
    <button v-if="authenticated" type="button" @click="logout">Log out</button>
  </nav>
  <main>
    <RouterView />
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { logout as signOut, isAuthenticated } from '@/auth'

const router = useRouter()
const authenticated = ref(false)

isAuthenticated().then((value) => {
  authenticated.value = value
})

async function logout() {
  await signOut()
  authenticated.value = false
  await router.push('/login')
}
</script>
