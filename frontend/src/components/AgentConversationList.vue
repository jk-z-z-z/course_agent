<template>
  <WorkspaceSidePanel
    title="会话"
    :collapsed="collapsed"
    panel-class="agent-conversation-panel"
    expand-label="展开会话列表"
    collapse-label="收起会话列表"
    @update:collapsed="$emit('toggle-collapsed')"
  >
    <template #actions>
      <button
        class="workspace-side-action"
        type="button"
        :disabled="creatingConversation || sendingQuestion"
        @click="$emit('create')"
      >
        {{ creatingConversation ? '创建中' : '新会话' }}
      </button>
    </template>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="!conversations.length" class="muted-copy">暂无会话</p>

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
  </WorkspaceSidePanel>
</template>

<script setup lang="ts">
import WorkspaceSidePanel from '@/components/WorkspaceSidePanel.vue'
import type { AgentConversationVO } from '@/types/agent'
import { formatDateTime } from '@/utils/date'

defineProps<{
  conversations: AgentConversationVO[]
  selectedConversationId: number | null
  errorMessage: string
  creatingConversation: boolean
  sendingQuestion: boolean
  collapsed: boolean
}>()

defineEmits<{
  select: [conversationId: number]
  create: []
  'toggle-collapsed': []
}>()
</script>
