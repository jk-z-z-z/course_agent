<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Materials</p>
        <h1>课程资料</h1>
        <p class="lead">资料模块先接入现有组件，再逐步贴近你提供的资料页原型。</p>
      </div>
    </div>

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
