<template>
  <section class="page-section">
    <div class="course-page-head course-page-head-compact">
      <div class="course-overview-title-row">
        <p class="eyebrow">Home</p>
        <div class="course-overview-heading">
          <h1>{{ course?.courseName || '课程详情' }}</h1>
          <span v-if="course" class="pill subtle">{{ statusLabel(course.status) }}</span>
        </div>
      </div>
    </div>

    <div class="overview-grid" v-if="course">
      <article class="workspace-panel overview-main-card">
        <div class="overview-summary">
          <div class="overview-copy">
            <p class="lead overview-lead">{{ course.courseDescription || '当前课程还没有填写课程简介。' }}</p>
          </div>
        </div>
      </article>

      <article class="workspace-panel overview-tabs-card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Information</p>
            <h2>基本信息</h2>
          </div>
        </div>

        <div class="overview-info-list">
          <div class="overview-info-item">
            <span class="label">课程编号</span>
            <strong>{{ course.courseCode }}</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">课程成员</span>
            <strong>{{ membersCount }} 人</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">创建时间</span>
            <strong>{{ formatDateTime(course.createdAt) }}</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">更新时间</span>
            <strong>{{ formatDateTime(course.updatedAt) }}</strong>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listCourseMembers } from '@/api/course'
import { useCourseContext } from '@/composables/useCourseContext'
import type { CourseStatus } from '@/types/course'
import { formatDateTime } from '@/utils/date'

const context = useCourseContext()
const course = computed(() => context.course.value)
const membersCount = ref(0)

onMounted(loadMembersCount)

async function loadMembersCount() {
  if (!course.value || !context.token.value) return
  try {
    const members = await listCourseMembers(context.token.value, course.value.id)
    membersCount.value = members.length
  } catch {
    membersCount.value = 0
  }
}

function statusLabel(status: CourseStatus) {
  switch (status) {
    case 'active':
      return '进行中'
    case 'archived':
      return '已归档'
    case 'deleted':
      return '已删除'
    default:
      return status
  }
}
</script>
