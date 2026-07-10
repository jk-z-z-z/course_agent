<template>
  <AppShell>
    <template #top>
      <PlatformTopNav />
    </template>

    <section class="page-section">
      <div class="page-hero course-list-hero">
        <div>
          <p class="eyebrow">Dashboard</p>
          <h1>课程列表</h1>
          <p class="lead">课程页骨架已切出，下一步会把筛选、创建课程和卡片列表完整迁入。</p>
        </div>
      </div>

      <article class="content-card">
        <div class="section-head">
          <div>
            <p class="eyebrow">Courses</p>
            <h2>我的课程</h2>
          </div>
        </div>
        <p class="muted-copy">课程列表页已独立，后续会迁入现有课程数据和创建逻辑。</p>
      </article>
    </section>
  </AppShell>
</template>

<script setup lang="ts">
import AppShell from '@/components/AppShell.vue'
import PlatformTopNav from '@/components/PlatformTopNav.vue'
</script>
