import { request } from './http'
import type { LoginResult, UserVO } from '@/types/user'

export interface RegisterPayload {
  username: string
  password: string
  phone?: string
}

export interface LoginPayload {
  username: string
  password: string
}

export function register(payload: RegisterPayload) {
  return request<UserVO>('/api/users/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function login(payload: LoginPayload) {
  return request<LoginResult>('/api/users/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function logout(token: string) {
  return request<null>('/api/users/logout', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function me(token: string) {
  return request<UserVO>('/api/users/me', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}
