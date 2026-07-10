<template>
  <div class="materials-toolbar-shell">
    <button v-if="canManage" class="button ghost compact" @click="$emit('create-folder')">新建文件夹</button>
    <button v-if="canManage" class="button primary compact" @click="$emit('upload')">上传文件</button>
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
