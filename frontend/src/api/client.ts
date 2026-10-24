import createClient from 'openapi-fetch'
import type { paths } from './schema'
import { accessToken, refreshSession } from './session'

/**
 * Логирование ответа (парсит JSON, если он есть, и выводит в консоль)
 */
async function logResponse(method: string, url: string, response: Response): Promise<void> {
  const clone = response.clone()
  try {
    const data = await clone.json()
    if (response.ok) {
      console.log(` [API Success] ${method} ${url} (${response.status})`, data)
    } else {
      console.warn(` [API Error Response] ${method} ${url} (${response.status})`, data)
    }
  } catch {
    console.log(` [API Response] ${method} ${url} (${response.status}) [не JSON]`)
  }
}

/**
 * fetch с одной повторной попыткой после 401 + логирование запросов/ответов
 */
async function fetchWithRefresh(input: Request): Promise<Response> {
  const method = input.method
  const url = input.url
  console.log(` [API Request] Выполняется запрос: ${method} ${url}`)

  const retry = input.clone()
  let response: Response
  
  try {
    response = await fetch(input)
  } catch (err) {
    console.error(` [API Network Error] Сбой сети для ${method} ${url}`, err)
    throw err
  }

  if (response.status !== 401) {
    await logResponse(method, url, response)
    return response
  }

  const { pathname } = new URL(input.url)
  if (pathname.startsWith('/api/auth/') || pathname.startsWith('/api/public/')) {
    await logResponse(method, url, response)
    return response
  }

  console.log(` [API Auth] Получен 401, пробуем обновить токен через /api/auth/refresh...`)
  if (!(await refreshSession())) {
    await logResponse(method, url, response)
    return response
  }

  retry.headers.set('Authorization', `Bearer ${accessToken.value}`)
  const retryResponse = await fetch(retry)
  await logResponse(method, `${url} (retry)`, retryResponse)
  return retryResponse
}

/** Типизированный клиент: пути, параметры и ответы берутся из openapi.yaml */
export const api = createClient<paths>({ baseUrl: '/api', fetch: fetchWithRefresh })

api.use({
  onRequest({ request }) {
    if (accessToken.value) request.headers.set('Authorization', `Bearer ${accessToken.value}`)
    return request
  },
})