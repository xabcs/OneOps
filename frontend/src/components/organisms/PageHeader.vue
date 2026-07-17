<script setup lang="ts">
interface Props {
  title?: string;
  description?: string;
}

withDefaults(defineProps<Props>(), {
  title: '',
  description: ''
});
</script>

<template>
  <div class="page-header">
    <div class="header-left">
      <slot name="title">
        <h2 class="page-title">{{ title }}</h2>
      </slot>

      <p v-if="description" class="page-description">
        {{ description }}
      </p>
    </div>

    <div class="header-right">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20px;
  background: var(--el-bg-color-page);
  border-radius: 4px;
  margin-bottom: 16px;
}

.header-left {
  flex: 1;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.page-description {
  margin: 8px 0 0 0;
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.header-right {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>
