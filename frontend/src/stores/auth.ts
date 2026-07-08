import { computed, ref } from 'vue'
import type { LoginResult, UserVO } from '@/types/user'

const storageTokenKey = 'course_agent_token'
const storageUserKey = 'course_agent_user'

const token = ref(localStorage.getItem(storageTokenKey) ?? '')
const user = ref<UserVO | null>(readUser())

function readUser(): UserVO | null {
  const raw = localStorage.getItem(storageUserKey)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserVO
  } catch {
    return null
  }
}

function persist(auth: LoginResult) {
  token.value = auth.token
  user.value = auth.user
  localStorage.setItem(storageTokenKey, auth.token)
  localStorage.setItem(storageUserKey, JSON.stringify(auth.user))
}

function clear() {
  token.value = ''
  user.value = null
  localStorage.removeItem(storageTokenKey)
  localStorage.removeItem(storageUserKey)
}

export function useAuthStore() {
  return {
    token: computed(() => token.value),
    user: computed(() => user.value),
    isLoggedIn: computed(() => Boolean(token.value)),
    persist,
    clear,
  }
}
