<template>
  <div class="basic-layout">
    <header class="basic-layout-header">
      <PlatformTopNav variant="dashboard" />
    </header>

    <main class="basic-layout-content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import PlatformTopNav from '@/components/PlatformTopNav.vue'
</script>
