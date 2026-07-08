<template>
  <main class="dashboard-shell">
    <section class="dashboard-grid">
      <aside class="sidebar card">
        <div class="sidebar-head">
          <div>
            <p class="eyebrow">Course Agent</p>
            <h1 class="dashboard-title">课程中心</h1>
          </div>
          <button class="button ghost compact" @click="handleLogout">退出</button>
        </div>

        <div class="sidebar-profile">
          <p class="label">当前用户</p>
          <p class="value">{{ auth.user.value?.username }}</p>
        </div>

        <div class="sidebar-section">
          <div class="section-head">
            <p class="label">我的课程</p>
            <div class="inline-actions">
              <button class="button primary compact" @click="openCreateDialog">新建课程</button>
              <button class="button ghost compact" @click="reloadCourses" :disabled="loadingCourses">
                {{ loadingCourses ? '刷新中' : '刷新' }}
              </button>
            </div>
          </div>

          <p v-if="courseError" class="error">{{ courseError }}</p>
          <p v-else-if="!courses.length && !loadingCourses" class="muted-copy">暂无课程，先创建一门课程开始使用。</p>

          <button
            v-for="course in courses"
            :key="course.id"
            class="course-list-item"
            :class="{ active: course.id === selectedCourseId }"
            @click="selectCourse(course.id)"
          >
            <span class="course-list-name">{{ course.courseName }}</span>
            <span class="course-list-meta">{{ course.courseCode }} · {{ roleLabel(course.myRole) }}</span>
          </button>
        </div>
      </aside>

      <section class="content-stack">
        <article class="card detail-card">
          <template v-if="selectedCourse">
            <div class="detail-head">
              <div>
                <p class="eyebrow">{{ selectedCourse.courseCode }}</p>
                <h2>{{ selectedCourse.courseName }}</h2>
              </div>
              <div class="inline-actions">
                <span class="pill">{{ roleLabel(selectedCourse.myRole) }}</span>
                <button v-if="canEditCourse" class="button ghost compact" @click="openEditDialog">编辑课程</button>
                <button v-if="canDeleteCourse" class="button danger compact" @click="handleDeleteCourse" :disabled="savingCourse">
                  {{ savingCourse ? '处理中' : '删除课程' }}
                </button>
              </div>
            </div>

            <p class="lead detail-copy">
              {{ selectedCourse.courseDescription || '当前课程还没有填写课程简介。' }}
            </p>

            <div class="detail-metrics">
              <div class="metric-card">
                <span class="label">课程状态</span>
                <strong>{{ statusLabel(selectedCourse.status) }}</strong>
              </div>
              <div class="metric-card">
                <span class="label">创建时间</span>
                <strong>{{ formatDateTime(selectedCourse.createdAt) }}</strong>
              </div>
              <div class="metric-card">
                <span class="label">最近更新</span>
                <strong>{{ formatDateTime(selectedCourse.updatedAt) }}</strong>
              </div>
            </div>
          </template>

          <div v-else class="empty-state">
            <p class="eyebrow">Course Agent</p>
            <h2>先选择一门课程</h2>
            <p class="lead">左侧会列出你已加入的课程，选择后可以查看详情和成员信息。</p>
          </div>
        </article>

        <article class="card members-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">Roster</p>
              <h3>成员列表</h3>
            </div>
            <button
              class="button ghost compact"
              @click="reloadMembers"
              :disabled="!selectedCourse || loadingMembers"
            >
              {{ loadingMembers ? '刷新中' : '刷新成员' }}
            </button>
          </div>

          <p v-if="memberError" class="error">{{ memberError }}</p>
          <p v-else-if="!selectedCourse" class="muted-copy">选择课程后查看成员列表。</p>
          <p v-else-if="!members.length && !loadingMembers" class="muted-copy">当前课程还没有可显示的成员。</p>

          <div v-else class="member-grid">
            <article v-for="member in members" :key="member.id" class="member-card">
              <div>
                <p class="member-name">{{ member.username }}</p>
                <p class="member-meta">ID {{ member.userId }}</p>
              </div>
              <div class="member-side">
                <span class="pill subtle">{{ roleLabel(member.role) }}</span>
                <span class="member-meta">加入于 {{ formatDateTime(member.joinedAt) }}</span>
              </div>
            </article>
          </div>
        </article>
      </section>
    </section>

    <div v-if="courseDialog.open" class="modal-backdrop" @click.self="closeCourseDialog">
      <section class="modal-card card">
        <div class="section-head">
          <div>
            <p class="eyebrow">{{ courseDialog.mode === 'create' ? 'Create Course' : 'Edit Course' }}</p>
            <h3>{{ courseDialog.mode === 'create' ? '新建课程' : '编辑课程' }}</h3>
          </div>
          <button class="button ghost compact" @click="closeCourseDialog">关闭</button>
        </div>

        <form class="form" @submit.prevent="submitCourseForm">
          <label class="field">
            <span>课程编号</span>
            <input v-model.trim="courseForm.courseCode" type="text" :disabled="courseDialog.mode === 'edit'" />
          </label>

          <label class="field">
            <span>课程名称</span>
            <input v-model.trim="courseForm.courseName" type="text" />
          </label>

          <label class="field">
            <span>课程简介</span>
            <textarea v-model.trim="courseForm.courseDescription"></textarea>
          </label>

          <p v-if="courseFormError" class="error">{{ courseFormError }}</p>

          <button class="button primary" type="submit" :disabled="savingCourse">
            {{ savingCourse ? '提交中...' : courseDialog.mode === 'create' ? '创建课程' : '保存修改' }}
          </button>
        </form>
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createCourse, deleteCourse, getCourse, listCourseMembers, listCourses, updateCourse } from '@/api/course'
import { logout } from '@/api/user'
import { useAuth } from '@/composables/useAuth'
import type { CourseMemberVO, CourseRole, CourseStatus, CourseVO } from '@/types/course'
import { formatDateTime } from '@/utils/date'

const auth = useAuth()
const router = useRouter()

const courses = ref<CourseVO[]>([])
const members = ref<CourseMemberVO[]>([])
const loadingCourses = ref(false)
const loadingMembers = ref(false)
const savingCourse = ref(false)
const courseError = ref('')
const memberError = ref('')
const courseFormError = ref('')
const selectedCourseId = ref<number | null>(null)
const selectedCourse = ref<CourseVO | null>(null)
const courseDialog = reactive({
  open: false,
  mode: 'create' as 'create' | 'edit',
})
const courseForm = reactive({
  courseCode: '',
  courseName: '',
  courseDescription: '',
})

const token = computed(() => auth.token.value)
const canEditCourse = computed(() => selectedCourse.value?.myRole === 'owner' || selectedCourse.value?.myRole === 'teacher')
const canDeleteCourse = computed(() => selectedCourse.value?.myRole === 'owner')

async function loadCourses() {
  if (!token.value) return
  loadingCourses.value = true
  courseError.value = ''
  try {
    const data = await listCourses(token.value)
    courses.value = data

    if (!data.length) {
      selectedCourseId.value = null
      selectedCourse.value = null
      members.value = []
      return
    }

    const nextCourseId = selectedCourseId.value && data.some((course) => course.id === selectedCourseId.value)
      ? selectedCourseId.value
      : data[0].id

    await selectCourse(nextCourseId)
  } catch (error) {
    courseError.value = error instanceof Error ? error.message : '课程列表加载失败'
  } finally {
    loadingCourses.value = false
  }
}

async function selectCourse(courseId: number) {
  if (!token.value) return
  selectedCourseId.value = courseId
  courseError.value = ''
  try {
    selectedCourse.value = await getCourse(token.value, courseId)
    await loadMembers(courseId)
  } catch (error) {
    courseError.value = error instanceof Error ? error.message : '课程详情加载失败'
  }
}

async function loadMembers(courseId: number) {
  if (!token.value) return
  loadingMembers.value = true
  memberError.value = ''
  try {
    members.value = await listCourseMembers(token.value, courseId)
  } catch (error) {
    memberError.value = error instanceof Error ? error.message : '成员列表加载失败'
    members.value = []
  } finally {
    loadingMembers.value = false
  }
}

async function reloadCourses() {
  await loadCourses()
}

async function reloadMembers() {
  if (!selectedCourseId.value) return
  await loadMembers(selectedCourseId.value)
}

function openCreateDialog() {
  courseDialog.open = true
  courseDialog.mode = 'create'
  courseFormError.value = ''
  courseForm.courseCode = ''
  courseForm.courseName = ''
  courseForm.courseDescription = ''
}

function openEditDialog() {
  if (!selectedCourse.value) return
  courseDialog.open = true
  courseDialog.mode = 'edit'
  courseFormError.value = ''
  courseForm.courseCode = selectedCourse.value.courseCode
  courseForm.courseName = selectedCourse.value.courseName
  courseForm.courseDescription = selectedCourse.value.courseDescription
}

function closeCourseDialog() {
  if (savingCourse.value) return
  courseDialog.open = false
  courseFormError.value = ''
}

async function submitCourseForm() {
  if (!token.value) return
  savingCourse.value = true
  courseFormError.value = ''
  try {
    if (courseDialog.mode === 'create') {
      const created = await createCourse(token.value, {
        courseCode: courseForm.courseCode,
        courseName: courseForm.courseName,
        courseDescription: courseForm.courseDescription,
      })
      courseDialog.open = false
      await loadCourses()
      await selectCourse(created.id)
      return
    }

    if (!selectedCourseId.value) return
    const updated = await updateCourse(token.value, selectedCourseId.value, {
      courseName: courseForm.courseName,
      courseDescription: courseForm.courseDescription,
    })
    selectedCourse.value = updated
    courseDialog.open = false
    await loadCourses()
  } catch (error) {
    courseFormError.value = error instanceof Error ? error.message : '课程保存失败'
  } finally {
    savingCourse.value = false
  }
}

async function handleDeleteCourse() {
  if (!token.value || !selectedCourseId.value || !selectedCourse.value) return
  const confirmed = window.confirm(`确认删除课程“${selectedCourse.value.courseName}”吗？`)
  if (!confirmed) return

  savingCourse.value = true
  courseError.value = ''
  try {
    await deleteCourse(token.value, selectedCourseId.value)
    selectedCourseId.value = null
    selectedCourse.value = null
    members.value = []
    await loadCourses()
  } catch (error) {
    courseError.value = error instanceof Error ? error.message : '课程删除失败'
  } finally {
    savingCourse.value = false
  }
}

async function handleLogout() {
  if (!token.value) return
  await logout(token.value)
  auth.clear()
  await router.push('/login')
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
      return '未知角色'
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

onMounted(async () => {
  await loadCourses()
})
</script>
