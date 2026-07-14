<template>
  <section class="page-section course-dashboard">
    <div class="course-page-head course-home-head">
      <div class="course-home-copy">
        <h1>欢迎回来，{{ auth.user.value?.username || '同学' }}</h1>
        <p class="lead course-home-lead">继续管理课程与学习资料</p>
      </div>
    </div>

    <div class="course-dashboard-grid">
      <aside class="course-filter-panel">
        <div class="course-filter-nav">
          <button
            v-for="item in filters"
            :key="item.value"
            class="course-filter-button"
            :class="{ active: activeFilter === item.value }"
            @click="selectFilter(item.value)"
          >
            <span class="course-filter-icon">{{ item.icon }}</span>
            <span>
              <strong>{{ item.label }}</strong>
            </span>
          </button>
        </div>

        <div class="course-filter-actions">
          <button
            class="course-filter-button course-filter-create-row"
            type="button"
            @click="openCreatePanel"
          >
            <span class="course-filter-icon">＋</span>
            <span>
              <strong>创建课程</strong>
            </span>
          </button>
        </div>
      </aside>

      <section class="course-grid-panel">
        <div class="course-grid-scroll">
          <p v-if="errorMessage" class="error top-gap">{{ errorMessage }}</p>
          <p v-else-if="loadingCourses" class="muted-copy top-gap">正在加载课程...</p>
          <p v-else-if="!filteredCourses.length && !loadingCourses" class="muted-copy top-gap">
            当前分类下还没有课程。
          </p>

          <div v-else class="course-grid">
            <article v-for="course in filteredCourses" :key="course.id" class="course-card-panel">
              <div class="course-card-cover" :class="coverClass(course)">
                <span class="course-card-cover-code">{{ course.courseCode }}</span>
              </div>
              <div class="course-card-body">
                <div class="course-card-topline">
                  <span class="pill subtle">{{ roleLabel(course.myRole) }}</span>
                  <span class="course-card-code">{{ statusLabel(course.status) }}</span>
                </div>

                <div class="course-card-main">
                  <h3>{{ course.courseName }}</h3>
                  <p class="course-card-copy">{{ course.courseDescription || '当前课程还没有填写简介。' }}</p>
                </div>

                <div class="course-card-footer">
                  <div class="course-card-meta">
                    <span>{{ statusLabel(course.status) }}</span>
                    <span>{{ formatDateTime(course.updatedAt) }}</span>
                  </div>
                  <button
                    v-if="course.myRole"
                    class="button primary compact"
                    @click="enterCourse(course.id)"
                  >
                    进入课程
                  </button>
                  <button
                    v-else
                    class="button ghost compact"
                    @click="joinAndEnterCourse(course.id)"
                    :disabled="joiningCourseId === course.id"
                  >
                    {{ joiningCourseId === course.id ? '加入中' : '加入课程' }}
                  </button>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>
    </div>

    <div v-if="createDialogOpen" class="modal-backdrop course-create-backdrop" @click.self="closeCreatePanel">
      <section class="modal-card card course-create-modal">
        <div class="section-head">
          <div>
            <h3>创建课程</h3>
          </div>
          <button class="button ghost compact" type="button" @click="closeCreatePanel" :disabled="savingCourse">
            关闭
          </button>
        </div>

        <form class="form course-create-form" @submit.prevent="submitCourseForm">
          <label class="field">
            <span>课程编号</span>
            <input v-model.trim="courseForm.courseCode" type="text" placeholder="例如 CS101" />
          </label>

          <label class="field">
            <span>课程名称</span>
            <input v-model.trim="courseForm.courseName" type="text" placeholder="输入课程名称" />
          </label>

          <label class="field">
            <span>课程简介</span>
            <textarea v-model.trim="courseForm.courseDescription" placeholder="简要说明课程内容和资料范围" />
          </label>

          <p v-if="formError" class="error">{{ formError }}</p>

          <div class="inline-actions">
            <button class="button ghost" type="button" @click="closeCreatePanel" :disabled="savingCourse">取消</button>
            <button class="button primary" type="submit" :disabled="savingCourse">
              {{ savingCourse ? '创建中...' : '确认创建' }}
            </button>
          </div>
        </form>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createCourse, joinCourse, listDiscoverableCourses } from '@/api/course'
import { useAuth } from '@/composables/useAuth'
import type { CourseRole, CourseStatus, CourseVO } from '@/types/course'
import { formatDateTime } from '@/utils/date'

type FilterValue = 'discover' | 'student' | 'managed' | 'owned'

const router = useRouter()
const route = useRoute()
const auth = useAuth()
const courses = ref<CourseVO[]>([])
const loadingCourses = ref(false)
const savingCourse = ref(false)
const joiningCourseId = ref<number | null>(null)
const errorMessage = ref('')
const formError = ref('')
const activeFilter = ref<FilterValue>('discover')
const createDialogOpen = ref(false)
const courseForm = reactive({
  courseCode: '',
  courseName: '',
  courseDescription: '',
})

const filters = [
  { value: 'discover' as FilterValue, label: '全局课程', icon: '⌂' },
  { value: 'student' as FilterValue, label: '我的课程', icon: '▣' },
  { value: 'managed' as FilterValue, label: '我管理的', icon: '☰' },
  { value: 'owned' as FilterValue, label: '我创建的', icon: '◇' },
]

const filteredCourses = computed(() => {
  if (activeFilter.value === 'student') {
    return courses.value.filter((course) => course.myRole === 'student')
  }
  if (activeFilter.value === 'managed') {
    return courses.value.filter((course) => course.myRole === 'teacher')
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
    courses.value = await listDiscoverableCourses(auth.token.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '课程列表加载失败'
  } finally {
    loadingCourses.value = false
  }
}

function openCreatePanel() {
  createDialogOpen.value = true
  formError.value = ''
  courseForm.courseCode = ''
  courseForm.courseName = ''
  courseForm.courseDescription = ''
}

function closeCreatePanel() {
  if (savingCourse.value) return
  createDialogOpen.value = false
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
    createDialogOpen.value = false
    await loadCourses()
    await enterCourse(created.id)
  } catch (error) {
    formError.value = error instanceof Error ? error.message : '课程创建失败'
  } finally {
    savingCourse.value = false
  }
}

async function enterCourse(courseId: number) {
  await router.push(`/courses/${courseId}/agent`)
}

async function joinAndEnterCourse(courseId: number) {
  if (!auth.token.value) return
  joiningCourseId.value = courseId
  errorMessage.value = ''
  try {
    await joinCourse(auth.token.value, courseId)
    await loadCourses()
    await enterCourse(courseId)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加入课程失败'
  } finally {
    joiningCourseId.value = null
  }
}

function maybeOpenCreatePanelFromQuery() {
  if (route.query.create === '1') {
    openCreatePanel()
  }
}

function selectFilter(value: FilterValue) {
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
      return '未加入'
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

function coverClass(course: CourseVO) {
  const seed = course.id % 3
  if (seed === 1) return 'mint'
  if (seed === 2) return 'amber'
  return 'blue'
}

</script>
