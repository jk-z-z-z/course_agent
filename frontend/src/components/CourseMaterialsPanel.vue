<template>
  <article class="materials-card">
    <input ref="fileInputRef" class="hidden-input" type="file" @change="handleUpload" />

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="materials-shell">
      <section class="materials-main-panel">
        <p v-if="!tree.length && !loadingTree" class="muted-copy">当前课程还没有资料。</p>

        <MaterialsDetailPanel
          :detail="selectedDetail"
          :can-manage="canManage"
          :loading-blob="loadingBlob"
          @rename="renameNode"
          @remove="removeNode"
          @preview="handlePreview"
          @download="handleDownload"
        />
      </section>

      <aside class="materials-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <button class="materials-sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed">
          {{ sidebarCollapsed ? '展开资料栏' : '收起资料栏' }}
        </button>

        <div v-if="!sidebarCollapsed" class="materials-sidebar-body">
          <div class="materials-pane-head">
            <div>
              <p class="eyebrow">Folders</p>
              <h4 class="materials-pane-title">资料导航</h4>
            </div>
          </div>

          <MaterialsToolbar
            :can-manage="canManage"
            @create-folder="createFolder"
            @upload="openUploader"
          />

          <section class="materials-tree-panel">
            <MaterialTree :nodes="tree" :selected-id="selectedNodeId" @select="selectNode" />
          </section>
        </div>
      </aside>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import MaterialTree from '@/components/MaterialTree.vue'
import MaterialsDetailPanel from '@/components/MaterialsDetailPanel.vue'
import MaterialsToolbar from '@/components/MaterialsToolbar.vue'
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
const sidebarCollapsed = ref(false)

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
  errorMessage.value = ''
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

function currentFolderTargetId() {
  if (!selectedDetail.value) return undefined
  if (selectedDetail.value.nodeType === 'folder') return selectedDetail.value.id
  return selectedDetail.value.parentId
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

watch(
  () => [props.courseId, props.token],
  async () => {
    selectedNodeId.value = null
    selectedDetail.value = null
    errorMessage.value = ''
    await loadTree()
  },
  { immediate: true },
)
</script>
