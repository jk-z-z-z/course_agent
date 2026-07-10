<template>
  <div class="app-shell" :class="{ 'with-sidebar': hasSidebar }">
    <header v-if="$slots.top" class="app-shell-top">
      <slot name="top" />
    </header>

    <aside v-if="hasSidebar" class="app-shell-side">
      <slot name="side" />
    </aside>

    <main class="app-shell-main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  hasSidebar?: boolean
}>()
</script>
