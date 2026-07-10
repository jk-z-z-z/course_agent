<template>
  <section class="page-section">
    <div class="page-hero">
      <div>
        <p class="eyebrow">Materials</p>
        <h1>课程资料</h1>
        <p class="lead">资料模块先接入现有组件，再逐步贴近你提供的资料页原型。</p>
      </div>
    </div>

    <div class="materials-page-layout">
      <CourseMaterialsPanel
        v-if="course && token"
        :course-id="course.id"
        :token="token"
        :can-manage="canManage"
      />

      <aside class="content-card materials-aside-card">
        <p class="eyebrow">Access</p>
        <h2>资料访问权限</h2>
        <div class="permission-guide-list">
          <div class="permission-guide-item">
            <strong>教师 / 创建者</strong>
            <span>可创建文件夹、上传资料、重命名与删除节点。</span>
          </div>
          <div class="permission-guide-item">
            <strong>学生</strong>
            <span>可预览与下载已发布资料，不参与目录管理。</span>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CourseMaterialsPanel from '@/components/CourseMaterialsPanel.vue'
import { useCourseContext } from '@/composables/useCourseContext'

const context = useCourseContext()
const course = computed(() => context.course.value)
const token = computed(() => context.token.value)
const canManage = computed(() => course.value?.myRole === 'owner' || course.value?.myRole === 'teacher')
</script>
