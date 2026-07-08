<template>
  <article class="card materials-card">
    <div class="section-head section-head-top">
      <div>
        <p class="eyebrow">Materials</p>
        <h3>资料中心</h3>
      </div>
      <button class="button ghost compact" @click="reload" :disabled="loadingTree">
        {{ loadingTree ? '刷新中' : '刷新资料' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="!tree.length && !loadingTree" class="muted-copy">当前课程还没有资料。</p>

    <div class="materials-layout">
      <section class="materials-tree-panel">
        <MaterialTree :nodes="tree" :selected-id="selectedNodeId" @select="selectNode" />
      </section>

      <section class="materials-detail-panel panel">
        <template v-if="selectedDetail">
          <p class="eyebrow">Detail</p>
          <h4 class="material-title">{{ selectedDetail.nodeName }}</h4>
          <div class="material-detail-grid">
            <div>
              <p class="label">类型</p>
              <p class="value minor">{{ selectedDetail.nodeType === 'folder' ? '文件夹' : '文件' }}</p>
            </div>
            <div>
              <p class="label">大小</p>
              <p class="value minor">{{ formatFileSize(selectedDetail.fileSize) }}</p>
            </div>
            <div>
              <p class="label">版本</p>
              <p class="value minor">v{{ selectedDetail.latestVersionNo }}</p>
            </div>
            <div>
              <p class="label">更新时间</p>
              <p class="value minor">{{ formatDateTime(selectedDetail.updatedAt) }}</p>
            </div>
          </div>

          <div v-if="selectedDetail.nodeType === 'file'" class="inline-actions top-gap">
            <button class="button ghost compact" @click="handlePreview" :disabled="loadingBlob">预览</button>
            <button class="button primary compact" @click="handleDownload" :disabled="loadingBlob">下载</button>
          </div>
        </template>

        <div v-else class="empty-state small">
          <p class="muted-copy">从左侧选择一个资料节点查看详情。</p>
        </div>
      </section>
    </div>
  </article>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import MaterialTree from '@/components/MaterialTree.vue'
import { downloadMaterial, getMaterialDetail, getMaterialTree, previewMaterial } from '@/api/material'
import type { MaterialDetailVO, MaterialTreeNodeVO } from '@/types/material'
import { formatDateTime } from '@/utils/date'

const props = defineProps<{
  courseId: number
  token: string
}>()

const tree = ref<MaterialTreeNodeVO[]>([])
const selectedNodeId = ref<number | null>(null)
const selectedDetail = ref<MaterialDetailVO | null>(null)
const loadingTree = ref(false)
const loadingBlob = ref(false)
const errorMessage = ref('')

async function loadTree() {
  loadingTree.value = true
  errorMessage.value = ''
  try {
    tree.value = await getMaterialTree(props.token, props.courseId)
    if (!tree.value.length) {
      selectedNodeId.value = null
      selectedDetail.value = null
      return
    }
    if (!selectedNodeId.value) {
      const first = findFirstNode(tree.value)
      if (first) {
        await selectNode(first)
      }
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料列表加载失败'
  } finally {
    loadingTree.value = false
  }
}

async function selectNode(node: MaterialTreeNodeVO) {
  selectedNodeId.value = node.id
  try {
    selectedDetail.value = await getMaterialDetail(props.token, props.courseId, node.id)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料详情加载失败'
  }
}

async function handlePreview() {
  if (!selectedDetail.value) return
  loadingBlob.value = true
  try {
    const blob = await previewMaterial(props.token, props.courseId, selectedDetail.value.id)
    const url = URL.createObjectURL(blob)
    window.open(url, '_blank', 'noopener,noreferrer')
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料预览失败'
  } finally {
    loadingBlob.value = false
  }
}

async function handleDownload() {
  if (!selectedDetail.value) return
  loadingBlob.value = true
  try {
    const blob = await downloadMaterial(props.token, props.courseId, selectedDetail.value.id)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = selectedDetail.value.nodeName
    anchor.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料下载失败'
  } finally {
    loadingBlob.value = false
  }
}

async function reload() {
  await loadTree()
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function findFirstNode(nodes: MaterialTreeNodeVO[]): MaterialTreeNodeVO | null {
  for (const node of nodes) {
    return node
  }
  return null
}

onMounted(async () => {
  await loadTree()
})
</script>
