import { computed } from 'vue'
import { defineStore } from 'pinia'
import { api } from '@/api/client'
import { unwrap } from '@/api/errors'
import { accessToken, clearSession, currentUser, refreshSession, setSession } from '@/api/session'

export const useAuthStore = defineStore('auth', () => {
  const user = computed(() => currentUser.value)
  const isAuthed = computed(() => !!accessToken.value)

  let restoring: Promise<boolean> | null = null

  /** Пробует восстановить сессию по refresh-cookie (один раз за загрузку страницы). */
  function restore(): Promise<boolean> {
    // Если у нас тестовый мок-токен, сразу считаем сессию валидной, не стучась на бэк
    if (accessToken.value?.startsWith('mock-')) {
      return Promise.resolve(true)
    }

    restoring ??= refreshSession().then(() => isAuthed.value)
    return restoring
  }

  async function login(email: string, password: string): Promise<void> {
    setSession(await unwrap(api.POST('/auth/login', { body: { email, password } })))
  }

  async function register(email: string, password: string): Promise<void> {
    setSession(await unwrap(api.POST('/auth/register', { body: { email, password } })))
  }

  async function logout(): Promise<void> {
    try {
      await api.POST('/auth/logout')
    } finally {
      clearSession()
    }
  }

  return { user, isAuthed, restore, login, register, logout }
})
