<template>
  <section class="materials-detail-panel panel materials-detail-shell">
    <template v-if="detail">
      <div class="section-head">
        <div>
          <p class="eyebrow">Detail</p>
          <h4 class="material-title">{{ detail.nodeName }}</h4>
        </div>
        <div v-if="canManage" class="inline-actions">
          <button class="button ghost compact" @click="$emit('rename')">重命名</button>
          <button class="button danger compact" @click="$emit('remove')">删除</button>
        </div>
      </div>

      <div class="material-detail-grid">
        <div class="material-stat-card">
          <p class="label">类型</p>
          <p class="value minor">{{ detail.nodeType === 'folder' ? '文件夹' : '文件' }}</p>
        </div>
        <div class="material-stat-card">
          <p class="label">大小</p>
          <p class="value minor">{{ formatFileSize(detail.fileSize) }}</p>
        </div>
        <div class="material-stat-card">
          <p class="label">版本</p>
          <p class="value minor">v{{ detail.latestVersionNo }}</p>
        </div>
        <div class="material-stat-card">
          <p class="label">更新时间</p>
          <p class="value minor">{{ formatDateTime(detail.updatedAt) }}</p>
        </div>
      </div>

      <div v-if="detail.nodeType === 'file'" class="inline-actions top-gap">
        <button class="button ghost compact" @click="$emit('preview')" :disabled="loadingBlob">预览</button>
        <button class="button primary compact" @click="$emit('download')" :disabled="loadingBlob">下载</button>
      </div>
    </template>

    <div v-else class="empty-state small">
      <p class="muted-copy">从左侧选择一个资料节点查看详情。</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { MaterialDetailVO } from '@/types/material'
import { formatDateTime } from '@/utils/date'

defineProps<{
  detail: MaterialDetailVO | null
  canManage: boolean
  loadingBlob: boolean
}>()

defineEmits<{
  rename: []
  remove: []
  preview: []
  download: []
}>()

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
</script>
