<template>
  <section class="panel agent-chat-panel">
    <div class="section-head section-head-top compact-head">
      <div>
        <p class="label">当前会话</p>
        <h3 class="agent-chat-title">{{ currentConversationTitle }}</h3>
      </div>
      <button
        v-if="selectedConversationId"
        class="button ghost compact"
        @click="$emit('reload')"
        :disabled="loadingConversation || sendingQuestion"
      >
        {{ loadingConversation ? '加载中' : '刷新会话' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="detail" class="agent-chat-shell">
      <div class="agent-messages">
        <article
          v-for="message in detail.messages"
          :key="message.id"
          class="agent-message"
          :class="message.senderType === 'user' ? 'user' : 'agent'"
        >
          <div class="agent-message-head">
            <span>{{ message.senderType === 'user' ? '我' : agentName }}</span>
            <span class="agent-message-time">{{ formatDateTime(message.createdAt) }}</span>
          </div>
          <p class="agent-message-content">{{ message.messageContent }}</p>
          <div v-if="message.sources?.length" class="agent-source-list">
            <div v-for="source in message.sources" :key="`${message.id}-${source.materialNodeId}-${source.fileName}`" class="agent-source-item">
              <strong>{{ source.fileName || `资料 ${source.materialNodeId}` }}</strong>
              <p>{{ source.snippetText }}</p>
            </div>
          </div>
        </article>
      </div>

      <form class="agent-ask-form" @submit.prevent="$emit('submit-question')">
        <label class="field">
          <span>向课程助教提问</span>
          <textarea
            :value="question"
            placeholder="例如：帮我总结本课程资料中对课程项目的要求"
            @input="$emit('update:question', ($event.target as HTMLTextAreaElement).value)"
          />
        </label>
        <button class="button primary" type="submit" :disabled="disabled">
          {{ sendingQuestion ? '流式发送中' : disabledReason }}
        </button>
      </form>
    </div>

    <div v-else class="empty-state small">
      <p class="lead">先创建或选择一个会话，然后开始提问。</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AgentConversationDetailVO } from '@/types/agent'
import { formatDateTime } from '@/utils/date'

const props = defineProps<{
  detail: AgentConversationDetailVO | null
  agentName: string
  currentConversationTitle: string
  selectedConversationId: number | null
  loadingConversation: boolean
  sendingQuestion: boolean
  errorMessage: string
  question: string
  agentEnabled: boolean
}>()

defineEmits<{
  reload: []
  'submit-question': []
  'update:question': [value: string]
}>()

const disabled = computed(() => props.sendingQuestion || !props.agentEnabled)
const disabledReason = computed(() => {
  if (props.sendingQuestion) return '流式发送中'
  if (!props.agentEnabled) return 'Agent 已停用'
  return '发送问题'
})
</script>
