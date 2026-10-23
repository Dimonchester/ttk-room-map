import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guestOnly?: boolean
  }
}

// Все страницы грузятся лениво: посетитель карты (/m, /embed) не скачивает код редактора.
export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: { name: 'maps' } },
    { path: '/login', name: 'login', component: () => import('@/pages/LoginPage.vue'), meta: { guestOnly: true } },
    { path: '/register', name: 'register', component: () => import('@/pages/RegisterPage.vue'), meta: { guestOnly: true } },

    // организатор
    { path: '/maps', name: 'maps', component: () => import('@/pages/MapsPage.vue'), meta: { requiresAuth: true } },
    {
      path: '/maps/:id/edit',
      name: 'editor',
      component: () => import('@/pages/EditorPage.vue'),
      props: true,
      meta: { requiresAuth: true },
    },

    // посетитель
    { path: '/m/:slug', name: 'viewer', component: () => import('@/pages/ViewerPage.vue'), props: true },
    { path: '/embed/:slug', name: 'embed', component: () => import('@/pages/EmbedPage.vue'), props: true },

    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/pages/NotFoundPage.vue') },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth) {
    if (auth.isAuthed || (await auth.restore())) return true
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly) {
    if (!auth.isAuthed) await auth.restore()
    if (auth.isAuthed) return { name: 'maps' }
  }

  return true
})
