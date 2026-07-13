<template>
  <div class="materials-toolbar-shell">
    <button v-if="canManage" class="workspace-side-action" @click="$emit('create-folder')">新建</button>
    <button v-if="canManage" class="workspace-side-action primary" @click="$emit('upload')">上传</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  canManage: boolean
}>()

defineEmits<{
  'create-folder': []
  upload: []
}>()
</script>
