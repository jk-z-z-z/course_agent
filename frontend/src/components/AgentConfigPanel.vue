<template>
  <section class="agent-sidebar-panel">
    <div class="panel agent-summary-panel">
      <div class="profile-row">
        <div>
          <p class="label">Agent 名称</p>
          <p class="value">{{ agent.agentName }}</p>
        </div>
        <span class="pill" :class="agent.status === 'enabled' ? '' : 'subtle'">
          {{ agent.status === 'enabled' ? '已启用' : '已停用' }}
        </span>
      </div>
      <p class="muted-copy top-gap">范围：{{ agent.retrievalScope === 'course_all' ? '课程全部资料' : agent.retrievalScope }}</p>
      <p class="muted-copy">最近更新：{{ formatDateTime(agent.updatedAt) }}</p>
    </div>

    <div v-if="canManage" class="panel">
      <div class="section-head section-head-top compact-head">
        <div>
          <p class="label">配置</p>
          <p class="muted-copy">教师和创建者可调整基础行为。</p>
        </div>
      </div>

      <form class="form agent-config-form" @submit.prevent="$emit('submit')">
        <label class="field">
          <span>Agent 名称</span>
          <input :value="configForm.agentName" type="text" @input="$emit('update:agentName', ($event.target as HTMLInputElement).value)" />
        </label>

        <label class="field">
          <span>状态</span>
          <select :value="configForm.status" @change="$emit('update:status', ($event.target as HTMLSelectElement).value)">
            <option value="enabled">启用</option>
            <option value="disabled">停用</option>
          </select>
        </label>

        <label class="field">
          <span>提示词</span>
          <textarea :value="configForm.promptTemplate" @input="$emit('update:promptTemplate', ($event.target as HTMLTextAreaElement).value)" />
        </label>

        <p v-if="configError" class="error">{{ configError }}</p>
        <button class="button ghost compact" type="submit" :disabled="savingConfig">
          {{ savingConfig ? '保存中' : '保存配置' }}
        </button>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { AgentStatus, CourseAgentVO } from '@/types/agent'
import { formatDateTime } from '@/utils/date'

defineProps<{
  agent: CourseAgentVO
  canManage: boolean
  savingConfig: boolean
  configError: string
  configForm: {
    agentName: string
    promptTemplate: string
    status: AgentStatus
  }
}>()

defineEmits<{
  submit: []
  'update:agentName': [value: string]
  'update:promptTemplate': [value: string]
  'update:status': [value: string]
}>()
</script>
