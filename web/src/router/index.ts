import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '@/auth'
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'
import ProfileView from '@/views/ProfileView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/profile' },
    { path: '/login', component: LoginView, meta: { guestOnly: true } },
    { path: '/register', component: RegisterView, meta: { guestOnly: true } },
    { path: '/profile', component: ProfileView, meta: { requiresAuth: true } },
  ],
})

router.beforeEach(async (to) => {
  const authenticated = await isAuthenticated()

  if (to.meta.requiresAuth && !authenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly && authenticated) {
    return '/profile'
  }
})

export default router
