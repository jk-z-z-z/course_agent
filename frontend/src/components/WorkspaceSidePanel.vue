<template>
  <aside class="workspace-side-panel" :class="[panelClass, { collapsed }]">
    <div class="workspace-side-toolbar">
      <button
        class="workspace-side-toggle"
        type="button"
        :class="{ collapsed }"
        :aria-label="collapsed ? expandLabel : collapseLabel"
        :title="collapsed ? expandLabel : collapseLabel"
        @click="$emit('update:collapsed', !collapsed)"
      >
        <span class="workspace-side-toggle-arrow" aria-hidden="true">{{ collapsed ? '←' : '→' }}</span>
      </button>

      <template v-if="!collapsed">
        <strong class="workspace-side-title">{{ title }}</strong>
        <slot name="actions" />
      </template>
    </div>

    <section v-if="!collapsed" class="workspace-side-list" :class="contentClass">
      <slot />
    </section>
  </aside>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    collapsed: boolean
    expandLabel?: string
    collapseLabel?: string
    panelClass?: string
    contentClass?: string
  }>(),
  {
    expandLabel: '展开侧栏',
    collapseLabel: '收起侧栏',
    panelClass: '',
    contentClass: '',
  },
)

defineEmits<{
  'update:collapsed': [value: boolean]
}>()
</script>
