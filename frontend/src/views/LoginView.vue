<template>
  <main class="auth-page">
    <section class="auth-split-shell">
      <article class="auth-brand-panel">
        <div>
          <p class="eyebrow">课程智学平台</p>
          <h1>管理课程、资料与智能助手</h1>
          <p class="lead">把课程详情、成员、资料中心和课程 Agent 收敛到一个统一的教师工作台。</p>
        </div>

        <div class="auth-illustration-card">
          <div class="auth-illustration-orbit auth-orbit-a" />
          <div class="auth-illustration-orbit auth-orbit-b" />
          <div class="auth-illustration-device">
            <div class="auth-device-screen">
              <div class="auth-device-card" />
              <div class="auth-device-card short" />
            </div>
          </div>
        </div>
      </article>

      <section class="auth-form-panel">
        <div class="auth-panel-head">
          <div>
            <p class="eyebrow">{{ mode === 'login' ? 'Login' : 'Register' }}</p>
            <h2>{{ mode === 'login' ? '登录课程平台' : '注册新账号' }}</h2>
          </div>

          <button class="button ghost compact" type="button" @click="toggleMode">
            {{ mode === 'login' ? '去注册' : '去登录' }}
          </button>
        </div>

        <form class="form auth-form" @submit.prevent="handleSubmit">
          <label class="field">
            <span>用户名</span>
            <input v-model.trim="form.username" type="text" autocomplete="username" />
          </label>

          <label class="field">
            <span>密码</span>
            <input v-model="form.password" type="password" autocomplete="current-password" />
          </label>

          <label v-if="mode === 'register'" class="field">
            <span>手机号</span>
            <input v-model.trim="form.phone" type="tel" autocomplete="tel" />
          </label>

          <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

          <button class="button primary auth-submit" type="submit" :disabled="loading">
            {{ loading ? '处理中...' : mode === 'login' ? '登录' : '注册' }}
          </button>
        </form>

        <p class="auth-switch-copy">
          {{ mode === 'login' ? '还没有账号？切换到注册后即可创建。' : '已经有账号？切换到登录继续使用。' }}
        </p>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { login, register } from '@/api/user'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const auth = useAuth()
const mode = ref<'login' | 'register'>('login')
const loading = ref(false)
const errorMessage = ref('')
const form = reactive({
  username: '',
  password: '',
  phone: '',
})

function toggleMode() {
  mode.value = mode.value === 'login' ? 'register' : 'login'
  errorMessage.value = ''
}

async function handleSubmit() {
  loading.value = true
  errorMessage.value = ''
  try {
    if (mode.value === 'login') {
      const result = await login({ username: form.username, password: form.password })
      auth.persist(result)
      await router.push('/')
      return
    }

    await register({ username: form.username, password: form.password, phone: form.phone })
    mode.value = 'login'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '操作失败'
  } finally {
    loading.value = false
  }
}
</script>
