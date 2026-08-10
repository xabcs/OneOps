<script setup lang="ts">
import { computed } from 'vue';
import { useForm, useFormRules } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({ name: 'AuthUserSearch' });

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useForm();

const model = defineModel<{
  username: string;
  nickname: string;
  email: string;
  phone: string;
}>('model', { required: true });

type RuleKey = Extract<'email', keyof typeof model.value>;

const rules = computed<Record<RuleKey, App.Global.FormRule>>(() => {
  const { patternRules } = useFormRules();

  return {
    email: patternRules.email
  };
});

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
      <ElCollapseItem :title="$t('common.search')" name="auth-user-search">
        <ElForm ref="formRef" :model="model" :rules="rules" label-position="right" :label-width="80">
          <ElRow :gutter="24">
            <ElCol :lg="6" :md="8" :sm="12">
              <ElFormItem label="用户名" prop="username">
                <ElInput v-model="model.username" placeholder="请输入用户名" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="6" :md="8" :sm="12">
              <ElFormItem label="昵称" prop="nickname">
                <ElInput v-model="model.nickname" placeholder="请输入昵称" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="6" :md="8" :sm="12">
              <ElFormItem label="邮箱" prop="email">
                <ElInput v-model="model.email" placeholder="请输入邮箱" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="6" :md="8" :sm="12">
              <ElFormItem label="电话" prop="phone">
                <ElInput v-model="model.phone" placeholder="请输入电话" clearable />
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
