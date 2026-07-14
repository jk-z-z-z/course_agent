<template>
  <section class="agent-chat-panel">
    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="agent-chat-shell">
      <div ref="messagesRef" class="agent-messages">
        <div v-if="loadingConversation" class="agent-chat-empty">
          <p>加载会话中...</p>
        </div>

        <div v-else-if="!detail || !detail.messages.length" class="agent-chat-empty">
          <h2>开始课程对话</h2>
          <p>可以直接询问课程资料、作业要求或知识点。</p>
        </div>

        <template v-else>
          <article
            v-for="message in detail.messages"
            :key="message.id"
            class="agent-message"
            :class="message.senderType === 'user' ? 'user' : 'agent'"
          >
            <div v-if="message.senderType === 'agent'" class="agent-message-avatar">{{ agentName.slice(0, 1) }}</div>
            <p class="agent-message-content">{{ message.messageContent }}</p>
            <div v-if="retrievedMaterialList(message).length" class="agent-source-list">
              <span class="agent-source-title">检索资料</span>
              <div
                v-for="source in retrievedMaterialList(message)"
                :key="`${message.id}-${source.materialNodeId}-${source.fileName}`"
                class="agent-source-item"
              >
                <strong>{{ source.fileName || `资料 ${source.materialNodeId}` }}</strong>
                <p>{{ source.snippetText }}</p>
              </div>
            </div>
          </article>
        </template>
      </div>

      <form class="agent-ask-form" @submit.prevent="$emit('submit-question')">
        <label class="agent-input-field">
          <textarea
            :value="question"
            placeholder="给课程 Agent 发送消息"
            rows="1"
            :disabled="disabled"
            @input="$emit('update:question', ($event.target as HTMLTextAreaElement).value)"
            @keydown.enter.exact.prevent="$emit('submit-question')"
          />
          <button class="agent-send-button" type="submit" :disabled="disabled" :title="disabledReason">
            <span aria-hidden="true">↑</span>
            <span class="sr-only">{{ sendingQuestion ? '生成中' : '发送' }}</span>
          </button>
        </label>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { AgentConversationDetailVO, AgentMessageVO } from '@/types/agent'

const props = defineProps<{
  detail: AgentConversationDetailVO | null
  agentName: string
  loadingConversation: boolean
  sendingQuestion: boolean
  errorMessage: string
  question: string
  agentEnabled: boolean
}>()

defineEmits<{
  'submit-question': []
  'update:question': [value: string]
}>()

const messagesRef = ref<HTMLElement | null>(null)
const disabled = computed(() => props.sendingQuestion || !props.agentEnabled)
const disabledReason = computed(() => {
  if (props.sendingQuestion) return '生成中'
  if (!props.agentEnabled) return 'Agent 已停用'
  return '发送'
})

function retrievedMaterialList(message: AgentMessageVO) {
  return message.retrievedMaterials?.length ? message.retrievedMaterials : (message.sources ?? [])
}

watch(
  () => props.detail?.messages.map((message) => message.messageContent).join('\n') ?? '',
  async () => {
    await nextTick()
    if (!messagesRef.value) return
    messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  },
)
</script>
