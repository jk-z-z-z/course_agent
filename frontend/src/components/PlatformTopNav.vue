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
      <button class="platform-link ghost-link" type="button">创建课程</button>
      <button class="platform-link ghost-link" type="button">消息中心</button>
    </nav>

    <div class="platform-top-actions">
      <button class="platform-icon-button" type="button" aria-label="搜索">
        <span>⌕</span>
      </button>
      <button class="platform-icon-button has-dot" type="button" aria-label="通知">
        <span>◌</span>
      </button>

      <div class="platform-user-menu">
        <button class="platform-user-trigger" type="button" @click="toggleMenu">
          <span class="platform-avatar">{{ userInitial }}</span>
          <span class="platform-user-copy">
            <strong>{{ auth.user.value?.username || '未登录用户' }}</strong>
            <small>教师工作台</small>
          </span>
          <span class="platform-caret">⌄</span>
        </button>

        <div v-if="menuOpen" class="platform-user-dropdown">
          <button class="platform-dropdown-item" type="button" @click="goCourses">回到课程列表</button>
          <button class="platform-dropdown-item danger" type="button" @click="handleLogout">退出登录</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const auth = useAuth()
const menuOpen = ref(false)
const userInitial = computed(() => auth.user.value?.username?.slice(0, 1).toUpperCase() || 'U')

async function goCourses() {
  menuOpen.value = false
  await router.push('/courses')
}

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

async function handleLogout() {
  menuOpen.value = false
  if (auth.token.value) {
    await logout(auth.token.value)
  }
  auth.clear()
  await router.push('/login')
}
</script>
