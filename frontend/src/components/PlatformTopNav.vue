<template>
  <div class="platform-top-nav">
    <button class="brand-button" @click="goCourses">
      <span class="brand-mark">C</span>
      <span class="brand-copy">
        <strong>课程智学平台</strong>
        <small>Course Agent</small>
      </span>
    </button>

    <nav class="platform-top-links">
      <RouterLink to="/courses" class="platform-link">我的课程</RouterLink>
    </nav>

    <div class="platform-top-actions">
      <span class="platform-user">{{ auth.user.value?.username || '未登录用户' }}</span>
      <button class="button ghost compact" @click="handleLogout">退出登录</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter, RouterLink } from 'vue-router'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const auth = useAuth()

async function goCourses() {
  await router.push('/courses')
}

async function handleLogout() {
  if (auth.token.value) {
    await logout(auth.token.value)
  }
  auth.clear()
  await router.push('/login')
}
</script>
