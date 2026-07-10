<template>
  <article class="materials-card">
    <input ref="fileInputRef" class="hidden-input" type="file" @change="handleUpload" />

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="materials-shell" :class="{ collapsed: sidebarCollapsed }">
      <section class="materials-main-panel">
        <p v-if="!tree.length && !loadingTree" class="muted-copy">当前课程还没有资料。</p>

        <MaterialsDetailPanel
          :detail="selectedDetail"
          :can-manage="canManage"
          :loading-blob="loadingBlob"
          :preview-url="previewUrl"
          :preview-text="previewText"
          :preview-mime-type="previewMimeType"
          :tree="tree"
          @remove="removeNode"
        />
      </section>

      <button class="materials-sidebar-toggle" @click="sidebarCollapsed = !sidebarCollapsed">
        <span class="materials-sidebar-toggle-arrow">{{ sidebarCollapsed ? '‹' : '›' }}</span>
      </button>

      <aside class="materials-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="materials-sidebar-body" :class="{ hidden: sidebarCollapsed }">
          <div class="materials-pane-head">
            <div>
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
import { onBeforeUnmount, ref, watch } from 'vue'
import MaterialTree from '@/components/MaterialTree.vue'
import MaterialsDetailPanel from '@/components/MaterialsDetailPanel.vue'
import MaterialsToolbar from '@/components/MaterialsToolbar.vue'
import {
  createMaterialFolder,
  deleteMaterialNode,
  getMaterialDetail,
  getMaterialTree,
  previewMaterial,
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
const previewUrl = ref('')
const previewText = ref('')
const previewMimeType = ref('')

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
    await loadInlinePreview()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料详情加载失败'
  }
}

async function loadInlinePreview() {
  resetPreview()
  if (!selectedDetail.value || selectedDetail.value.nodeType !== 'file') return

  loadingBlob.value = true
  try {
    const blob = await previewMaterial(props.token, props.courseId, selectedDetail.value.id)
    previewMimeType.value = blob.type || selectedDetail.value.mimeType || ''

    if (previewMimeType.value.startsWith('text/') || isCodeLikeFile(selectedDetail.value.fileExt)) {
      previewText.value = await blob.text()
    } else {
      previewUrl.value = URL.createObjectURL(blob)
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料预览失败'
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

function resetPreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
  }
  previewUrl.value = ''
  previewText.value = ''
  previewMimeType.value = ''
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

function isCodeLikeFile(fileExt?: string) {
  if (!fileExt) return false
  return ['txt', 'md', 'json', 'js', 'ts', 'tsx', 'jsx', 'css', 'scss', 'html', 'xml', 'yml', 'yaml'].includes(
    fileExt.toLowerCase(),
  )
}

onBeforeUnmount(() => {
  resetPreview()
})

watch(
  () => [props.courseId, props.token],
  async () => {
    selectedNodeId.value = null
    selectedDetail.value = null
    errorMessage.value = ''
    resetPreview()
    await loadTree()
  },
  { immediate: true },
)
</script>
