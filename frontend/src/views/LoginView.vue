<template>
  <main class="page-shell auth-shell">
    <section class="card auth-card">
      <div class="auth-header">
        <p class="eyebrow">Course Agent</p>
        <h1>{{ mode === 'login' ? '登录' : '注册' }}</h1>
        <p class="lead">{{ mode === 'login' ? '使用账号登录课程平台。' : '创建你的课程平台账号。' }}</p>
      </div>

      <form class="form" @submit.prevent="handleSubmit">
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

        <button class="button primary" type="submit" :disabled="loading">
          {{ loading ? '处理中...' : mode === 'login' ? '登录' : '注册' }}
        </button>
      </form>

      <button class="button ghost switcher" type="button" @click="toggleMode">
        {{ mode === 'login' ? '没有账号，去注册' : '已有账号，去登录' }}
      </button>
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
