<template>
  <div class="course-top-tabs">
    <div class="course-top-tabs-copy">
      <p class="eyebrow">Course Workspace</p>
      <h1>{{ courseName }}</h1>
      <p class="course-top-tabs-subtitle">{{ courseCode }} · {{ roleLabel }} · {{ statusLabel }}</p>
    </div>

    <nav class="course-top-tabs-nav">
      <RouterLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="course-top-tab-link"
        active-class="active"
      >
        {{ item.label }}
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { CourseRole, CourseStatus } from '@/types/course'

const props = defineProps<{
  courseId: number
  courseName: string
  courseCode: string
  role?: CourseRole
  status: CourseStatus
}>()

const items = computed(() => [
  { to: `/courses/${props.courseId}/overview`, label: '课程详情' },
  { to: `/courses/${props.courseId}/members`, label: '课程成员' },
  { to: `/courses/${props.courseId}/materials`, label: '课程资料' },
  { to: `/courses/${props.courseId}/agent`, label: '课程 Agent' },
])

const roleLabel = computed(() => {
  switch (props.role) {
    case 'owner':
      return '创建者'
    case 'teacher':
      return '教师'
    case 'student':
      return '学生'
    default:
      return '课程成员'
  }
})

const statusLabel = computed(() => {
  switch (props.status) {
    case 'active':
      return '进行中'
    case 'archived':
      return '已归档'
    case 'deleted':
      return '已删除'
    default:
      return props.status
  }
})
</script>
