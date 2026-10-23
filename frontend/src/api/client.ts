import createClient from 'openapi-fetch'
import type { paths } from './schema'
import { accessToken, refreshSession } from './session'

/**
 * fetch с одной повторной попыткой после 401: обновляем токен и повторяем исходный запрос.
 * /auth/* и /public/* не повторяем: там 401 — штатный ответ (неверный пароль, password_required).
 */
async function fetchWithRefresh(input: Request): Promise<Response> {
  const retry = input.clone() // тело запроса можно прочитать только один раз
  const response = await fetch(input)
  if (response.status !== 401) return response

  const { pathname } = new URL(input.url)
  if (pathname.startsWith('/api/auth/') || pathname.startsWith('/api/public/')) return response

  if (!(await refreshSession())) return response

  retry.headers.set('Authorization', `Bearer ${accessToken.value}`)
  return fetch(retry)
}

/** Типизированный клиент: пути, параметры и ответы берутся из openapi.yaml */
export const api = createClient<paths>({ baseUrl: '/api', fetch: fetchWithRefresh })

api.use({
  onRequest({ request }) {
    if (accessToken.value) request.headers.set('Authorization', `Bearer ${accessToken.value}`)
    return request
  },
})
