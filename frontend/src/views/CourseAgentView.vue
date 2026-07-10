<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Agent</p>
        <h1>课程 Agent</h1>
        <p class="lead">Agent 配置与会话会放在独立页面，避免继续和课程详情混在一起。</p>
      </div>
    </div>

    <CourseAgentPanel
      v-if="course && token"
      :course-id="course.id"
      :token="token"
      :can-manage="canManage"
    />
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CourseAgentPanel from '@/components/CourseAgentPanel.vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const canManage = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
</script>
