<template>
  <div class="materials-toolbar-shell">
    <div>
      <p class="eyebrow">Materials</p>
      <h3>资料中心</h3>
    </div>

    <div class="inline-actions">
      <button class="button ghost compact" @click="$emit('reload')" :disabled="loading">
        {{ loading ? '刷新中' : '刷新资料' }}
      </button>
      <button v-if="canManage" class="button ghost compact" @click="$emit('create-folder')">新建文件夹</button>
      <button v-if="canManage" class="button primary compact" @click="$emit('upload')">上传文件</button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean
  canManage: boolean
}>()

defineEmits<{
  reload: []
  'create-folder': []
  upload: []
}>()
</script>
