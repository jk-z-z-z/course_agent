<template>
  <main class="page-shell">
    <section class="card hero-card">
      <p class="eyebrow">Course Agent</p>
      <h1>课程平台前端</h1>
      <p class="lead">当前已接入用户模块，可登录、注册并持久化登录态。</p>

      <div v-if="auth.isLoggedIn.value" class="panel">
        <div class="profile-row">
          <div>
            <p class="label">当前用户</p>
            <p class="value">{{ auth.user.value?.username }}</p>
          </div>
          <button class="button ghost" @click="handleLogout">退出登录</button>
        </div>
      </div>

      <div v-else class="panel">
        <p class="label">未登录</p>
        <RouterLink class="button primary" to="/login">前往登录</RouterLink>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

const auth = useAuth()
const router = useRouter()

async function handleLogout() {
  if (!auth.token.value) return
  await logout(auth.token.value)
  auth.clear()
  await router.push('/login')
}
</script>
