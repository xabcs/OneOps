<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchCreateUserGroup, fetchUpdateUserGroup } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';
  import { $t } from '@/locales';

  defineOptions({ name: 'UserGroupOperateDrawer' });

  interface Props {
    /** the type of operation */
    operateType: UI.TableOperateType;
    /** the edit row data */
    rowData?: Api.SystemManage.UserGroup | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', {
    default: false
  });

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '新增用户组' : '编辑用户组'));

  type Model = Pick<Api.SystemManage.UserGroup, 'code' | 'name' | 'description' | 'status'>;

  const model = ref(createDefaultModel());

  function createDefaultModel(): Model {
    return {
      code: '',
      name: '',
      description: '',
      status: 1
    };
  }

  const rules = computed(() => {
    const baseRules: Record<string, App.Global.FormRule> = {
      name: defaultRequiredRule,
      status: defaultRequiredRule
    };
    if (!isEdit.value) {
      baseRules.code = defaultRequiredRule;
    }
    return baseRules;
  });

  const groupId = computed(() => props.rowData?.id || -1);
  const isEdit = computed(() => props.operateType === 'edit');

  function handleInitModel() {
    model.value = createDefaultModel();

    if (props.operateType === 'edit' && props.rowData) {
      model.value = {
        code: props.rowData.code,
        name: props.rowData.name,
        description: props.rowData.description,
        status: props.rowData.status
      };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    const { error } = isEdit.value
      ? await fetchUpdateUserGroup(groupId.value, {
          name: model.value.name,
          description: model.value.description,
          status: model.value.status
        })
      : await fetchCreateUserGroup({
          code: model.value.code,
          name: model.value.name,
          description: model.value.description
        });

    if (!error) {
      window.$message?.success(isEdit.value ? $t('common.updateSuccess') : '添加成功');
      closeDrawer();
      emit('submitted');
    }
  }

  watch(visible, () => {
    if (visible.value) {
      handleInitModel();
      restoreValidation();
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :size="360">
    <ElForm ref="formRef" :model="model" :rules="rules" label-position="top">
      <ElFormItem label="用户组编码" prop="code">
        <ElInput v-model="model.code" placeholder="唯一编码，如 sre-team" :disabled="isEdit" />
      </ElFormItem>
      <ElFormItem label="用户组名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入用户组名称" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="model.description" placeholder="请输入描述" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElSpace :size="16">
        <ElButton @click="closeDrawer">{{ $t('common.cancel') }}</ElButton>
        <PermissionButton
          :code="isEdit ? 'system.user.update' : 'system.user.create'"
          type="primary"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </PermissionButton>
      </ElSpace>
    </template>
  </ElDrawer>
</template>

<style scoped></style>
