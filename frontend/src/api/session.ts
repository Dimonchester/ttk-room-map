import { shallowRef } from 'vue'
import type { AuthResponse, User } from './types'

// Access-токен живёт ТОЛЬКО в памяти (не localStorage): так его не украсть через XSS.
// Refresh лежит в httpOnly-cookie, и после перезагрузки страницы мы получаем новый access через /auth/refresh.
export const accessToken = shallowRef<string | null>(null)
export const currentUser = shallowRef<User | null>(null)

export function setSession(auth: AuthResponse): void {
  accessToken.value = auth.accessToken
  currentUser.value = auth.user
}

export function clearSession(): void {
  accessToken.value = null
  currentUser.value = null
}

export function setMockSession(): void {
  accessToken.value = 'mock-test-token-12345'
  currentUser.value = {
    id: 'd3b07384-d113-4ec2-a4f9-2603f29b465d',
    email: 'admin@ttk.ru',
  } as User
}


let inFlight: Promise<boolean> | null = null

/** Обновляет access-токен. Параллельные вызовы объединяются в один запрос (single-flight). */
export function refreshSession(): Promise<boolean> {
  inFlight ??= doRefresh().finally(() => {
    inFlight = null
  })
  return inFlight
}

async function doRefresh(): Promise<boolean> {
  try {
    const res = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'same-origin' })
    if (!res.ok) {
      clearSession()
      return false
    }
    setSession((await res.json()) as AuthResponse)
    return true
  } catch {
    // нет сети: не разлогиниваем, просто сообщаем, что обновить не удалось
    return false
  }
}
