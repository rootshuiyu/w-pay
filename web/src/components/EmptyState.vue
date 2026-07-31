<template>
  <div class="empty-state">
    <el-empty :image-size="imageSize" :description="description">
      <template v-if="title" #description>
        <p class="empty-title">{{ title }}</p>
        <p v-if="description" class="empty-desc">{{ description }}</p>
      </template>
      <slot />
    </el-empty>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  imageSize: { type: Number, default: 88 },
})
</script>

<style scoped>
.empty-state { padding: 24px 0; }
.empty-title { font-size: 15px; font-weight: 600; color: var(--text-primary); margin-bottom: 6px; }
.empty-desc { font-size: 13px; color: var(--text-secondary); }
</style>
