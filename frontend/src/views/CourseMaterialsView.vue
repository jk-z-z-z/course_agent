<template>
  <section class="page-section">
    <div class="materials-page-layout materials-page-layout-single">
      <CourseMaterialsPanel
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
import CourseMaterialsPanel from '@/components/CourseMaterialsPanel.vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const canManage = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
</script>
