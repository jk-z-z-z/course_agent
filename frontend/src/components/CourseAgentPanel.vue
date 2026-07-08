<template>
  <article class="card agent-card">
    <div class="section-head section-head-top">
      <div>
        <p class="eyebrow">Agent</p>
        <h3>课程助教</h3>
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

    <div v-if="agent" class="agent-grid">
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

          <form class="form agent-config-form" @submit.prevent="submitConfig">
            <label class="field">
              <span>Agent 名称</span>
              <input v-model.trim="configForm.agentName" type="text" />
            </label>

            <label class="field">
              <span>状态</span>
              <select v-model="configForm.status">
                <option value="enabled">启用</option>
                <option value="disabled">停用</option>
              </select>
            </label>

            <label class="field">
              <span>提示词</span>
              <textarea v-model.trim="configForm.promptTemplate"></textarea>
            </label>

            <p v-if="configError" class="error">{{ configError }}</p>
            <button class="button ghost compact" type="submit" :disabled="savingConfig">
              {{ savingConfig ? '保存中' : '保存配置' }}
            </button>
          </form>
        </div>

        <div class="panel agent-conversation-panel">
          <div class="section-head compact-head">
            <div>
              <p class="label">会话列表</p>
              <p class="muted-copy">学生仅可见自己的会话。</p>
            </div>
          </div>

          <p v-if="conversationError" class="error">{{ conversationError }}</p>
          <p v-else-if="!conversations.length" class="muted-copy">暂无会话，先新建一个。</p>

          <div v-else class="agent-conversation-list">
            <button
              v-for="conversation in conversations"
              :key="conversation.id"
              class="agent-conversation-item"
              :class="{ active: conversation.id === selectedConversationId }"
              @click="selectConversation(conversation.id)"
            >
              <span class="agent-conversation-title">{{ conversation.conversationTitle || '未命名会话' }}</span>
              <span class="agent-conversation-meta">{{ formatDateTime(conversation.updatedAt) }}</span>
            </button>
          </div>
        </div>
      </section>

      <section class="panel agent-chat-panel">
        <div class="section-head section-head-top compact-head">
          <div>
            <p class="label">当前会话</p>
            <h3 class="agent-chat-title">{{ currentConversationTitle }}</h3>
          </div>
          <button
            v-if="selectedConversationId"
            class="button ghost compact"
            @click="reloadConversation"
            :disabled="loadingConversation || sendingQuestion"
          >
            {{ loadingConversation ? '加载中' : '刷新会话' }}
          </button>
        </div>

        <p v-if="chatError" class="error">{{ chatError }}</p>

        <div v-if="selectedConversationDetail" class="agent-chat-shell">
          <div class="agent-messages">
            <article
              v-for="message in selectedConversationDetail.messages"
              :key="message.id"
              class="agent-message"
              :class="message.senderType === 'user' ? 'user' : 'agent'"
            >
              <div class="agent-message-head">
                <span>{{ message.senderType === 'user' ? '我' : agent?.agentName || '课程助教' }}</span>
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

          <form class="agent-ask-form" @submit.prevent="submitQuestion">
            <label class="field">
              <span>向课程助教提问</span>
              <textarea v-model.trim="questionForm.question" placeholder="例如：帮我总结本课程资料中对课程项目的要求"></textarea>
            </label>
            <button class="button primary" type="submit" :disabled="sendingQuestion || !agent || agent.status === 'disabled'">
              {{ sendingQuestion ? '发送中' : agent?.status === 'disabled' ? 'Agent 已停用' : '发送问题' }}
            </button>
          </form>
        </div>

        <div v-else class="empty-state small">
          <p class="lead">先创建或选择一个会话，然后开始提问。</p>
        </div>
      </section>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  askCourseAgent,
  createAgentConversation,
  getAgentConversation,
  getCourseAgent,
  listAgentConversations,
  updateCourseAgent,
} from '@/api/agent'
import type {
  AgentConversationDetailVO,
  AgentConversationVO,
  AgentStatus,
  CourseAgentVO,
} from '@/types/agent'
import { formatDateTime } from '@/utils/date'

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
const savingConfig = ref(false)
const sendingQuestion = ref(false)
const overviewError = ref('')
const conversationError = ref('')
const configError = ref('')
const chatError = ref('')
const configForm = reactive({
  agentName: '',
  promptTemplate: '',
  status: 'enabled' as AgentStatus,
})
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
  configError.value = ''
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
    configForm.agentName = agentData.agentName
    configForm.promptTemplate = agentData.promptTemplate ?? ''
    configForm.status = agentData.status
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

async function submitConfig() {
  if (!agent.value) return
  savingConfig.value = true
  configError.value = ''
  try {
    const updated = await updateCourseAgent(props.token, props.courseId, {
      agentName: configForm.agentName,
      promptTemplate: configForm.promptTemplate,
      status: configForm.status,
      retrievalScope: 'course_all',
    })
    agent.value = updated
    configForm.agentName = updated.agentName
    configForm.promptTemplate = updated.promptTemplate ?? ''
    configForm.status = updated.status
  } catch (error) {
    configError.value = error instanceof Error ? error.message : '配置保存失败'
  } finally {
    savingConfig.value = false
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
  try {
    const conversationId = await ensureConversationId()
    await askCourseAgent(props.token, props.courseId, {
      conversationId,
      question,
    })
    questionForm.question = ''
    await Promise.all([reloadConversation(), reloadConversations()])
  } catch (error) {
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
</script>
