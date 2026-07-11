<script setup lang="ts">
import { useRouter } from 'vue-router';
import {
  ElButton,
  ElTag,
  ElTooltip
} from 'element-plus';

interface MetaItem {
  label: string;
  value: string;
}

interface ActionItem {
  label: string;
  type?: '' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text';
  icon?: string;
  handler: () => void;
  tooltip?: string;
}

interface StatusTag {
  type: 'success' | 'warning' | 'danger' | 'info';
  text: string;
}

interface Props {
  /** 资源名称 */
  name: string;
  /** 命名空间 */
  namespace: string;
  /** 状态标签 */
  statusTag: StatusTag;
  /** 元信息（可选） */
  meta?: MetaItem[];
  /** 操作按钮列表 */
  actions?: ActionItem[];
  /** 返回路径（可选） */
  backPath?: string;
}

defineOptions({ name: 'K8sResourceActionBar' });

const props = withDefaults(defineProps<Props>(), {
  meta: () => [],
  actions: () => [],
  backPath: '/k8s/workloads'
});

const router = useRouter();

// 返回上一页
const handleBack = () => {
  console.log('[ActionBar] 点击返回按钮，backPath:', props.backPath);
  // 使用 router.push 而不是 router.back()，确保列表页能重新加载
  router.push(props.backPath);
};
</script>

<template>
  <div class="resource-action-bar">
    <!-- 主要信息行：返回箭头 + 名称 + 状态标签 -->
    <div class="action-bar-primary">
      <span class="back-arrow" @click="handleBack">←</span>
      <span class="resource-name">{{ name }}</span>
      <ElTag v-if="statusTag" :type="statusTag.type" size="small">{{ statusTag.text }}</ElTag>
    </div>

    <!-- 操作按钮组 -->
    <div class="action-bar-actions">
      <template v-for="action in actions" :key="action.label">
        <ElTooltip v-if="action.tooltip" :content="action.tooltip" placement="top">
          <ElButton :type="action.type" size="small" @click="action.handler">
            {{ action.label }}
          </ElButton>
        </ElTooltip>
        <ElButton v-else :type="action.type" size="small" @click="action.handler">
          {{ action.label }}
        </ElButton>
      </template>
    </div>
  </div>

  <!-- 元信息行：命名空间、副本、创建时间等（仅在有 meta 数据时显示） -->
  <div v-if="meta && meta.length > 0" class="action-bar-meta">
    <template v-for="(item, index) in meta" :key="index">
      <span class="meta-item">{{ item.label }}: {{ item.value }}</span>
      <span v-if="index < meta.length - 1" class="meta-divider">|</span>
    </template>
  </div>
</template>

<style scoped>
/* 顶部操作栏 - 透明背景，保持轻量 */
.resource-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.action-bar-primary {
  display: flex;
  align-items: center;
  gap: 12px;
}

.resource-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.back-arrow {
  font-size: 20px;
  color: #0052d9;
  cursor: pointer;
  transition: opacity 0.2s;
}

.back-arrow:hover {
  opacity: 0.8;
}

.action-bar-actions {
  display: flex;
  gap: 8px;
}

/* 元信息行 - 灰色小字 */
.action-bar-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #909399;
}

.meta-item {
  white-space: nowrap;
}

.meta-divider {
  color: #dcdfe6;
}
</style>
