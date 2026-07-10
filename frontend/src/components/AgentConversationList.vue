<template>
  <div class="panel agent-conversation-panel">
    <div class="section-head compact-head">
      <div>
        <p class="label">会话列表</p>
        <p class="muted-copy">学生仅可见自己的会话。</p>
      </div>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="!conversations.length" class="muted-copy">暂无会话，先新建一个。</p>

    <div v-else class="agent-conversation-list">
      <button
        v-for="conversation in conversations"
        :key="conversation.id"
        class="agent-conversation-item"
        :class="{ active: conversation.id === selectedConversationId }"
        @click="$emit('select', conversation.id)"
      >
        <span class="agent-conversation-title">{{ conversation.conversationTitle || '未命名会话' }}</span>
        <span class="agent-conversation-meta">{{ formatDateTime(conversation.updatedAt) }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AgentConversationVO } from '@/types/agent'
import { formatDateTime } from '@/utils/date'

defineProps<{
  conversations: AgentConversationVO[]
  selectedConversationId: number | null
  errorMessage: string
}>()

defineEmits<{
  select: [conversationId: number]
}>()
</script>
