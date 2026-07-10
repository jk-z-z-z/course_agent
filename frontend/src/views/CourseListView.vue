<template>
  <section class="page-section course-dashboard">
    <div class="course-list-hero">
      <div class="course-list-hero-copy">
        <p class="eyebrow">Welcome</p>
        <h1>欢迎回来，{{ auth.user.value?.username || '同学' }}</h1>
        <p class="lead">在这里统一进入课程工作区，管理你创建和加入的全部课程。</p>
      </div>
    </div>

    <div class="course-dashboard-grid">
      <aside class="course-filter-panel">
        <div class="course-filter-nav">
          <button
            v-for="item in filters"
            :key="item.value"
            class="course-filter-button"
            :class="{ active: activeFilter === item.value && viewMode === 'list' }"
            @click="selectFilter(item.value)"
          >
            <strong>{{ item.label }}</strong>
          </button>
        </div>

        <div class="course-filter-actions">
          <button
            class="course-filter-button course-filter-create-row"
            :class="{ active: viewMode === 'create' }"
            @click="openCreatePanel"
          >
            <strong>创建课程</strong>
          </button>
        </div>
      </aside>

      <section class="course-grid-panel">
        <div v-if="viewMode === 'create'" class="course-create-panel">
          <div class="course-create-head">
            <div>
              <p class="eyebrow">Create Course</p>
              <h2>创建课程</h2>
            </div>
            <button class="button ghost compact" @click="closeCreatePanel" :disabled="savingCourse">返回课程列表</button>
          </div>

          <form class="form course-create-form" @submit.prevent="submitCourseForm">
            <label class="field">
              <span>课程编号</span>
              <input v-model.trim="courseForm.courseCode" type="text" />
            </label>

            <label class="field">
              <span>课程名称</span>
              <input v-model.trim="courseForm.courseName" type="text" />
            </label>

            <label class="field">
              <span>课程简介</span>
              <textarea v-model.trim="courseForm.courseDescription" />
            </label>

            <p v-if="formError" class="error">{{ formError }}</p>

            <div class="inline-actions">
              <button class="button ghost" type="button" @click="closeCreatePanel" :disabled="savingCourse">取消</button>
              <button class="button primary" type="submit" :disabled="savingCourse">
                {{ savingCourse ? '创建中...' : '确认创建' }}
              </button>
            </div>
          </form>
        </div>

        <div v-else class="course-grid-scroll">
          <p v-if="errorMessage" class="error top-gap">{{ errorMessage }}</p>
          <p v-else-if="!filteredCourses.length && !loadingCourses" class="muted-copy top-gap">
            当前筛选下还没有课程，先创建一门课程开始使用。
          </p>

          <div v-else class="course-grid">
            <article v-for="course in filteredCourses" :key="course.id" class="course-card-panel">
              <div class="course-card-cover" :class="statusClass(course.status)">
                <span class="pill subtle">{{ roleLabel(course.myRole) }}</span>
                <span class="course-card-code">{{ course.courseCode }}</span>
              </div>

              <div class="course-card-body">
                <div>
                  <h3>{{ course.courseName }}</h3>
                  <p class="course-card-copy">{{ course.courseDescription || '当前课程还没有填写简介。' }}</p>
                </div>

                <div class="course-card-meta">
                  <span>状态：{{ statusLabel(course.status) }}</span>
                  <span>更新：{{ formatDateTime(course.updatedAt) }}</span>
                </div>

                <div class="inline-actions">
                  <button class="button primary compact" @click="enterCourse(course.id)">进入课程</button>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createCourse, listCourses } from '@/api/course'
import { useAuth } from '@/composables/useAuth'
import type { CourseRole, CourseStatus, CourseVO } from '@/types/course'
import { formatDateTime } from '@/utils/date'

type FilterValue = 'all' | 'joined' | 'managed' | 'owned'

const router = useRouter()
const route = useRoute()
const auth = useAuth()
const courses = ref<CourseVO[]>([])
const loadingCourses = ref(false)
const savingCourse = ref(false)
const errorMessage = ref('')
const formError = ref('')
const activeFilter = ref<FilterValue>('all')
const viewMode = ref<'list' | 'create'>('list')
const courseForm = reactive({
  courseCode: '',
  courseName: '',
  courseDescription: '',
})

const filters = [
  { value: 'all' as FilterValue, label: '全部课程' },
  { value: 'joined' as FilterValue, label: '我参与的' },
  { value: 'managed' as FilterValue, label: '我管理的' },
  { value: 'owned' as FilterValue, label: '我创建的' },
]

const filteredCourses = computed(() => {
  if (activeFilter.value === 'joined') {
    return courses.value.filter((course) => course.myRole === 'student')
  }
  if (activeFilter.value === 'managed') {
    return courses.value.filter((course) => course.myRole === 'owner' || course.myRole === 'teacher')
  }
  if (activeFilter.value === 'owned') {
    return courses.value.filter((course) => course.myRole === 'owner')
  }
  return courses.value
})

onMounted(async () => {
  await loadCourses()
  maybeOpenCreatePanelFromQuery()
})

async function loadCourses() {
  if (!auth.token.value) return
  loadingCourses.value = true
  errorMessage.value = ''
  try {
    courses.value = await listCourses(auth.token.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '课程列表加载失败'
  } finally {
    loadingCourses.value = false
  }
}

function openCreatePanel() {
  viewMode.value = 'create'
  formError.value = ''
  courseForm.courseCode = ''
  courseForm.courseName = ''
  courseForm.courseDescription = ''
}

function closeCreatePanel() {
  if (savingCourse.value) return
  viewMode.value = 'list'
  if (route.query.create) {
    void router.replace('/courses')
  }
}

async function submitCourseForm() {
  if (!auth.token.value) return
  savingCourse.value = true
  formError.value = ''
  try {
    const created = await createCourse(auth.token.value, {
      courseCode: courseForm.courseCode,
      courseName: courseForm.courseName,
      courseDescription: courseForm.courseDescription,
    })
    viewMode.value = 'list'
    await loadCourses()
    await enterCourse(created.id)
  } catch (error) {
    formError.value = error instanceof Error ? error.message : '课程创建失败'
  } finally {
    savingCourse.value = false
  }
}

async function enterCourse(courseId: number) {
  await router.push(`/courses/${courseId}/overview`)
}

function maybeOpenCreatePanelFromQuery() {
  if (route.query.create === '1') {
    openCreatePanel()
  }
}

function selectFilter(value: FilterValue) {
  viewMode.value = 'list'
  activeFilter.value = value
}

function roleLabel(role?: CourseRole) {
  switch (role) {
    case 'owner':
      return '创建者'
    case 'teacher':
      return '教师'
    case 'student':
      return '学生'
    default:
      return '成员'
  }
}

function statusLabel(status: CourseStatus) {
  switch (status) {
    case 'active':
      return '进行中'
    case 'archived':
      return '已归档'
    case 'deleted':
      return '已删除'
    default:
      return status
  }
}

function statusClass(status: CourseStatus) {
  switch (status) {
    case 'active':
      return 'is-active'
    case 'archived':
      return 'is-archived'
    case 'deleted':
      return 'is-deleted'
    default:
      return ''
  }
}
</script>
