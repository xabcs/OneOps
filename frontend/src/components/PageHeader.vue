<template>
  <div class="page-header">
    <div class="header-left">
      <div class="title-row">
        <h2 class="page-title">{{ title }}</h2>
        <span class="accent-dot"></span>
      </div>
      <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
    </div>
    <div class="header-right">
      <el-space :size="12">
        <slot name="actions"></slot>
      </el-space>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
  margin-bottom: 16px; /* 添加底部边距 */
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary, #1e293b);
}

.page-subtitle {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-secondary, #64748b);
}

.accent-dot {
  width: 4px;
  height: 4px;
  background-color: var(--accent, #2563eb);
  border-radius: 50%;
  box-shadow: 0 0 8px rgba(37, 99, 235, 0.5);
  flex-shrink: 0;
}

.header-right {
  display: flex;
  align-items: center;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-right {
    justify-content: flex-end;
  }
}
</style>
