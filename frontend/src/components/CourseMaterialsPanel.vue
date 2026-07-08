<template>
  <article class="card materials-card">
    <div class="section-head section-head-top">
      <div>
        <p class="eyebrow">Materials</p>
        <h3>资料中心</h3>
      </div>
      <div class="inline-actions">
        <button class="button ghost compact" @click="reload" :disabled="loadingTree">{{ loadingTree ? '刷新中' : '刷新资料' }}</button>
        <button v-if="canManage" class="button ghost compact" @click="createFolder">新建文件夹</button>
        <button v-if="canManage" class="button primary compact" @click="openUploader">上传文件</button>
      </div>
    </div>

    <input ref="fileInputRef" class="hidden-input" type="file" @change="handleUpload" />

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-else-if="!tree.length && !loadingTree" class="muted-copy">当前课程还没有资料。</p>

    <div class="materials-layout">
      <section class="materials-tree-panel">
        <MaterialTree :nodes="tree" :selected-id="selectedNodeId" @select="selectNode" />
      </section>

      <section class="materials-detail-panel panel">
        <template v-if="selectedDetail">
          <div class="section-head">
            <div>
              <p class="eyebrow">Detail</p>
              <h4 class="material-title">{{ selectedDetail.nodeName }}</h4>
            </div>
            <div v-if="canManage" class="inline-actions">
              <button class="button ghost compact" @click="renameNode">重命名</button>
              <button class="button danger compact" @click="removeNode">删除</button>
            </div>
          </div>

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
import {
  createMaterialFolder,
  deleteMaterialNode,
  downloadMaterial,
  getMaterialDetail,
  getMaterialTree,
  previewMaterial,
  updateMaterialNode,
  uploadMaterialFile,
} from '@/api/material'
import type { MaterialDetailVO, MaterialTreeNodeVO } from '@/types/material'
import { formatDateTime } from '@/utils/date'

const props = defineProps<{
  courseId: number
  token: string
  canManage: boolean
}>()

const tree = ref<MaterialTreeNodeVO[]>([])
const selectedNodeId = ref<number | null>(null)
const selectedDetail = ref<MaterialDetailVO | null>(null)
const loadingTree = ref(false)
const loadingBlob = ref(false)
const errorMessage = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

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
    if (!selectedNodeId.value || !findNodeById(tree.value, selectedNodeId.value)) {
      const first = findFirstNode(tree.value)
      if (first) {
        await selectNode(first)
      }
      return
    }
    const current = findNodeById(tree.value, selectedNodeId.value)
    if (current) {
      await selectNode(current)
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

async function createFolder() {
  if (!props.canManage) return
  const folderName = window.prompt('输入文件夹名称')?.trim() ?? ''
  if (!folderName) return
  try {
    const created = await createMaterialFolder(props.token, props.courseId, {
      parentId: currentFolderTargetId(),
      folderName,
    })
    await loadTree()
    const createdNode = findNodeById(tree.value, created.id)
    if (createdNode) await selectNode(createdNode)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '文件夹创建失败'
  }
}

function openUploader() {
  fileInputRef.value?.click()
}

async function handleUpload(event: Event) {
  if (!props.canManage) return
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const created = await uploadMaterialFile(props.token, props.courseId, file, currentFolderTargetId())
    input.value = ''
    await loadTree()
    const createdNode = findNodeById(tree.value, created.id)
    if (createdNode) await selectNode(createdNode)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '文件上传失败'
  }
}

async function renameNode() {
  if (!props.canManage || !selectedDetail.value) return
  const nextName = window.prompt('输入新的名称', selectedDetail.value.nodeName)?.trim() ?? ''
  if (!nextName || nextName === selectedDetail.value.nodeName) return
  try {
    await updateMaterialNode(props.token, props.courseId, selectedDetail.value.id, { nodeName: nextName })
    await loadTree()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料重命名失败'
  }
}

async function removeNode() {
  if (!props.canManage || !selectedDetail.value) return
  const confirmed = window.confirm(`确认删除“${selectedDetail.value.nodeName}”吗？`)
  if (!confirmed) return
  try {
    await deleteMaterialNode(props.token, props.courseId, selectedDetail.value.id)
    selectedNodeId.value = null
    selectedDetail.value = null
    await loadTree()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料删除失败'
  }
}

async function reload() {
  await loadTree()
}

function currentFolderTargetId() {
  if (!selectedDetail.value) return undefined
  if (selectedDetail.value.nodeType === 'folder') return selectedDetail.value.id
  return selectedDetail.value.parentId
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

onMounted(async () => {
  await loadTree()
})
</script>
