<template>
  <article class="agent-card">
    <div class="section-head section-head-top">
      <div>
        <p class="eyebrow">Chat</p>
        <h3>课程对话</h3>
      </div>
      <div class="inline-actions">
        <button class="button ghost compact" @click="reloadAll" :disabled="loadingOverview || sendingQuestion">
          {{ loadingOverview ? '刷新中' : '刷新' }}
        </button>
        <button class="button primary compact" @click="handleCreateConversation" :disabled="creatingConversation || sendingQuestion">
          {{ creatingConversation ? '创建中' : '新会话' }}
        </button>
      </div>
    </div>

    <p v-if="overviewError" class="error">{{ overviewError }}</p>

    <div v-if="agent" class="agent-main-panel">
        <AgentConversationList
          :conversations="conversations"
          :selected-conversation-id="selectedConversationId"
          :error-message="conversationError"
          @select="selectConversation"
        />

        <AgentChatPanel
          :detail="selectedConversationDetail"
          :agent-name="agent.agentName"
          :current-conversation-title="currentConversationTitle"
          :selected-conversation-id="selectedConversationId"
          :loading-conversation="loadingConversation"
          :sending-question="sendingQuestion"
          :error-message="chatError"
          :question="questionForm.question"
          :agent-enabled="agent.status === 'enabled'"
          @reload="reloadConversation"
          @submit-question="submitQuestion"
          @update:question="questionForm.question = $event"
        />
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import AgentChatPanel from '@/components/AgentChatPanel.vue'
import AgentConversationList from '@/components/AgentConversationList.vue'
import {
  createAgentConversation,
  getAgentConversation,
  getCourseAgent,
  listAgentConversations,
  streamCourseAgent,
} from '@/api/agent'
import type {
  AgentMessageSourceVO,
  AgentMessageVO,
  AgentConversationDetailVO,
  AgentConversationVO,
  CourseAgentVO,
} from '@/types/agent'

const props = defineProps<{
  courseId: number
  token: string
  canManage: boolean
}>()

const agent = ref<CourseAgentVO | null>(null)
const conversations = ref<AgentConversationVO[]>([])
const selectedConversationId = ref<number | null>(null)
const selectedConversationDetail = ref<AgentConversationDetailVO | null>(null)
const loadingOverview = ref(false)
const loadingConversation = ref(false)
const creatingConversation = ref(false)
const sendingQuestion = ref(false)
const overviewError = ref('')
const conversationError = ref('')
const chatError = ref('')
const questionForm = reactive({
  question: '',
})

const currentConversationTitle = computed(() => selectedConversationDetail.value?.conversation.conversationTitle || '未选择会话')

watch(
  () => props.courseId,
  async () => {
    resetState()
    await loadOverview()
  },
)

onMounted(async () => {
  await loadOverview()
})

function resetState() {
  agent.value = null
  conversations.value = []
  selectedConversationId.value = null
  selectedConversationDetail.value = null
  overviewError.value = ''
  conversationError.value = ''
  chatError.value = ''
  questionForm.question = ''
}

async function loadOverview() {
  if (!props.token || !props.courseId) return
  loadingOverview.value = true
  overviewError.value = ''
  try {
    const [agentData, conversationData] = await Promise.all([
      getCourseAgent(props.token, props.courseId),
      listAgentConversations(props.token, props.courseId),
    ])
    agent.value = agentData
    conversations.value = conversationData

    if (!conversationData.length) {
      selectedConversationId.value = null
      selectedConversationDetail.value = null
      return
    }

    const nextConversationId = selectedConversationId.value && conversationData.some((item) => item.id === selectedConversationId.value)
      ? selectedConversationId.value
      : conversationData[0].id
    await selectConversation(nextConversationId)
  } catch (error) {
    overviewError.value = error instanceof Error ? error.message : 'Agent 信息加载失败'
  } finally {
    loadingOverview.value = false
  }
}

async function reloadAll() {
  await loadOverview()
}

async function selectConversation(conversationId: number) {
  selectedConversationId.value = conversationId
  chatError.value = ''
  loadingConversation.value = true
  try {
    selectedConversationDetail.value = await getAgentConversation(props.token, props.courseId, conversationId)
  } catch (error) {
    chatError.value = error instanceof Error ? error.message : '会话加载失败'
    selectedConversationDetail.value = null
  } finally {
    loadingConversation.value = false
  }
}

async function reloadConversation() {
  if (!selectedConversationId.value) return
  await selectConversation(selectedConversationId.value)
}

async function handleCreateConversation() {
  creatingConversation.value = true
  conversationError.value = ''
  try {
    const created = await createAgentConversation(props.token, props.courseId, { title: '' })
    conversations.value = [created, ...conversations.value]
    await selectConversation(created.id)
  } catch (error) {
    conversationError.value = error instanceof Error ? error.message : '会话创建失败'
  } finally {
    creatingConversation.value = false
  }
}

async function ensureConversationId() {
  if (selectedConversationId.value) {
    return selectedConversationId.value
  }
  const created = await createAgentConversation(props.token, props.courseId, { title: '' })
  conversations.value = [created, ...conversations.value]
  selectedConversationId.value = created.id
  return created.id
}

async function submitQuestion() {
  const question = questionForm.question.trim()
  if (!question) return
  sendingQuestion.value = true
  chatError.value = ''
  const rollback = snapshotMessages()
  try {
    const conversationId = await ensureConversationId()
    const { agentMessage } = appendLocalMessages(conversationId, question)
    questionForm.question = ''
    await streamCourseAgent(props.token, props.courseId, {
      conversationId,
      question,
    }, {
      onDelta: ({ content }) => {
        agentMessage.messageContent += content
      },
      onComplete: ({ answer, sources, tokenUsage }) => {
        agentMessage.messageContent = answer || agentMessage.messageContent
        agentMessage.sources = sources
        agentMessage.tokenUsage = tokenUsage
      },
    })
    await Promise.all([reloadConversation(), reloadConversations()])
  } catch (error) {
    rollback()
    chatError.value = error instanceof Error ? error.message : '提问失败'
  } finally {
    sendingQuestion.value = false
  }
}

async function reloadConversations() {
  try {
    conversations.value = await listAgentConversations(props.token, props.courseId)
  } catch (error) {
    conversationError.value = error instanceof Error ? error.message : '会话列表刷新失败'
  }
}

function ensureConversationDetail(conversationId: number) {
  if (selectedConversationDetail.value?.conversation.id === conversationId) {
    return selectedConversationDetail.value
  }
  const baseConversation = conversations.value.find((item) => item.id === conversationId) ?? {
    id: conversationId,
    courseId: props.courseId,
    agentId: agent.value?.id ?? 0,
    userId: 0,
    conversationTitle: '新会话',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }
  selectedConversationDetail.value = {
    conversation: { ...baseConversation },
    messages: [],
  }
  selectedConversationId.value = conversationId
  return selectedConversationDetail.value
}

function appendLocalMessages(conversationId: number, question: string) {
  const detail = ensureConversationDetail(conversationId)
  const timestamp = new Date().toISOString()
  const seed = Date.now()
  const userMessage: AgentMessageVO = {
    id: seed,
    conversationId,
    senderType: 'user',
    messageContent: question,
    tokenUsage: 0,
    createdAt: timestamp,
  }
  const agentMessage: AgentMessageVO = {
    id: seed + 1,
    conversationId,
    senderType: 'agent',
    messageContent: '',
    tokenUsage: 0,
    createdAt: timestamp,
    sources: [] as AgentMessageSourceVO[],
  }
  detail.messages = [...detail.messages, userMessage, agentMessage]
  return { agentMessage }
}

function snapshotMessages() {
  const previousDetail = selectedConversationDetail.value
    ? {
        conversation: { ...selectedConversationDetail.value.conversation },
        messages: selectedConversationDetail.value.messages.map((message) => ({
          ...message,
          sources: message.sources?.map((source) => ({ ...source })),
        })),
      }
    : null
  return () => {
    selectedConversationDetail.value = previousDetail
  }
}
</script>
