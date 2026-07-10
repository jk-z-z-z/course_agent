<template>
  <section class="page-section">
    <div class="course-page-head">
      <div>
        <p class="eyebrow">Overview</p>
        <h1>{{ course?.courseName || '课程详情' }}</h1>
        <p class="lead">{{ course?.courseDescription || '当前课程还没有填写课程简介。' }}</p>
      </div>

      <div class="hero-actions" v-if="course">
        <button v-if="canEditCourse" class="button ghost" @click="openEditDialog">编辑课程</button>
        <button
          v-if="canDeleteCourse"
          class="button danger"
          @click="handleDeleteCourse"
          :disabled="savingCourse"
        >
          {{ savingCourse ? '处理中' : '删除课程' }}
        </button>
      </div>
    </div>

    <div class="overview-grid" v-if="course">
      <article class="workspace-panel overview-main-card">
        <div class="overview-summary">
          <div class="overview-cover">
            <span class="pill">课程名片</span>
          </div>

          <div class="overview-copy">
            <p class="eyebrow">{{ course.courseCode }}</p>
            <h2>{{ course.courseName }}</h2>
            <div class="inline-actions">
              <span class="pill subtle">{{ roleLabel(course.myRole) }}</span>
              <span class="pill subtle">{{ statusLabel(course.status) }}</span>
            </div>
          </div>
        </div>

        <div class="detail-metrics">
          <div class="metric-card">
            <span class="label">课程状态</span>
            <strong>{{ statusLabel(course.status) }}</strong>
          </div>
          <div class="metric-card">
            <span class="label">课程成员</span>
            <strong>{{ membersCount }} 人</strong>
          </div>
          <div class="metric-card">
            <span class="label">最近更新</span>
            <strong>{{ formatDateTime(course.updatedAt) }}</strong>
          </div>
        </div>
      </article>

      <article class="workspace-panel overview-tabs-card">
        <div class="overview-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            class="overview-tab-button"
            :class="{ active: activeTab === tab.value }"
            @click="activeTab = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>

        <div v-if="activeTab === 'intro'" class="overview-tab-panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Introduction</p>
              <h2>课程简介</h2>
            </div>
          </div>
          <div class="overview-rich-block">
            <p class="lead overview-lead">{{ course.courseDescription || '当前课程还没有填写课程简介。' }}</p>
          </div>
        </div>

        <div v-else-if="activeTab === 'info'" class="overview-tab-panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Information</p>
              <h2>课程信息</h2>
            </div>
          </div>

          <div class="overview-info-list">
            <div class="overview-info-item">
              <span class="label">课程编号</span>
              <strong>{{ course.courseCode }}</strong>
            </div>
            <div class="overview-info-item">
              <span class="label">创建时间</span>
              <strong>{{ formatDateTime(course.createdAt) }}</strong>
            </div>
            <div class="overview-info-item">
              <span class="label">更新时间</span>
              <strong>{{ formatDateTime(course.updatedAt) }}</strong>
            </div>
            <div class="overview-info-item">
              <span class="label">我的角色</span>
              <strong>{{ roleLabel(course.myRole) }}</strong>
            </div>
          </div>
        </div>

        <div v-else class="overview-tab-panel">
          <div class="section-head">
            <div>
              <p class="eyebrow">Workspace</p>
              <h2>进入课程工作区</h2>
            </div>
          </div>

          <div class="overview-links">
            <RouterLink :to="`/courses/${course.id}/members`" class="overview-link-card">
              <strong>成员管理</strong>
              <span>维护成员、角色与权限边界</span>
            </RouterLink>
            <RouterLink :to="`/courses/${course.id}/materials`" class="overview-link-card">
              <strong>资料中心</strong>
              <span>管理文件夹、上传资料与访问控制</span>
            </RouterLink>
            <RouterLink :to="`/courses/${course.id}/agent`" class="overview-link-card">
              <strong>课程 Agent</strong>
              <span>配置课程助教并进入问答会话</span>
            </RouterLink>
          </div>
        </div>
      </article>
    </div>

    <div v-if="dialogOpen" class="modal-backdrop" @click.self="closeDialog">
      <section class="modal-card card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Edit Course</p>
            <h3>编辑课程</h3>
          </div>
          <button class="button ghost compact" @click="closeDialog">关闭</button>
        </div>

        <form class="form" @submit.prevent="submitCourseForm">
          <label class="field">
            <span>课程编号</span>
            <input :value="course?.courseCode || ''" type="text" disabled />
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

          <button class="button primary" type="submit" :disabled="savingCourse">
            {{ savingCourse ? '保存中...' : '保存修改' }}
          </button>
        </form>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { deleteCourse, listCourseMembers, updateCourse } from '@/api/course'
import { useCourseContext } from '@/composables/useCourseContext'
import type { CourseRole, CourseStatus } from '@/types/course'
import { formatDateTime } from '@/utils/date'

type OverviewTab = 'intro' | 'info' | 'workspace'

const context = useCourseContext()
const course = computed(() => context.course.value)
const router = useRouter()
const membersCount = ref(0)
const savingCourse = ref(false)
const dialogOpen = ref(false)
const formError = ref('')
const activeTab = ref<OverviewTab>('intro')
const courseForm = reactive({
  courseName: '',
  courseDescription: '',
})
const canEditCourse = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
const canDeleteCourse = computed(() => course.value?.myRole === 'owner')
const tabs = [
  { value: 'intro' as OverviewTab, label: '课程简介' },
  { value: 'info' as OverviewTab, label: '基本信息' },
  { value: 'workspace' as OverviewTab, label: '工作区入口' },
]

onMounted(loadMembersCount)

async function loadMembersCount() {
  if (!course.value || !context.token.value) return
  try {
    const members = await listCourseMembers(context.token.value, course.value.id)
    membersCount.value = members.length
  } catch {
    membersCount.value = 0
  }
}

function openEditDialog() {
  if (!course.value) return
  dialogOpen.value = true
  formError.value = ''
  courseForm.courseName = course.value.courseName
  courseForm.courseDescription = course.value.courseDescription
}

function closeDialog() {
  if (savingCourse.value) return
  dialogOpen.value = false
}

async function submitCourseForm() {
  if (!course.value || !context.token.value) return
  savingCourse.value = true
  formError.value = ''
  try {
    await updateCourse(context.token.value, course.value.id, {
      courseName: courseForm.courseName,
      courseDescription: courseForm.courseDescription,
    })
    dialogOpen.value = false
    await context.refreshCourse()
  } catch (error) {
    formError.value = error instanceof Error ? error.message : '课程保存失败'
  } finally {
    savingCourse.value = false
  }
}

async function handleDeleteCourse() {
  if (!course.value || !context.token.value) return
  const confirmed = window.confirm(`确认删除课程“${course.value.courseName}”吗？`)
  if (!confirmed) return
  savingCourse.value = true
  try {
    await deleteCourse(context.token.value, course.value.id)
    await router.push('/courses')
  } finally {
    savingCourse.value = false
  }
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
</script>
