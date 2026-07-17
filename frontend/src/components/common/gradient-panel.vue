<script setup lang="ts">
interface Props {
  /** 面板标题 */
  title?: string;
  /** 面板变体 */
  variant?: 'light' | 'brand' | 'toolbar' | 'glass';
  /** 自定义样式 */
  customStyle?: Record<string, string>;
}

defineProps<Props>();
</script>

<template>
  <div class="gradient-panel" :class="[`gradient-panel-${variant}`]" :style="customStyle">
    <div v-if="$slots.header || title" class="gradient-panel-header">
      <slot name="header">
        <h3 class="gradient-panel-title">{{ title }}</h3>
      </slot>
      <div v-if="$slots.extra" class="gradient-panel-extra">
        <slot name="extra" />
      </div>
    </div>
    <div class="gradient-panel-body">
      <slot />
    </div>
    <div v-if="$slots.footer" class="gradient-panel-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<style scoped>
.gradient-panel {
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.gradient-panel-light {
  background: linear-gradient(135deg, #fbfdff, #f7faff 52%, #f9fbfd);
  border: 1px solid rgba(36, 91, 219, 0.09);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
  padding: 16px;
}

.gradient-panel-brand {
  background: linear-gradient(135deg, #67b7ab 0%, #5586b6 58%, #49639a 100%);
  color: white;
  padding: 20px;
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(103, 183, 171, 0.25);
}

.gradient-panel-toolbar {
  background: linear-gradient(180deg, #f8fafceb, #fffffff5);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  padding: 12px;
}

.gradient-panel-glass {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.6), transparent 52%);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 12px;
  padding: 16px;
}

.gradient-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
}

.gradient-panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.gradient-panel-brand .gradient-panel-title {
  color: white;
}

.gradient-panel-extra {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gradient-panel-body {
  flex: 1;
}

.gradient-panel-footer {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
}

@media (max-width: 768px) {
  .gradient-panel {
    border-radius: 8px;
    padding: 12px;
  }

  .gradient-panel-brand {
    padding: 16px;
  }
}
</style>
