<script setup lang="ts">
  interface Props {
    /** 按钮变体 */
    variant?: 'light' | 'white' | 'brand' | 'blue';
    /** 是否禁用 */
    disabled?: boolean;
  }

  defineProps<Props>();

  const emit = defineEmits<{
    (e: 'click', event: MouseEvent): void;
  }>();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      emit('click', event);
    }
  };
</script>

<template>
  <button
    class="gradient-button"
    :class="[`gradient-button-${variant}`, { 'gradient-button-disabled': disabled }]"
    :disabled="disabled"
    @click="handleClick"
  >
    <span v-if="$slots.icon" class="gradient-button-icon">
      <slot name="icon" />
    </span>
    <span class="gradient-button-content">
      <slot />
    </span>
  </button>
</template>

<style scoped>
  .gradient-button {
    border-radius: 10px;
    padding: 8px 16px;
    cursor: pointer;
    transition: all 0.3s ease;
    border: 1px solid rgba(96, 165, 250, 0.24);
    box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
    font-size: 14px;
  }

  .gradient-button-light {
    background: linear-gradient(135deg, #eff6ffeb, #ecfdf5bd);
    color: #3b82f6;
  }

  .gradient-button-light:hover {
    background: linear-gradient(135deg, #dbeafefa, #d1fae5d6);
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgba(59, 130, 246, 0.2);
  }

  .gradient-button-white {
    background: linear-gradient(135deg, #eff6ffeb, #fff7edb8);
    color: #3b82f6;
  }

  .gradient-button-white:hover {
    background: linear-gradient(135deg, #dbeafef5, #ffedd5d1);
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgba(59, 130, 246, 0.2);
  }

  .gradient-button-brand {
    background: linear-gradient(135deg, #67b7ab 0%, #5586b6 58%, #49639a 100%);
    color: white;
    border-color: rgba(103, 183, 171, 0.3);
  }

  .gradient-button-brand:hover {
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgba(103, 183, 171, 0.3);
  }

  .gradient-button-blue {
    background: linear-gradient(135deg, #3b82f6, #38bdf8);
    color: white;
    border-color: rgba(59, 130, 246, 0.3);
  }

  .gradient-button-blue:hover {
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgba(59, 130, 246, 0.3);
  }

  .gradient-button:disabled {
    opacity: 0.58;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
  }

  .gradient-button-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
  }

  .gradient-button-content {
    display: inline-flex;
    align-items: center;
  }

  @media (max-width: 768px) {
    .gradient-button {
      border-radius: 8px;
      padding: 6px 12px;
      font-size: 13px;
    }
  }
</style>
