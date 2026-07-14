<template>
  <div class="tree-branch">
    <div v-for="node in nodes" :key="node.id" class="tree-item">
      <button
        class="tree-node"
        :class="{ active: node.id === selectedId }"
        type="button"
        @click="handleNodeClick(node)"
      >
        <span
          v-if="node.type === 'folder'"
          class="tree-node-caret"
          :class="{ expanded: isExpanded(node) }"
          aria-hidden="true"
        />
        <span v-else class="tree-node-caret placeholder" aria-hidden="true" />
        <span class="tree-node-label">
          <span class="tree-node-icon" :class="node.type === 'folder' ? 'folder' : 'file'" aria-hidden="true" />
          <span class="tree-node-name">
            <strong>{{ node.name }}</strong>
          </span>
        </span>
      </button>

      <button
        v-if="canManage"
        class="tree-node-delete"
        type="button"
        title="删除"
        @click.stop="$emit('remove', node)"
      >
        ×
      </button>

      <div v-if="node.children?.length && isExpanded(node)" class="tree-children">
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
import { ref, watch } from 'vue'
import type { MaterialTreeNodeVO } from '@/types/material'

defineOptions({ name: 'MaterialTree' })

const props = defineProps<{
  nodes: MaterialTreeNodeVO[]
  selectedId?: number | null
  canManage: boolean
}>()

const emit = defineEmits<{
  select: [node: MaterialTreeNodeVO]
  remove: [node: MaterialTreeNodeVO]
}>()

const expandedIds = ref<Set<number>>(new Set())

watch(
  () => props.nodes,
  (nodes) => {
    const next = new Set(expandedIds.value)
    addFolderIds(nodes, next)
    expandedIds.value = next
  },
  { immediate: true },
)

function handleNodeClick(node: MaterialTreeNodeVO) {
  emit('select', node)
  if (node.type !== 'folder' || !node.children?.length) return
  const next = new Set(expandedIds.value)
  if (next.has(node.id)) {
    next.delete(node.id)
  } else {
    next.add(node.id)
  }
  expandedIds.value = next
}

function isExpanded(node: MaterialTreeNodeVO) {
  return node.type === 'folder' && expandedIds.value.has(node.id)
}

function addFolderIds(nodes: MaterialTreeNodeVO[], target: Set<number>) {
  for (const node of nodes) {
    if (node.type !== 'folder') continue
    target.add(node.id)
    if (node.children?.length) {
      addFolderIds(node.children, target)
    }
  }
}
</script>
