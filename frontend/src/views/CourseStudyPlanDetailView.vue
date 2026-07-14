<template>
  <section class="page-section study-plan-page">
    <div class="course-page-head course-page-head-compact study-plan-page-head">
      <div class="course-overview-title-row">
        <p class="eyebrow">Steps</p>
        <div class="course-overview-heading">
          <h1>计划步骤</h1>
          <span class="course-status-inline">执行并跟踪当前计划</span>
        </div>
      </div>

      <RouterLink class="button ghost compact" :to="{ name: 'course-study-plan', params: { courseId } }">返回计划</RouterLink>
    </div>

    <article class="workspace-panel study-plan-step-workspace">
      <p v-if="loadingDetail" class="muted-copy">正在加载详情...</p>
      <p v-else-if="detailError" class="error">{{ detailError }}</p>
      <div v-else-if="!selectedPlan" class="empty-state small">计划不存在或暂无权限查看。</div>

      <template v-else>
        <div class="study-plan-detail-head">
          <div>
            <p class="eyebrow">Plan</p>
            <h2>{{ selectedPlan.goal }}</h2>
            <p class="study-plan-summary-copy">{{ selectedPlan.generatedSummary }}</p>
          </div>
          <div class="study-plan-detail-meta">
            <span>截止 {{ formatDate(selectedPlan.deadlineDate) }}</span>
            <span>{{ selectedPlan.dailyMinutes }} 分钟/天</span>
            <span>{{ doneCount }}/{{ selectedPlan.items.length }} 步</span>
          </div>
        </div>

        <div class="study-plan-item-list">
          <article
            v-for="item in selectedPlan.items"
            :key="item.id"
            class="study-plan-item"
            :class="{ done: item.status === 'done' }"
          >
            <div class="study-plan-item-head">
              <div>
                <p class="study-plan-item-date">第 {{ item.dayIndex }} 天 · {{ formatDate(item.planDate) }}</p>
                <h3>{{ item.title }}</h3>
              </div>
              <button
                class="button compact"
                :class="item.status === 'done' ? 'ghost' : 'primary'"
                :disabled="updatingItemId === item.id"
                type="button"
                @click="toggleItemStatus(item.id, item.status)"
              >
                {{ updatingItemId === item.id ? '提交中' : item.status === 'done' ? '已完成' : '标记完成' }}
              </button>
            </div>

            <p class="study-plan-item-copy">{{ item.tasksText }}</p>

            <div class="study-plan-item-meta">
              <span>{{ item.suggestedMinutes }} 分钟</span>
              <span v-if="item.materialNodeIds.length">关联资料 {{ item.materialNodeIds.join(' / ') }}</span>
            </div>
          </article>
        </div>
      </template>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { getStudyPlan, updateStudyPlanItemStatus } from '@/api/studyPlan'
import { useCourseContext } from '@/composables/useCourseContext'
import type { StudyPlanItemStatus, StudyPlanVO } from '@/types/studyPlan'

const route = useRoute()
const context = useCourseContext()
const course = computed(() => context.course.value)
const courseId = computed(() => course.value?.id ?? Number(route.params.courseId))
const token = computed(() => context.token.value)
const selectedPlan = ref<StudyPlanVO | null>(null)
const loadingDetail = ref(false)
const updatingItemId = ref<number | null>(null)
const detailError = ref('')

const planId = computed(() => Number(route.params.planId))
const doneCount = computed(() => selectedPlan.value?.items.filter((item) => item.status === 'done').length ?? 0)

watch([courseId, token, planId], () => {
  if (courseId.value && token.value && planId.value) {
    void loadPlanDetail()
  }
}, { immediate: true })

async function loadPlanDetail() {
  if (!courseId.value || !token.value || !planId.value) return
  loadingDetail.value = true
  detailError.value = ''
  try {
    selectedPlan.value = await getStudyPlan(token.value, courseId.value, planId.value)
  } catch (error) {
    selectedPlan.value = null
    detailError.value = error instanceof Error ? error.message : '学习计划详情加载失败'
  } finally {
    loadingDetail.value = false
  }
}

async function toggleItemStatus(itemId: number, status: StudyPlanItemStatus) {
  if (!courseId.value || !token.value || !selectedPlan.value) return
  updatingItemId.value = itemId
  detailError.value = ''
  try {
    selectedPlan.value = await updateStudyPlanItemStatus(
      token.value,
      courseId.value,
      selectedPlan.value.id,
      itemId,
      status === 'done' ? 'pending' : 'done',
    )
  } catch (error) {
    detailError.value = error instanceof Error ? error.message : '任务状态更新失败'
  } finally {
    updatingItemId.value = null
  }
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString('zh-CN')
}
</script>
