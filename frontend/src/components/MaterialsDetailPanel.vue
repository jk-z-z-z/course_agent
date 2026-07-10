<template>
  <section class="materials-detail-panel panel materials-detail-shell">
    <template v-if="detail">
      <div class="section-head materials-preview-head">
        <div>
          <h4 class="material-title">{{ detail.nodeName }}</h4>
        </div>
        <div class="inline-actions">
          <button
            v-if="detail.nodeType === 'file'"
            class="button ghost compact"
            @click="$emit('preview')"
            :disabled="loadingBlob"
          >
            新窗口预览
          </button>
          <button
            v-if="detail.nodeType === 'file'"
            class="button primary compact"
            @click="$emit('download')"
            :disabled="loadingBlob"
          >
            下载
          </button>
          <button v-if="canManage" class="button ghost compact" @click="$emit('rename')">重命名</button>
          <button v-if="canManage" class="button danger compact" @click="$emit('remove')">删除</button>
        </div>
      </div>

      <div v-if="detail.nodeType === 'folder'" class="materials-folder-view">
        <div v-if="folderChildren.length" class="materials-folder-list">
          <article v-for="node in folderChildren" :key="node.id" class="materials-folder-row">
            <div class="materials-folder-row-main">
              <span class="tree-node-icon" :class="node.type === 'folder' ? 'folder' : 'file'">
                {{ node.type === 'folder' ? 'F' : 'D' }}
              </span>
              <div class="materials-folder-row-copy">
                <strong>{{ node.name }}</strong>
                <small>{{ node.type === 'folder' ? '文件夹' : node.fileExt?.toUpperCase() || '文件' }}</small>
              </div>
            </div>
            <span class="materials-folder-row-meta">
              {{ node.type === 'folder' ? `${node.children?.length ?? 0} 项` : formatFileSize(node.fileSize ?? 0) }}
            </span>
          </article>
        </div>
        <div v-else class="empty-state small">
          <p class="muted-copy">当前文件夹下还没有内容。</p>
        </div>
      </div>

      <div v-else class="materials-preview-body">
        <div v-if="loadingBlob" class="empty-state small">
          <p class="muted-copy">文件内容加载中。</p>
        </div>

        <pre v-else-if="previewText" class="materials-text-preview">{{ previewText }}</pre>

        <iframe
          v-else-if="canRenderInFrame"
          class="materials-file-frame"
          :src="previewUrl"
          title="materials-preview"
        />

        <div v-else class="empty-state small">
          <p class="muted-copy">当前文件暂不支持页内渲染，请使用下载或新窗口预览。</p>
        </div>
      </div>
    </template>

    <div v-else class="empty-state small">
      <p class="muted-copy">从右侧资料栏选择一个文件或文件夹。</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MaterialDetailVO, MaterialTreeNodeVO } from '@/types/material'

const props = defineProps<{
  detail: MaterialDetailVO | null
  canManage: boolean
  loadingBlob: boolean
  previewUrl: string
  previewText: string
  previewMimeType: string
  tree: MaterialTreeNodeVO[]
}>()

defineEmits<{
  rename: []
  remove: []
  preview: []
  download: []
}>()

const folderChildren = computed(() => {
  if (!props.detail || props.detail.nodeType !== 'folder') return []
  return findNodeById(props.tree, props.detail.id)?.children ?? []
})

const canRenderInFrame = computed(() => {
  if (!props.previewUrl) return false
  if (!props.previewMimeType) return true
  return (
    props.previewMimeType.startsWith('image/') ||
    props.previewMimeType === 'application/pdf' ||
    props.previewMimeType.startsWith('video/') ||
    props.previewMimeType.startsWith('audio/')
  )
})

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function findNodeById(nodes: MaterialTreeNodeVO[], id: number): MaterialTreeNodeVO | null {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children?.length) {
      const found = findNodeById(node.children, id)
      if (found) return found
    }
  }
  return null
}
</script>
