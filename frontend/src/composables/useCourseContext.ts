import { inject, provide, readonly, ref } from 'vue'
import type { Ref } from 'vue'
import type { CourseVO } from '@/types/course'

interface CourseContextValue {
  course: Readonly<Ref<CourseVO | null>>
  token: Readonly<Ref<string>>
  refreshCourse: () => Promise<void>
}

const courseContextKey = Symbol('course-context')

export function provideCourseContext(value: CourseContextValue) {
  provide(courseContextKey, value)
}

export function useCourseContext() {
  const context = inject<CourseContextValue | null>(courseContextKey, null)
  if (!context) {
    throw new Error('Course context is not available')
  }
  return context
}

export function createEmptyCourseRef() {
  return ref<CourseVO | null>(null)
}
