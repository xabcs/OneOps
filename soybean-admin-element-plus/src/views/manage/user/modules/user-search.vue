<script setup lang="ts">
import { computed } from 'vue';
import { useForm, useFormRules } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({ name: 'UserSearch' });

interface Emits {
  (e: 'reset'): void;
  (e: 'search'): void;
}

const emit = defineEmits<Emits>();

const { formRef, validate, restoreValidation } = useForm();

const model = defineModel<Api.SystemManage.UserSearchParams>('model', { required: true });

type RuleKey = Extract<keyof Api.SystemManage.UserSearchParams, 'email'>;

const rules = computed<Record<RuleKey, App.Global.FormRule>>(() => {
  const { patternRules } = useFormRules(); // inside computed to make locale reactive

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
  <ElCard class="card-wrapper">
    <ElCollapse>
      <ElCollapseItem :title="$t('common.search')" name="user-search">
        <ElForm ref="formRef" :model="model" :rules="rules" label-position="right" :label-width="80">
          <ElRow :gutter="24">
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem :label="$t('page.manage.user.userName')" prop="username">
                <ElInput v-model="model.username" :placeholder="$t('page.manage.user.form.userName')" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem :label="$t('page.manage.user.nickName')" prop="nickname">
                <ElInput v-model="model.nickname" :placeholder="$t('page.manage.user.form.nickName')" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="8" :md="12" :sm="24">
              <ElFormItem :label="$t('page.manage.user.userEmail')" prop="email">
                <ElInput v-model="model.email" :placeholder="$t('page.manage.user.form.userEmail')" clearable />
              </ElFormItem>
            </ElCol>
            <ElCol :lg="12" :md="24" :sm="24">
              <ElSpace class="w-full justify-end" alignment="end">
                <ElButton @click="reset">
                  <template #icon>
                    <icon-ic-round-refresh class="text-icon" />
                  </template>
                  {{ $t('common.reset') }}
                </ElButton>
                <ElButton type="primary" plain @click="search">
                  <template #icon>
                    <icon-ic-round-search class="text-icon" />
                  </template>
                  {{ $t('common.search') }}
                </ElButton>
              </ElSpace>
            </ElCol>
          </ElRow>
        </ElForm>
      </ElCollapseItem>
    </ElCollapse>
  </ElCard>
</template>

<style scoped></style>
