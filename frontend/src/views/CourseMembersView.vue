<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Members</p>
        <h1>课程成员管理</h1>
        <p class="lead">按角色控制课程权限与资料访问，教师与创建者可以在这里维护成员名单。</p>
      </div>

      <div class="hero-actions">
        <button class="button ghost" @click="loadMembers" :disabled="loadingMembers">
          {{ loadingMembers ? '刷新中' : '刷新成员' }}
        </button>
        <button v-if="canAddMember" class="button primary" @click="dialogOpen = true">添加成员</button>
      </div>
    </div>

    <div class="members-layout">
      <article class="content-card members-table-card">
        <div class="members-toolbar">
          <label class="field search-field">
            <span>搜索成员</span>
            <input v-model.trim="searchKeyword" type="text" placeholder="按用户名搜索成员" />
          </label>
        </div>

        <p v-if="memberActionError" class="error">{{ memberActionError }}</p>
        <p v-if="memberError" class="error top-gap">{{ memberError }}</p>
        <p v-else-if="!filteredMembers.length && !loadingMembers" class="muted-copy top-gap">
          当前课程还没有可显示的成员。
        </p>

        <div v-else class="member-table">
          <div class="member-table-head">
            <span>成员</span>
            <span>角色</span>
            <span>加入时间</span>
            <span>操作</span>
          </div>

          <article v-for="member in filteredMembers" :key="member.id" class="member-table-row">
            <div class="member-profile-cell">
              <div class="member-avatar">{{ member.username.slice(0, 1).toUpperCase() }}</div>
              <div>
                <p class="member-name">{{ member.username }}</p>
                <p class="member-meta">ID {{ member.userId }}</p>
              </div>
            </div>

            <div class="member-role-cell">
              <template v-if="canChangeMemberRole(member)">
                <select
                  :value="member.role"
                  class="member-select member-select-light"
                  @change="handleRoleChange(member, ($event.target as HTMLSelectElement).value as EditableRole)"
                >
                  <option value="student">学生</option>
                  <option value="teacher">教师</option>
                </select>
              </template>
              <span v-else class="pill subtle">{{ roleLabel(member.role) }}</span>
            </div>

            <span class="member-row-copy">{{ formatDateTime(member.joinedAt) }}</span>

            <div class="inline-actions">
              <button
                v-if="canRemoveMember(member)"
                class="button danger compact"
                @click="handleRemoveMember(member)"
                :disabled="savingMemberMutation"
              >
                移除
              </button>
            </div>
          </article>
        </div>
      </article>

      <aside class="content-card permission-guide-card">
        <p class="eyebrow">Roles</p>
        <h2>角色权限说明</h2>
        <div class="permission-guide-list">
          <div class="permission-guide-item">
            <strong>教师</strong>
            <span>可管理课程与学生成员</span>
          </div>
          <div class="permission-guide-item">
            <strong>助教/教师视角</strong>
            <span>可管理资料与回复 Agent 会话</span>
          </div>
          <div class="permission-guide-item">
            <strong>学生</strong>
            <span>可访问资料并发起提问</span>
          </div>
        </div>
      </aside>
    </div>

    <div v-if="dialogOpen" class="modal-backdrop" @click.self="closeDialog">
      <section class="modal-card card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Add Member</p>
            <h3>添加成员</h3>
          </div>
          <button class="button ghost compact" @click="closeDialog">关闭</button>
        </div>

        <form class="form" @submit.prevent="submitMemberForm">
          <label class="field">
            <span>用户名</span>
            <input v-model.trim="memberForm.username" type="text" placeholder="输入用户名" />
          </label>

          <label class="field">
            <span>角色</span>
            <select v-model="memberForm.role">
              <option value="student">学生</option>
              <option v-if="course?.myRole === 'owner'" value="teacher">教师</option>
            </select>
          </label>

          <p v-if="memberActionError" class="error">{{ memberActionError }}</p>

          <button class="button primary" type="submit" :disabled="savingMemberMutation">
            {{ savingMemberMutation ? '提交中...' : '确认添加' }}
          </button>
        </form>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  addCourseMember,
  deleteCourseMember,
  listCourseMembers,
  updateCourseMember,
} from '@/api/course'
import { useAuth } from '@/composables/useAuth'
import { useCourseContext } from '@/composables/useCourseContext'
import type { CourseMemberVO, CourseRole } from '@/types/course'
import { formatDateTime } from '@/utils/date'

type EditableRole = 'teacher' | 'student'

const auth = useAuth()
const context = useCourseContext()
const course = computed(() => context.course.value)
const members = ref<CourseMemberVO[]>([])
const loadingMembers = ref(false)
const savingMemberMutation = ref(false)
const memberError = ref('')
const memberActionError = ref('')
const dialogOpen = ref(false)
const searchKeyword = ref('')
const memberForm = reactive({
  username: '',
  role: 'student' as EditableRole,
})

const currentUserId = computed(() => auth.user.value?.id ?? 0)
const canAddMember = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
const filteredMembers = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return members.value
  return members.value.filter((member) => member.username.toLowerCase().includes(keyword))
})

onMounted(loadMembers)

async function loadMembers() {
  if (!course.value || !context.token.value) return
  loadingMembers.value = true
  memberError.value = ''
  try {
    members.value = await listCourseMembers(context.token.value, course.value.id)
  } catch (error) {
    memberError.value = error instanceof Error ? error.message : '成员列表加载失败'
    members.value = []
  } finally {
    loadingMembers.value = false
  }
}

function closeDialog() {
  if (savingMemberMutation.value) return
  dialogOpen.value = false
  memberActionError.value = ''
}

async function submitMemberForm() {
  if (!course.value || !context.token.value) return
  savingMemberMutation.value = true
  memberActionError.value = ''
  try {
    await addCourseMember(context.token.value, course.value.id, {
      username: memberForm.username,
      role: memberForm.role,
    })
    memberForm.username = ''
    memberForm.role = 'student'
    dialogOpen.value = false
    await loadMembers()
  } catch (error) {
    memberActionError.value = error instanceof Error ? error.message : '成员添加失败'
  } finally {
    savingMemberMutation.value = false
  }
}

async function handleRoleChange(member: CourseMemberVO, role: EditableRole) {
  if (!course.value || !context.token.value || member.role === role) return
  savingMemberMutation.value = true
  memberActionError.value = ''
  try {
    await updateCourseMember(context.token.value, course.value.id, member.id, { role })
    await loadMembers()
  } catch (error) {
    memberActionError.value = error instanceof Error ? error.message : '成员角色更新失败'
  } finally {
    savingMemberMutation.value = false
  }
}

async function handleRemoveMember(member: CourseMemberVO) {
  if (!course.value || !context.token.value) return
  const confirmed = window.confirm(`确认将 ${member.username} 移出课程吗？`)
  if (!confirmed) return
  savingMemberMutation.value = true
  memberActionError.value = ''
  try {
    await deleteCourseMember(context.token.value, course.value.id, member.id)
    await loadMembers()
  } catch (error) {
    memberActionError.value = error instanceof Error ? error.message : '成员移除失败'
  } finally {
    savingMemberMutation.value = false
  }
}

function canChangeMemberRole(member: CourseMemberVO) {
  if (!course.value) return false
  if (course.value.myRole !== 'owner') return false
  if (member.userId === currentUserId.value) return false
  return member.role === 'teacher' || member.role === 'student'
}

function canRemoveMember(member: CourseMemberVO) {
  if (!course.value) return false
  if (member.userId === currentUserId.value) return false
  if (course.value.myRole === 'owner') {
    return member.role === 'teacher' || member.role === 'student'
  }
  if (course.value.myRole === 'teacher') {
    return member.role === 'student'
  }
  return false
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
</script>
