<template>
  <section class="agent-page">
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
