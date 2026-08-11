<script setup lang="ts">
    import { computed, type Component } from 'vue';
    import { Loading } from '@element-plus/icons-vue';

    interface Props {
      type?: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text';
      size?: 'large' | 'default' | 'small';
      disabled?: boolean;
      loading?: boolean;
      icon?: Component;
      nativeType?: 'button' | 'submit' | 'reset';
      plain?: boolean;
      round?: boolean;
      circle?: boolean;
    }

    const props = withDefaults(defineProps<Props>(), {
      type: 'default',
      size: 'default',
      disabled: false,
      loading: false,
      nativeType: 'button',
      plain: false,
      round: false,
      circle: false
    });

    const emit = defineEmits<{
      click: [event: MouseEvent];
    }>();

    const buttonClass = computed(() => {
      return [
        'a-button',
        `a-button--${props.type}`,
        `a-button--${props.size}`,
        {
          'is-disabled': props.disabled,
          'is-loading': props.loading,
          'is-plain': props.plain,
          'is-round': props.round,
          'is-circle': props.circle
        }
      ];
    });

    const handleClick = (event: MouseEvent) => {
      if (!props.disabled && !props.loading) {
        emit('click', event);
      }
    };
</script>

<template>
    <button :type="nativeType" :class="buttonClass" :disabled="disabled || loading" @click="handleClick">
        <ElIcon v-if="loading" class="is-loading">
            <Loading />
        </ElIcon>
        <ElIcon v-else-if="icon">
            <component :is="icon" />
        </ElIcon>
        <span v-if="$slots.default" class="button-content">
            <slot />
        </span>
    </button>
</template>

<style scoped>
    .a-button {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      padding: 8px 16px;
      border: 1px solid transparent;
      border-radius: 4px;
      background: var(--el-bg-color);
      color: var(--el-text-color-primary);
      font-size: 14px;
      line-height: 1.5;
      cursor: pointer;
      transition: all 0.2s;
      user-select: none;
    }

    .a-button:hover {
      opacity: 0.8;
    }

    .a-button--primary {
      background: var(--el-color-primary);
      border-color: var(--el-color-primary);
      color: #fff;
    }

    .a-button--success {
      background: var(--el-color-success);
      border-color: var(--el-color-success);
      color: #fff;
    }

    .a-button--warning {
      background: var(--el-color-warning);
      border-color: var(--el-color-warning);
      color: #fff;
    }

    .a-button--danger {
      background: var(--el-color-danger);
      border-color: var(--el-color-danger);
      color: #fff;
    }

    .a-button--info {
      background: var(--el-color-info);
      border-color: var(--el-color-info);
      color: #fff;
    }

    .a-button--large {
      padding: 12px 20px;
      font-size: 16px;
    }

    .a-button--small {
      padding: 6px 12px;
      font-size: 12px;
    }

    .a-button.is-disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    .a-button.is-loading {
      pointer-events: none;
    }

    .a-button.is-plain {
      background: transparent;
    }

    .a-button.is-plain.a-button--primary {
      color: var(--el-color-primary);
      border-color: var(--el-color-primary);
    }

    .a-button.is-round {
      border-radius: 20px;
    }

    .a-button.is-circle {
      border-radius: 50%;
      padding: 8px;
    }

    .button-content {
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .is-loading {
      animation: rotating 1s linear infinite;
    }

    @keyframes rotating {
      from {
        transform: rotate(0deg);
      }
      to {
        transform: rotate(360deg);
      }
    }
</style>
