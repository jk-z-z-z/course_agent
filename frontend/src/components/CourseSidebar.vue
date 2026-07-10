<template>
  <div class="course-sidebar">
    <div class="course-sidebar-head">
      <p class="eyebrow">课程导航</p>
      <h2 class="course-sidebar-title">课程工作区</h2>
      <p class="course-sidebar-copy">{{ courseName }}</p>
    </div>

    <nav class="course-sidebar-nav">
      <RouterLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="course-sidebar-link"
        active-class="active"
      >
        <span class="course-sidebar-icon">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps<{
  courseId: number
  courseName: string
}>()

const items = computed(() => [
  { to: `/courses/${props.courseId}/overview`, label: '首页', icon: '◧' },
  { to: `/courses/${props.courseId}/materials`, label: '资料页', icon: '▤' },
  { to: `/courses/${props.courseId}/agent`, label: 'Agent页', icon: '◌' },
  { to: `/courses/${props.courseId}/members`, label: '成员页', icon: '◎' },
])
</script>
