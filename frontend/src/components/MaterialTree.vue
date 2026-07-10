<template>
  <div class="tree-branch">
    <div v-for="node in nodes" :key="node.id" class="tree-item">
      <button
        class="tree-node"
        :class="{ active: node.id === selectedId }"
        @click="$emit('select', node)"
      >
        <span class="tree-node-label">
          <span class="tree-node-icon" :class="node.type === 'folder' ? 'folder' : 'file'">
            {{ node.type === 'folder' ? 'F' : 'D' }}
          </span>
          <span class="tree-node-name">
            <strong>{{ node.name }}</strong>
            <small>{{ node.type === 'file' ? '文件节点' : '目录节点' }}</small>
          </span>
        </span>
        <span class="tree-node-meta">{{ node.type === 'file' ? formatFileSize(node.fileSize ?? 0) : `${node.children?.length ?? 0} 项` }}</span>
      </button>

      <button
        v-if="canManage"
        class="tree-node-delete"
        @click.stop="$emit('remove', node)"
      >
        删除
      </button>

      <div v-if="node.children?.length" class="tree-children">
        <MaterialTree
          :nodes="node.children"
          :selected-id="selectedId"
          :can-manage="canManage"
          @select="$emit('select', $event)"
          @remove="$emit('remove', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MaterialTreeNodeVO } from '@/types/material'

defineOptions({ name: 'MaterialTree' })

defineProps<{
  nodes: MaterialTreeNodeVO[]
  selectedId?: number | null
  canManage: boolean
}>()

defineEmits<{
  select: [node: MaterialTreeNodeVO]
  remove: [node: MaterialTreeNodeVO]
}>()

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
</script>
