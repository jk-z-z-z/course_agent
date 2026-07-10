<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Agent</p>
        <h1>课程 Agent</h1>
        <p class="lead">Agent 配置与会话会放在独立页面，避免继续和课程详情混在一起。</p>
      </div>
    </div>

    <div class="agent-page-layout">
      <CourseAgentPanel
        v-if="course && token"
        :course-id="course.id"
        :token="token"
        :can-manage="canManage"
      />

      <aside class="content-card agent-aside-card">
        <p class="eyebrow">Guide</p>
        <h2>Agent 使用说明</h2>
        <div class="permission-guide-list">
          <div class="permission-guide-item">
            <strong>教师 / 创建者</strong>
            <span>可配置 Agent 名称、提示词、状态，以及创建会话入口。</span>
          </div>
          <div class="permission-guide-item">
            <strong>学生</strong>
            <span>可在自己会话内提问，查看引用资料片段和回答结果。</span>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CourseAgentPanel from '@/components/CourseAgentPanel.vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const canManage = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
</script>
