export interface UserVO {
  id: number
  username: string
  phone?: string
  status: string
}

export interface LoginResult {
  token: string
  expiredAt: string
  user: UserVO
}

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}
