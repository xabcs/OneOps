<script setup lang="ts">
  import { computed } from 'vue';
  import { useAuthStore } from '@/store/modules/auth';

  interface Props {
    permission: string | string[];
    mode?: 'hidden' | 'disabled' | 'request' | 'placeholder';
    icon?: string;
    tooltip?: string;
    placeholder?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    mode: 'hidden'
  });

  const emit = defineEmits<{
    click: [event: MouseEvent];
    requestPermission: [];
  }>();

  const authStore = useAuthStore();

  const hasPermission = computed(() => {
    if (typeof props.permission === 'string') {
      return authStore.hasPermission(props.permission);
    }
    return authStore.hasAnyPermission(props.permission);
  });

  const handleClick = (event: MouseEvent) => {
    if (hasPermission.value) {
      emit('click', event);
    }
  };

  const handleRequestPermission = () => {
    emit('requestPermission');
  };
</script>

<template>
  <div class="smart-permission-button">
    <!-- 有权限：显示按钮 -->
    <ElButton v-if="hasPermission" v-bind="$attrs" @click="handleClick">
      <slot name="icon">
        <ElIcon v-if="icon"><component :is="icon" /></ElIcon>
      </slot>
      <slot></slot>
    </ElButton>

    <!-- 无权限：根据模式显示不同内容 -->
    <template v-else>
      <!-- 模式1：完全隐藏 -->
      <div v-if="mode === 'hidden'" class="hidden-content"></div>

      <!-- 模式2：禁用 + 提示 -->
      <ElTooltip v-else-if="mode === 'disabled'" :content="tooltip || '您没有此操作权限'" placement="top">
        <ElButton v-bind="$attrs" disabled class="permission-disabled">
          <slot name="icon">
            <ElIcon v-if="icon"><component :is="icon" /></ElIcon>
          </slot>
          <slot></slot>
        </ElButton>
      </ElTooltip>

      <!-- 模式3：显示申请按钮 -->
      <div v-else-if="mode === 'request'" class="permission-request">
        <ElText type="info" size="small">
          <ElIcon><Lock /></ElIcon>
          需要权限
        </ElText>
        <ElButton type="text" size="small" @click="handleRequestPermission">申请</ElButton>
      </div>

      <!-- 模式4：显示提示信息 -->
      <div v-else-if="mode === 'placeholder'" class="permission-placeholder">
        <ElText type="info" size="small">
          <ElIcon><Lock /></ElIcon>
          {{ placeholder || '暂无权限' }}
        </ElText>
      </div>
    </template>
  </div>
</template>

<style scoped>
  .permission-disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .permission-request {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background-color: #f5f7fa;
    border-radius: 4px;
  }

  .permission-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 32px;
    color: #909399;
  }

  .hidden-content {
    display: none;
  }
</style>
