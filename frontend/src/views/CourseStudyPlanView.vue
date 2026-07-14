<template>
  <section class="page-section study-plan-page">
    <div class="study-plan-actions">
      <button class="button primary compact" type="button" @click="openCreateDialog">创建计划</button>
    </div>

    <article class="workspace-panel study-plan-index-panel">
      <p v-if="loadingPlans" class="muted-copy">正在加载计划...</p>
      <p v-else-if="loadError" class="error">{{ loadError }}</p>
      <div v-else-if="!planList.length" class="empty-state small">还没有计划，点击右上角创建。</div>

      <div v-else class="study-plan-card-grid">
        <RouterLink
          v-for="plan in planList"
          :key="plan.id"
          class="study-plan-card"
          :to="{ name: 'course-study-plan-detail', params: { courseId: course?.id, planId: plan.id } }"
        >
          <span class="study-plan-card-title">{{ plan.goal }}</span>
          <span class="study-plan-card-summary">{{ plan.generatedSummary || '已生成学习步骤，进入后查看执行安排。' }}</span>
          <span class="study-plan-card-meta">
            <span>截止 {{ formatDate(plan.deadlineDate) }}</span>
            <span>{{ plan.dailyMinutes }} 分钟/天</span>
          </span>
          <span class="study-plan-card-footer">
            <span>{{ plan.doneItemCount }}/{{ plan.itemCount }} 步完成</span>
            <strong>查看步骤</strong>
          </span>
        </RouterLink>
      </div>
    </article>

    <div v-if="createDialogOpen" class="modal-backdrop" @click.self="closeCreateDialog">
      <section class="modal-card card study-plan-create-modal">
        <div class="section-head">
          <div>
            <h2>创建学习计划</h2>
          </div>
          <button class="button ghost compact" type="button" @click="closeCreateDialog">关闭</button>
        </div>

        <form class="form" @submit.prevent="handleGeneratePlan">
          <label class="field">
            <span>学习目标</span>
            <textarea
              v-model.trim="planForm.goal"
              rows="4"
              placeholder="例如：两周内完成第一到第三章，并掌握核心概念与重点题型"
            />
          </label>

          <div class="study-plan-form-grid">
            <label class="field">
              <span>截止时间</span>
              <input v-model="planForm.deadlineDate" type="date" />
            </label>

            <label class="field">
              <span>每日可用时间（分钟）</span>
              <input v-model.number="planForm.dailyMinutes" type="number" min="15" max="480" step="15" />
            </label>
          </div>

          <p v-if="actionError" class="error">{{ actionError }}</p>

          <div class="inline-actions">
            <button class="button primary" type="submit" :disabled="generatingPlan">
              {{ generatingPlan ? '生成中...' : '生成计划' }}
            </button>
            <button class="button ghost" type="button" :disabled="generatingPlan" @click="closeCreateDialog">取消</button>
          </div>
        </form>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { generateStudyPlan, listStudyPlans } from '@/api/studyPlan'
import { useCourseContext } from '@/composables/useCourseContext'
import type { StudyPlanSummaryVO } from '@/types/studyPlan'

const router = useRouter()
const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const planList = ref<StudyPlanSummaryVO[]>([])
const loadingPlans = ref(false)
const generatingPlan = ref(false)
const createDialogOpen = ref(false)
const loadError = ref('')
const actionError = ref('')
const planForm = reactive({
  goal: '',
  deadlineDate: defaultDeadlineDate(),
  dailyMinutes: 60,
})

watch([course, token], () => {
  if (course.value && token.value) {
    void loadPlans()
  }
}, { immediate: true })

async function loadPlans() {
  if (!course.value || !token.value) return
  loadingPlans.value = true
  loadError.value = ''
  try {
    planList.value = await listStudyPlans(token.value, course.value.id)
  } catch (error) {
    loadError.value = error instanceof Error ? error.message : '学习计划加载失败'
  } finally {
    loadingPlans.value = false
  }
}

async function handleGeneratePlan() {
  if (!course.value || !token.value) return
  generatingPlan.value = true
  actionError.value = ''
  try {
    const plan = await generateStudyPlan(token.value, course.value.id, {
      goal: planForm.goal,
      deadlineDate: planForm.deadlineDate,
      dailyMinutes: planForm.dailyMinutes,
    })
    createDialogOpen.value = false
    resetPlanForm()
    await router.push({
      name: 'course-study-plan-detail',
      params: { courseId: course.value.id, planId: plan.id },
    })
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : '学习计划生成失败'
  } finally {
    generatingPlan.value = false
  }
}

function openCreateDialog() {
  actionError.value = ''
  createDialogOpen.value = true
}

function closeCreateDialog() {
  if (generatingPlan.value) return
  createDialogOpen.value = false
  actionError.value = ''
}

function resetPlanForm() {
  planForm.goal = ''
  planForm.deadlineDate = defaultDeadlineDate()
  planForm.dailyMinutes = 60
}

function defaultDeadlineDate() {
  const date = new Date()
  date.setDate(date.getDate() + 14)
  return date.toISOString().slice(0, 10)
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString('zh-CN')
}
</script>
