<template>
  <el-card shadow="never" class="table-card" v-loading="loading">
    <template #header>
      <div class="card-header-content">
        <div class="header-left">
          <div style="display: flex; align-items: center; gap: 12px">
            <h2 class="page-title">{{ title }}</h2>
            <span class="accent-dot"></span>
          </div>
          <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
        </div>
        <div class="header-right">
          <el-space>
            <slot name="actions"></slot>
          </el-space>
        </div>
      </div>
    </template>

    <slot></slot>
  </el-card>
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
  },
  loading: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.table-card {
  border-radius: 12px;
  border: 1px solid var(--border, #e2e8f0);
}

.card-header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
}

.header-right {
  display: flex;
  align-items: center;
}
</style>
