<template>
  <section class="page-section">
    <div class="course-page-head course-page-head-compact">
      <div>
        <p class="eyebrow">Agent</p>
        <h1>课程对话</h1>
      </div>
    </div>

    <div class="agent-page-layout agent-page-layout-single">
      <CourseAgentPanel
        v-if="course && token"
        :course-id="course.id"
        :token="token"
        :can-manage="canManage"
      />
    </div>
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
