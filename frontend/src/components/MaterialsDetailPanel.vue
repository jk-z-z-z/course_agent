<template>
  <section class="materials-detail-panel panel materials-detail-shell">
    <template v-if="detail">
      <div class="section-head materials-preview-head">
        <div>
          <h4 class="material-title">{{ detail.nodeName }}</h4>
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

        <div v-else-if="csvRows.length" class="materials-table-preview-wrap">
          <table class="materials-table-preview">
            <tbody>
              <tr v-for="(row, rowIndex) in csvRows" :key="rowIndex">
                <td v-for="(cell, cellIndex) in row" :key="cellIndex">{{ cell }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <pre v-else-if="previewText" class="materials-text-preview">{{ previewText }}</pre>

        <img
          v-else-if="previewKind === 'image'"
          class="materials-image-preview"
          :src="previewUrl"
          :alt="detail.nodeName"
        />

        <video
          v-else-if="previewKind === 'video'"
          class="materials-media-preview"
          :src="previewUrl"
          controls
        />

        <audio
          v-else-if="previewKind === 'audio'"
          class="materials-audio-preview"
          :src="previewUrl"
          controls
        />

        <iframe
          v-else-if="previewKind === 'pdf' || previewKind === 'html'"
          class="materials-file-frame"
          :src="previewUrl"
          :sandbox="previewKind === 'html' ? '' : undefined"
          title="materials-preview"
        />

        <div v-else-if="previewKind === 'office'" class="materials-preview-fallback">
          <strong>{{ fileExtLabel }} 文件</strong>
          <p class="muted-copy">当前文件需要下载后使用本地应用查看。</p>
          <button class="button ghost compact" type="button" @click="$emit('download', detail)">下载文件</button>
        </div>

        <div v-else class="empty-state small">
          <p class="muted-copy">当前文件暂不支持页内渲染。</p>
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
  loadingBlob: boolean
  previewUrl: string
  previewText: string
  previewMimeType: string
  tree: MaterialTreeNodeVO[]
  canManage: boolean
}>()

defineEmits<{
  rename: [detail: MaterialDetailVO]
  download: [detail: MaterialDetailVO]
}>()

const folderChildren = computed(() => {
  if (!props.detail || props.detail.nodeType !== 'folder') return []
  return findNodeById(props.tree, props.detail.id)?.children ?? []
})

const previewKind = computed(() => {
  if (!props.previewUrl) return false
  const mime = normalizeMimeType(props.previewMimeType)
  const ext = fileExt.value
  if (mime.startsWith('image/') || ext === 'svg') return 'image'
  if (mime === 'application/pdf' || ext === 'pdf') return 'pdf'
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  if (mime === 'text/html' || ext === 'html' || ext === 'htm') return 'html'
  if (isOfficeExt(ext)) return 'office'
  return 'unsupported'
})

const fileExt = computed(() => (props.detail?.fileExt ?? '').trim().replace(/^\./, '').toLowerCase())

const fileExtLabel = computed(() => {
  const ext = fileExt.value
  return ext ? ext.toUpperCase() : '此类'
})

const csvRows = computed(() => {
  if (!props.previewText) return []
  const mime = normalizeMimeType(props.previewMimeType)
  if (fileExt.value !== 'csv' && mime !== 'text/csv') return []
  return parseCsv(props.previewText).slice(0, 100).map((row) => row.slice(0, 20))
})

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function normalizeMimeType(mimeType?: string) {
  return (mimeType ?? '').split(';')[0].trim().toLowerCase()
}

function isOfficeExt(ext: string) {
  return ['doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx'].includes(ext)
}

function parseCsv(text: string) {
  const rows: string[][] = []
  let row: string[] = []
  let cell = ''
  let inQuotes = false

  for (let index = 0; index < text.length; index++) {
    const char = text[index]
    const next = text[index + 1]
    if (char === '"' && inQuotes && next === '"') {
      cell += '"'
      index++
      continue
    }
    if (char === '"') {
      inQuotes = !inQuotes
      continue
    }
    if (char === ',' && !inQuotes) {
      row.push(cell)
      cell = ''
      continue
    }
    if ((char === '\n' || char === '\r') && !inQuotes) {
      if (char === '\r' && next === '\n') index++
      row.push(cell)
      rows.push(row)
      row = []
      cell = ''
      continue
    }
    cell += char
  }

  if (cell || row.length) {
    row.push(cell)
    rows.push(row)
  }
  return rows.filter((cells) => cells.some((value) => value.trim() !== ''))
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
