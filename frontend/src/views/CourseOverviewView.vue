<template>
  <section class="page-section">
    <div class="course-page-head course-page-head-compact">
      <div class="course-overview-title-row">
        <p class="eyebrow">Home</p>
        <div class="course-overview-heading">
          <h1>{{ course?.courseName || '课程详情' }}</h1>
          <span v-if="course" class="course-status-inline">{{ statusLabel(course.status) }}</span>
        </div>
      </div>
      <div v-if="course" class="inline-actions">
        <button
          v-if="canEditCourse"
          class="button ghost compact"
          type="button"
          @click="editingCourse = !editingCourse"
        >
          {{ editingCourse ? '收起编辑' : '编辑课程' }}
        </button>
        <button
          v-if="canDeleteCourse"
          class="button danger compact"
          type="button"
          :disabled="savingCourse"
          @click="removeCourse"
        >
          删除课程
        </button>
      </div>
    </div>

    <div class="overview-grid" v-if="course">
      <article class="workspace-panel overview-main-card">
        <div class="overview-summary">
          <div class="overview-copy">
            <p class="lead overview-lead">{{ course.courseDescription || '当前课程还没有填写课程简介。' }}</p>
          </div>
        </div>
      </article>

      <article v-if="editingCourse && canEditCourse" class="workspace-panel overview-tabs-card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Edit</p>
            <h2>编辑课程</h2>
          </div>
        </div>

        <form class="form" @submit.prevent="saveCourse">
          <label class="field">
            <span>课程名称</span>
            <input v-model.trim="courseForm.courseName" type="text" />
          </label>
          <label class="field">
            <span>课程简介</span>
            <textarea v-model.trim="courseForm.courseDescription" />
          </label>
          <p v-if="courseActionError" class="error">{{ courseActionError }}</p>
          <button class="button primary" type="submit" :disabled="savingCourse">
            {{ savingCourse ? '保存中' : '保存修改' }}
          </button>
        </form>
      </article>

      <article class="workspace-panel overview-tabs-card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Information</p>
            <h2>基本信息</h2>
          </div>
        </div>

        <div class="overview-info-list">
          <div class="overview-info-item">
            <span class="label">课程编号</span>
            <strong>{{ course.courseCode }}</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">课程成员</span>
            <strong>{{ membersCount }} 人</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">创建时间</span>
            <strong>{{ formatDateTime(course.createdAt) }}</strong>
          </div>
          <div class="overview-info-item">
            <span class="label">更新时间</span>
            <strong>{{ formatDateTime(course.updatedAt) }}</strong>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { deleteCourse, listCourseMembers, updateCourse } from '@/api/course'
import { useCourseContext } from '@/composables/useCourseContext'
import type { CourseStatus } from '@/types/course'
import { formatDateTime } from '@/utils/date'

const router = useRouter()
const context = useCourseContext()
const course = computed(() => context.course.value)
const membersCount = ref(0)
const editingCourse = ref(false)
const savingCourse = ref(false)
const courseActionError = ref('')
const courseForm = reactive({
  courseName: '',
  courseDescription: '',
})
const canEditCourse = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
const canDeleteCourse = computed(() => course.value?.myRole === 'owner')

onMounted(loadMembersCount)
watch(course, syncCourseForm, { immediate: true })

async function loadMembersCount() {
  if (!course.value || !context.token.value) return
  try {
    const members = await listCourseMembers(context.token.value, course.value.id)
    membersCount.value = members.length
  } catch {
    membersCount.value = 0
  }
}

function syncCourseForm() {
  courseForm.courseName = course.value?.courseName ?? ''
  courseForm.courseDescription = course.value?.courseDescription ?? ''
}

async function saveCourse() {
  if (!course.value || !context.token.value) return
  savingCourse.value = true
  courseActionError.value = ''
  try {
    await updateCourse(context.token.value, course.value.id, {
      courseName: courseForm.courseName,
      courseDescription: courseForm.courseDescription,
    })
    await context.refreshCourse()
    editingCourse.value = false
  } catch (error) {
    courseActionError.value = error instanceof Error ? error.message : '课程保存失败'
  } finally {
    savingCourse.value = false
  }
}

async function removeCourse() {
  if (!course.value || !context.token.value) return
  const confirmed = window.confirm(`确认删除课程“${course.value.courseName}”吗？`)
  if (!confirmed) return
  savingCourse.value = true
  courseActionError.value = ''
  try {
    await deleteCourse(context.token.value, course.value.id)
    await router.push('/courses')
  } catch (error) {
    courseActionError.value = error instanceof Error ? error.message : '课程删除失败'
  } finally {
    savingCourse.value = false
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
