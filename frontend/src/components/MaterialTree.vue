<template>
  <div class="tree-branch">
    <div v-for="node in nodes" :key="node.id" class="tree-item">
      <button
        class="tree-node"
        :class="{ active: node.id === selectedId }"
        @click="$emit('select', node)"
      >
        <span class="tree-node-label">
          <span class="tree-node-icon">{{ node.type === 'folder' ? '📁' : '📄' }}</span>
          <span>{{ node.name }}</span>
        </span>
        <span class="tree-node-meta">{{ node.type === 'file' ? formatFileSize(node.fileSize ?? 0) : `${node.children?.length ?? 0} 项` }}</span>
      </button>

      <div v-if="node.children?.length" class="tree-children">
        <MaterialTree
          :nodes="node.children"
          :selected-id="selectedId"
          @select="$emit('select', $event)"
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
}>()

defineEmits<{
  select: [node: MaterialTreeNodeVO]
}>()

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
</script>
