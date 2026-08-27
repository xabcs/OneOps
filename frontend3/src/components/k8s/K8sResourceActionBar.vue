<script setup lang="ts">
    import { useRouter } from 'vue-router';
    import { ElButton, ElTag, ElTooltip, ElMessage } from 'element-plus';
    import { useAuthStore } from '@/store/modules/auth';

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
    /** 禁用（如缺少操作权限），置灰显示 */
    disabled?: boolean;
    /** 操作权限码：无权限时置灰 + tooltip/点击提示缺失的权限码 */
    permission?: string;
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
    const authStore = useAuthStore();

    // 返回上一页
    const handleBack = () => {
      // 使用 router.push 而不是 router.back()，确保列表页能重新加载
      router.push(props.backPath);
    };

    /** 是否因缺少权限而置灰 */
    const missingPermission = (action: ActionItem): boolean =>
      !!action.permission && !authStore.hasPermission(action.permission);

    const isDisabled = (action: ActionItem): boolean =>
      !!action.disabled || missingPermission(action);

    const actionTooltip = (action: ActionItem): string =>
      missingPermission(action)
        ? `缺少权限：${action.permission}\n请联系管理员在角色管理中开通`
        : action.tooltip || '';

    // 置灰按钮不响应 click（disabled），由外层 span 承接点击并提示缺失权限
    const handleWrapClick = (action: ActionItem) => {
      if (missingPermission(action)) {
        ElMessage({
          type: 'warning',
          message: `缺少权限：${action.permission}，请联系管理员在角色管理中开通`,
          grouping: true,
          showClose: true
        });
      }
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
                <span class="action-wrap" @click="handleWrapClick(action)">
                    <ElTooltip v-if="actionTooltip(action)" :content="actionTooltip(action)" placement="top">
                        <ElButton :type="action.type" size="small" :disabled="isDisabled(action)" @click="action.handler">
                            {{ action.label }}
                        </ElButton>
                    </ElTooltip>
                    <ElButton v-else :type="action.type" size="small" :disabled="isDisabled(action)" @click="action.handler">
                        {{ action.label }}
                    </ElButton>
                </span>
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

    .action-wrap {
      display: inline-flex;
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
