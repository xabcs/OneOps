<script setup lang="ts">
  import { computed } from 'vue';
  import { ElButton, ElMessage } from 'element-plus';
  import type { ButtonProps } from 'element-plus';
  import { useAuthStore } from '@/store/modules/auth';

  /**
   * 权限按钮组件（置灰模式）
   *
   * 设计原则：无权限一律置灰显示，不做隐藏——用户能看到功能入口，
   * 通过 tooltip / 点击提示得知缺失的权限码，便于向管理员申请开通。
   *
   * ElButton 的属性（type/size/link/icon/loading/disabled 等）声明在 props 中，
   * 经 buttonAttrs 透传；@click 经 emits 转发——template 与 TSX 中均可通过类型检查。
   */
  interface Props {
    /** 单个权限码 */
    code?: string;
    /** 多个权限码，任一/全部满足（取决于 mode） */
    codes?: string[];
    /** 多权限码判定方式：any 任一满足即可，all 需全部满足 */
    mode?: 'any' | 'all';
  }

  const props = withDefaults(defineProps<Props & Partial<ButtonProps>>(), {
    mode: 'any'
  });

  const emit = defineEmits<{ click: [evt: MouseEvent] }>();

  const authStore = useAuthStore();

  /** 提示文案中的权限描述 */
  const permLabel = computed(() => props.code || (props.codes || []).join(' 或 ') || '所需权限');

  /** 是否缺少权限 */
  const missing = computed(() => {
    if (props.code) return !authStore.hasPermission(props.code);
    if (props.codes?.length) {
      return props.mode === 'all'
        ? !authStore.hasAllPermissions(props.codes)
        : !authStore.hasAnyPermission(props.codes);
    }
    return false;
  });

  const tip = computed(() => `缺少权限：${permLabel.value}\n请联系管理员在角色管理中开通`);

  /** 剔除权限字段后，其余（含 ElButton 属性）透传给 ElButton */
  const buttonAttrs = computed(() => {
    const { code, codes, mode, ...rest } = props;
    return rest;
  });

  // 原生 disabled button 不派发 click，由外层 span 承接点击并提示
  const notifyMissing = () => {
    ElMessage({
      type: 'warning',
      message: `缺少权限：${permLabel.value}，请联系管理员在角色管理中开通`,
      grouping: true,
      showClose: true
    });
  };
</script>

<template>
  <span v-if="missing" class="perm-btn-wrap" :title="tip" @click="notifyMissing">
    <ElButton v-bind="{ ...$attrs, ...buttonAttrs, disabled: true }">
      <slot />
    </ElButton>
  </span>
  <ElButton v-else v-bind="{ ...$attrs, ...buttonAttrs }" @click="emit('click', $event)">
    <slot />
  </ElButton>
</template>

<style scoped>
  .perm-btn-wrap {
    display: inline-flex;
    /* 与 .el-button 默认的 vertical-align 保持一致，行内环境下与相邻按钮对齐 */
    vertical-align: middle;
    cursor: not-allowed;
  }

  /* 原生 disabled button 不派发 click 也不冒泡，穿透到外层 span 才能触发提示 */
  .perm-btn-wrap :deep(.el-button.is-disabled) {
    pointer-events: none;
  }
</style>
