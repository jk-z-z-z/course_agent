export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
    ...init,
  })

  const responseText = await response.text()
  const payload = parsePayload(responseText, response)
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.message ?? '请求失败')
  }
  return payload.data as T
}

interface ApiPayload<T = unknown> {
  code: number
  message?: string
  data: T
}

function parsePayload(responseText: string, response: Response): ApiPayload {
  if (!responseText.trim()) {
    if (response.ok) {
      return { code: 0, data: null }
    }
    return { code: response.status, message: `请求失败（${response.status}）`, data: null }
  }

  try {
    return JSON.parse(responseText) as ApiPayload
  } catch {
    if (response.status === 404) {
      return { code: 404, message: '接口不存在，请确认后端服务已重启到最新代码', data: null }
    }
    const message = response.ok ? '接口返回格式错误' : responseText.trim()
    return { code: response.status || 500, message, data: null }
  }
}
