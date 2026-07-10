<template>
  <div class="course-sidebar">
    <div class="course-sidebar-head">
      <p class="eyebrow">课程导航</p>
      <h2 class="course-sidebar-title">{{ courseName }}</h2>
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
  { to: `/courses/${props.courseId}/overview`, label: '课程详情', icon: '◧' },
  { to: `/courses/${props.courseId}/members`, label: '课程成员', icon: '◎' },
  { to: `/courses/${props.courseId}/materials`, label: '课程资料', icon: '▤' },
  { to: `/courses/${props.courseId}/agent`, label: '课程 Agent', icon: '◌' },
])
</script>
