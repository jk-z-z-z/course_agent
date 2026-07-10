<template>
  <AppShell has-sidebar>
    <template #top>
      <PlatformTopNav />
    </template>

    <template #side>
      <CourseSidebar
        v-if="course"
        :course-id="course.id"
        :course-name="course.courseName"
      />
    </template>

    <div v-if="loading" class="page-state-card">
      <p class="eyebrow">Loading</p>
      <h1>正在加载课程</h1>
    </div>

    <div v-else-if="errorMessage" class="page-state-card">
      <p class="eyebrow">Unavailable</p>
      <h1>课程加载失败</h1>
      <p class="lead">{{ errorMessage }}</p>
      <button class="button primary" @click="refreshCourse">重新加载</button>
    </div>

    <RouterView v-else />
  </AppShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AppShell from '@/components/AppShell.vue'
import CourseSidebar from '@/components/CourseSidebar.vue'
import PlatformTopNav from '@/components/PlatformTopNav.vue'
import { useAuth } from '@/composables/useAuth'
import { provideCourseContext } from '@/composables/useCourseContext'
import { getCourse } from '@/api/course'
import type { CourseVO } from '@/types/course'

const route = useRoute()
const auth = useAuth()
const course = ref<CourseVO | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const token = computed(() => auth.token.value)

provideCourseContext({
  course: computed(() => course.value),
  token,
  refreshCourse,
})

async function refreshCourse() {
  const courseId = Number(route.params.courseId)
  if (!token.value || !courseId) return
  loading.value = true
  errorMessage.value = ''
  try {
    course.value = await getCourse(token.value, courseId)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '课程加载失败'
  } finally {
    loading.value = false
  }
}

watch(() => route.params.courseId, refreshCourse)

onMounted(refreshCourse)
</script>
