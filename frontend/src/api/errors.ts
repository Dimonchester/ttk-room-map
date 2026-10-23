import type { components } from './schema'

export type Problem = components['schemas']['Problem']
export type FieldError = NonNullable<Problem['fields']>[number]

/** Единый тип ошибки API. UI опирается на `code`, а не на текст сообщения. */
export class ApiError extends Error {
  status: number
  code: string
  fields: FieldError[]
  currentVersion?: number
  retryAfter?: number

  constructor(init: {
    status: number
    code: string
    message: string
    fields?: FieldError[]
    currentVersion?: number
    retryAfter?: number
  }) {
    super(init.message)
    this.name = 'ApiError'
    this.status = init.status
    this.code = init.code
    this.fields = init.fields ?? []
    this.currentVersion = init.currentVersion
    this.retryAfter = init.retryAfter
  }

  /** Ошибка конкретного поля формы по JSON-path, например `floors[0].pois[2].title` */
  fieldError(path: string): FieldError | undefined {
    return this.fields.find((f) => f.field === path)
  }

  get isNetwork(): boolean {
    return this.status === 0
  }
}

type Result<T> = { data?: T; error?: unknown; response: Response }

/**
 * Превращает результат openapi-fetch в «данные или исключение ApiError».
 * Использование: const map = await unwrap(api.GET('/maps/{mapId}', { params: { path: { mapId } } }))
 */
export async function unwrap<T>(request: Promise<Result<T>>): Promise<T> {
  let result: Result<T>
  try {
    result = await request
  } catch {
    throw new ApiError({ status: 0, code: 'network_error', message: 'Нет соединения с сервером' })
  }

  const { data, error, response } = result
  if (response.ok) return data as T

  const problem = (typeof error === 'object' && error !== null ? error : {}) as Partial<Problem>
  throw new ApiError({
    status: response.status,
    code: problem.code ?? 'unknown_error',
    message: problem.detail ?? problem.title ?? `HTTP ${response.status}`,
    fields: problem.fields,
    currentVersion: problem.currentVersion,
    retryAfter: Number(response.headers.get('Retry-After')) || undefined,
  })
}
