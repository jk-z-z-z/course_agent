<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Overview</p>
        <h1>{{ course?.courseName || '课程详情' }}</h1>
        <p class="lead">{{ course?.courseDescription || '这里将展示课程摘要、状态与快捷操作。' }}</p>
      </div>
    </div>

    <article class="content-card">
      <h2>课程详情页骨架</h2>
      <p class="muted-copy">下一步迁入课程信息卡、编辑课程和删除课程功能。</p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
</script>
