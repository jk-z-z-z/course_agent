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
      <RouterLink to="/courses" class="platform-link">主页</RouterLink>
      <RouterLink v-if="courseDetailLink" :to="courseDetailLink" class="platform-link">课程详情</RouterLink>
    </nav>

    <div class="platform-top-actions" :class="{ compact: props.variant === 'dashboard' }">
      <button
        v-if="props.variant !== 'dashboard'"
        class="platform-icon-button"
        type="button"
        aria-label="搜索"
        @click="togglePanel('search')"
      >
        <span>⌕</span>
      </button>
      <button
        v-if="props.variant !== 'dashboard'"
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
          :class="{ dashboard: props.variant === 'dashboard' }"
          type="button"
          @click="handleUserTrigger"
        >
          <span class="platform-avatar">{{ userInitial }}</span>
          <span class="platform-user-copy">
            <strong>{{ props.variant === 'dashboard' ? '退出登录' : auth.user.value?.username || '未登录用户' }}</strong>
            <small>{{ props.variant === 'dashboard' ? auth.user.value?.username || '未登录用户' : '教师工作台' }}</small>
          </span>
          <span class="platform-caret">⌄</span>
        </button>

        <div v-if="menuOpen && props.variant !== 'dashboard'" class="platform-user-dropdown">
          <button class="platform-dropdown-item" type="button" @click="goCourses">回到课程列表</button>
          <button class="platform-dropdown-item danger" type="button" @click="handleLogout">退出登录</button>
        </div>
      </div>
    </div>

    <div v-if="activePanel" class="platform-floating-panel">
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
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

const props = withDefaults(
  defineProps<{
    variant?: 'default' | 'dashboard'
  }>(),
  {
    variant: 'default',
  },
)

const router = useRouter()
const route = useRoute()
const auth = useAuth()
const menuOpen = ref(false)
const activePanel = ref<'messages' | 'notifications' | 'search' | null>(null)
const userInitial = computed(() => auth.user.value?.username?.slice(0, 1).toUpperCase() || 'U')
const courseDetailLink = computed(() => {
  const courseId = route.params.courseId
  if (!courseId) return ''
  const normalizedCourseId = Array.isArray(courseId) ? courseId[0] : courseId
  return normalizedCourseId ? `/courses/${normalizedCourseId}/overview` : ''
})

async function goCourses() {
  menuOpen.value = false
  activePanel.value = null
  await router.push('/courses')
}

function toggleMenu() {
  activePanel.value = null
  menuOpen.value = !menuOpen.value
}

async function handleUserTrigger() {
  if (props.variant === 'dashboard') {
    await handleLogout()
    return
  }
  toggleMenu()
}

function togglePanel(panel: 'messages' | 'notifications' | 'search') {
  menuOpen.value = false
  activePanel.value = activePanel.value === panel ? null : panel
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
