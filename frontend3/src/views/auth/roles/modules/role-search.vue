<script setup lang="ts">
  import { useForm } from '@/hooks/common/form';
  import { $t } from '@/locales';

  defineOptions({ name: 'AuthRoleSearch' });

  interface Emits {
    (e: 'reset'): void;
    (e: 'search'): void;
  }

  const emit = defineEmits<Emits>();

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();

  const model = defineModel<{
    name: string;
    code: string;
    description: string;
  }>('model', { required: true });

  async function reset() {
    await restoreValidation();
    emit('reset');
  }

  async function search() {
    await validate();
    emit('search');
  }
</script>

<template>
  <ElCard class="search-card-wrapper" shadow="never">
    <ElCollapse>
      <ElCollapseItem :title="$t('common.search')" name="auth-role-search">
        <ElForm ref="formRef" :model="model" label-position="right" :label-width="80">
          <ElRow :gutter="24">
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="角色名称" prop="name">
                <ElInput v-model="model.name" placeholder="请输入角色名称" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="角色编码" prop="code">
                <ElInput v-model="model.code" placeholder="请输入角色编码" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="描述" prop="description">
                <ElInput v-model="model.description" placeholder="请输入描述" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="12" :md="24" :sm="24">
              <ElSpace class="w-full justify-end" alignment="end">
                <ElButton @click="reset">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  重置
                </ElButton>
                <ElButton type="primary" plain @click="search">
                  <template #icon>
                    <icon-ic-round-search class="text-icon" />
                  </template>
                  搜索
                </ElButton>
              </ElSpace>
            </ElCol>
          </ElRow>
        </ElForm>
      </ElCollapseItem>
    </ElCollapse>
  </ElCard>
</template>

<style scoped lang="scss">
  /* 搜索卡片样式 - 无背景、无边框、无阴影 */
  :deep(.search-card-wrapper) {
    background-color: transparent !important;
    border: none !important;
    box-shadow: none !important;

    .el-card__body {
      padding: 0 !important;
      background-color: transparent !important;
    }
  }
</style>
