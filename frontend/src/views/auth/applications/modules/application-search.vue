<script setup lang="ts">
import { computed } from 'vue';
import { useForm } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({ name: 'AuthApplicationSearch' });

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useForm();

const model = defineModel<{
  name: string;
  code: string;
  type: string;
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
      <ElCollapseItem :title="$t('common.search')" name="auth-application-search">
        <ElForm ref="formRef" :model="model" label-position="right" :label-width="80">
          <ElRow :gutter="24">
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="应用名称" prop="name">
                <ElInput v-model="model.name" placeholder="请输入应用名称" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="应用编码" prop="code">
                <ElInput v-model="model.code" placeholder="请输入应用编码" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem label="应用类型" prop="type">
                <ElSelect v-model="model.type" placeholder="请选择应用类型" clearable class="w-full">
                  <ElOption label="Jenkins" value="jenkins" />
                  <ElOption label="Jumpserver" value="jumpserver" />
                  <ElOption label="GitLab" value="gitlab" />
                </ElSelect>
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
