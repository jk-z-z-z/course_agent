<template>
  <div class="platform-top-nav">
    <button class="brand-button" @click="goCourses">
      <span class="brand-mark">C</span>
      <span class="brand-copy">
        <strong>课程智学平台</strong>
        <small>Course Agent</small>
      </span>
    </button>

    <nav v-if="variant !== 'dashboard'" class="platform-top-links">
      <RouterLink to="/courses" class="platform-link">我的课程</RouterLink>
      <button class="platform-link ghost-link" type="button" @click="openCreateCourse">创建课程</button>
      <button class="platform-link ghost-link" type="button" @click="togglePanel('messages')">消息中心</button>
    </nav>

    <div class="platform-top-actions" :class="{ compact: variant === 'dashboard' }">
      <button
        v-if="variant !== 'dashboard'"
        class="platform-icon-button"
        type="button"
        aria-label="搜索"
        @click="togglePanel('search')"
      >
        <span>⌕</span>
      </button>
      <button
        v-if="variant !== 'dashboard'"
        class="platform-icon-button has-dot"
        type="button"
        aria-label="通知"
        @click="togglePanel('notifications')"
      >
        <span>◌</span>
      </button>

      <div class="platform-user-menu">
        <button
          class="platform-user-trigger"
          :class="{ dashboard: variant === 'dashboard' }"
          type="button"
          @click="toggleMenu"
        >
          <span class="platform-avatar">{{ userInitial }}</span>
          <span class="platform-user-copy">
            <strong>{{ variant === 'dashboard' ? '个人主页' : auth.user.value?.username || '未登录用户' }}</strong>
            <small>{{ variant === 'dashboard' ? auth.user.value?.username || '未登录用户' : '教师工作台' }}</small>
          </span>
          <span class="platform-caret">⌄</span>
        </button>

        <div v-if="menuOpen" class="platform-user-dropdown">
          <button class="platform-dropdown-item" type="button" @click="goCourses">回到课程列表</button>
          <button class="platform-dropdown-item danger" type="button" @click="handleLogout">退出登录</button>
        </div>
      </div>
    </div>

    <div v-if="activePanel && variant !== 'dashboard'" class="platform-floating-panel">
      <template v-if="activePanel === 'messages'">
        <p class="eyebrow">Messages</p>
        <h3>消息中心</h3>
        <p class="muted-copy">消息中心入口已接通，后续可以接真实消息列表。</p>
      </template>

      <template v-else-if="activePanel === 'notifications'">
        <p class="eyebrow">Notifications</p>
        <h3>通知</h3>
        <p class="muted-copy">当前没有新的系统通知。</p>
      </template>

      <template v-else>
        <p class="eyebrow">Search</p>
        <h3>全局搜索</h3>
        <p class="muted-copy">搜索交互已激活，后续可接课程与资料检索。</p>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

withDefaults(
  defineProps<{
    variant?: 'default' | 'dashboard'
  }>(),
  {
    variant: 'default',
  },
)

const router = useRouter()
const auth = useAuth()
const menuOpen = ref(false)
const activePanel = ref<'messages' | 'notifications' | 'search' | null>(null)
const userInitial = computed(() => auth.user.value?.username?.slice(0, 1).toUpperCase() || 'U')

async function goCourses() {
  menuOpen.value = false
  activePanel.value = null
  await router.push('/courses')
}

function toggleMenu() {
  activePanel.value = null
  menuOpen.value = !menuOpen.value
}

function togglePanel(panel: 'messages' | 'notifications' | 'search') {
  menuOpen.value = false
  activePanel.value = activePanel.value === panel ? null : panel
}

async function openCreateCourse() {
  menuOpen.value = false
  activePanel.value = null
  await router.push('/courses?create=1')
}

async function handleLogout() {
  menuOpen.value = false
  activePanel.value = null
  if (auth.token.value) {
    await logout(auth.token.value)
  }
  auth.clear()
  await router.push('/login')
}
</script>
