<template>
  <section class="materials-page">
    <CourseMaterialsPanel
      v-if="course && token"
      :course-id="course.id"
      :token="token"
      :can-manage="canManage"
    />
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CourseMaterialsPanel from '@/components/CourseMaterialsPanel.vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const canManage = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
</script>
