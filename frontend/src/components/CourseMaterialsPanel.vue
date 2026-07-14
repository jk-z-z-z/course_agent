<template>
  <article class="materials-card">
    <input ref="fileInputRef" class="hidden-input" type="file" @change="handleUpload" />

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="workspace-split-shell materials-shell" :class="{ 'is-collapsed': sidebarCollapsed }">
      <section class="materials-main-panel">
        <p v-if="!tree.length && !loadingTree" class="muted-copy">当前课程还没有资料。</p>

        <MaterialsDetailPanel
          :detail="selectedDetail"
          :loading-blob="loadingBlob"
          :preview-url="previewUrl"
          :preview-text="previewText"
          :preview-mime-type="previewMimeType"
          :tree="tree"
          :can-manage="canManage"
          @rename="renameNode"
          @download="downloadNode"
        />
      </section>

      <WorkspaceSidePanel
        title="资料"
        :collapsed="sidebarCollapsed"
        panel-class="materials-sidebar"
        content-class="materials-tree-panel"
        expand-label="展开资料列表"
        collapse-label="收起资料列表"
        @update:collapsed="sidebarCollapsed = $event"
      >
        <template #actions>
          <MaterialsToolbar
            :can-manage="canManage"
            @create-folder="createFolder"
            @upload="openUploader"
          />
        </template>

        <MaterialTree
          :nodes="tree"
          :selected-id="selectedNodeId"
          :can-manage="canManage"
          @select="selectNode"
          @remove="removeTreeNode"
        />
      </WorkspaceSidePanel>
    </div>

    <div v-if="folderDialogOpen" class="modal-backdrop" @click.self="closeFolderDialog">
      <section class="modal-card card materials-folder-modal">
        <div class="section-head">
          <div>
            <h3>新建文件夹</h3>
          </div>
          <button class="button ghost compact" type="button" @click="closeFolderDialog">关闭</button>
        </div>

        <form class="form" @submit.prevent="submitCreateFolder">
          <label class="field">
            <span>父文件夹</span>
            <select v-model="folderForm.parentId">
              <option value="">课程根目录</option>
              <option v-for="option in folderOptions" :key="option.id" :value="String(option.id)">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>文件夹名称</span>
            <input v-model.trim="folderForm.folderName" type="text" placeholder="输入文件夹名称" />
          </label>

          <p v-if="folderFormError" class="error">{{ folderFormError }}</p>

          <div class="inline-actions">
            <button class="button primary" type="submit">确认创建</button>
            <button class="button ghost" type="button" @click="closeFolderDialog">取消</button>
          </div>
        </form>
      </section>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import MaterialTree from '@/components/MaterialTree.vue'
import MaterialsDetailPanel from '@/components/MaterialsDetailPanel.vue'
import MaterialsToolbar from '@/components/MaterialsToolbar.vue'
import WorkspaceSidePanel from '@/components/WorkspaceSidePanel.vue'
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
const previewUrl = ref('')
const previewText = ref('')
const previewMimeType = ref('')
const folderDialogOpen = ref(false)
const folderFormError = ref('')
const folderForm = reactive({
  parentId: '',
  folderName: '',
})

const folderOptions = computed(() => flattenFolderOptions(tree.value))

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

    if (shouldReadAsText(selectedDetail.value.fileExt, previewMimeType.value)) {
      const rawText = await blob.text()
      previewText.value = formatPreviewText(rawText, selectedDetail.value.fileExt, previewMimeType.value)
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
  const currentFolderId = currentFolderTargetId()
  folderForm.parentId = typeof currentFolderId === 'number' ? String(currentFolderId) : ''
  folderForm.folderName = ''
  folderFormError.value = ''
  folderDialogOpen.value = true
}

async function submitCreateFolder() {
  if (!props.canManage) return
  const folderName = folderForm.folderName.trim()
  if (!folderName) {
    folderFormError.value = '请输入文件夹名称'
    return
  }
  const parentId = folderForm.parentId ? Number(folderForm.parentId) : undefined
  try {
    const created = await createMaterialFolder(props.token, props.courseId, {
      parentId,
      folderName,
    })
    closeFolderDialog()
    await loadTree()
    const createdNode = findNodeById(tree.value, created.id)
    if (createdNode) await selectNode(createdNode)
  } catch (error) {
    folderFormError.value = error instanceof Error ? error.message : '文件夹创建失败'
  }
}

function closeFolderDialog() {
  folderDialogOpen.value = false
  folderFormError.value = ''
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

async function removeTreeNode(node: MaterialTreeNodeVO) {
  if (!props.canManage) return
  const confirmed = window.confirm(`确认删除“${node.name}”吗？`)
  if (!confirmed) return
  try {
    await deleteMaterialNode(props.token, props.courseId, node.id)
    if (selectedNodeId.value === node.id) {
      selectedNodeId.value = null
      selectedDetail.value = null
      resetPreview()
    }
    await loadTree()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料删除失败'
  }
}

async function renameNode(detail: MaterialDetailVO) {
  if (!props.canManage) return
  const nodeName = window.prompt('输入新的名称', detail.nodeName)?.trim() ?? ''
  if (!nodeName || nodeName === detail.nodeName) return
  try {
    const updated = await updateMaterialNode(props.token, props.courseId, detail.id, {
      nodeName,
    })
    selectedDetail.value = updated
    await loadTree()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料重命名失败'
  }
}

async function downloadNode(detail: MaterialDetailVO) {
  if (detail.nodeType !== 'file') return
  try {
    const blob = await downloadMaterial(props.token, props.courseId, detail.id)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = detail.nodeName
    anchor.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '资料下载失败'
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

function flattenFolderOptions(nodes: MaterialTreeNodeVO[], depth = 0): Array<{ id: number; label: string }> {
  const result: Array<{ id: number; label: string }> = []
  for (const node of nodes) {
    if (node.type !== 'folder') continue
    result.push({
      id: node.id,
      label: `${'　'.repeat(depth)}${node.name}`,
    })
    if (node.children?.length) {
      result.push(...flattenFolderOptions(node.children, depth + 1))
    }
  }
  return result
}

function shouldReadAsText(fileExt?: string, mimeType?: string) {
  const ext = normalizeFileExt(fileExt)
  const mime = normalizeMimeType(mimeType)
  if (ext === 'html' || ext === 'htm' || ext === 'svg') return false
  if (mime === 'text/html' || mime === 'image/svg+xml') return false
  if (mime.startsWith('text/')) return true
  if (['application/json', 'application/xml', 'application/x-yaml', 'application/yaml'].includes(mime)) return true
  return [
    'txt',
    'log',
    'md',
    'markdown',
    'csv',
    'json',
    'js',
    'mjs',
    'ts',
    'tsx',
    'jsx',
    'css',
    'scss',
    'xml',
    'yml',
    'yaml',
    'go',
    'java',
    'py',
    'rs',
    'c',
    'cpp',
    'h',
    'hpp',
    'sql',
    'sh',
  ].includes(ext)
}

function formatPreviewText(text: string, fileExt?: string, mimeType?: string) {
  const ext = normalizeFileExt(fileExt)
  const mime = normalizeMimeType(mimeType)
  if (ext === 'json' || mime === 'application/json') {
    try {
      return JSON.stringify(JSON.parse(text), null, 2)
    } catch {
      return text
    }
  }
  return text
}

function normalizeFileExt(fileExt?: string) {
  return (fileExt ?? '').trim().replace(/^\./, '').toLowerCase()
}

function normalizeMimeType(mimeType?: string) {
  return (mimeType ?? '').split(';')[0].trim().toLowerCase()
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
