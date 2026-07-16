<template>
  <div class="info-card">
    <div v-if="$slots.header || title" class="card-header">
      <slot name="header">
        <h3 class="card-title">{{ title }}</h3>
      </slot>
    </div>

    <div class="card-body">
      <slot />
    </div>

    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string
}

withDefaults(defineProps<Props>(), {
  title: ''
})
</script>

<style scoped>
.info-card {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.info-card:hover {
  border-color: var(--el-border-color-hover);
}

.card-header {
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color);
  background: var(--el-fill-color-light);
}

.card-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.card-body {
  padding: 16px;
}

.card-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color);
  background: var(--el-fill-color-light);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
