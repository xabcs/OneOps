<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useForm, useFormRules } from '@/hooks/common/form';
import { fetchCreateRole, fetchUpdateRole } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'RoleOperateDrawer' });

interface Props {
  /** the type of operation */
  operateType: UI.TableOperateType;
  /** the edit row data */
  rowData?: Api.SystemManage.Role | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => {
  const titles: Record<UI.TableOperateType, string> = {
    add: $t('page.manage.role.addRole'),
    edit: $t('page.manage.role.editRole')
  };
  return titles[props.operateType];
});

type Model = Pick<Api.SystemManage.Role, 'name' | 'code' | 'description' | 'status'>;

const model = ref(createDefaultModel());

function createDefaultModel(): Model {
  return {
    name: '',
    code: '',
    description: '',
    status: 1
  };
}

type RuleKey = Exclude<keyof Model, 'description'>;

const rules = computed(() => {
  const baseRules: Record<string, App.Global.FormRule> = {
    name: defaultRequiredRule,
    status: defaultRequiredRule
  };

  // 只在新增模式下验证code字段
  if (!isEdit.value) {
    baseRules.code = defaultRequiredRule;
  }

  return baseRules;
});

const roleId = computed(() => props.rowData?.id || -1);

const isEdit = computed(() => props.operateType === 'edit');

function handleInitModel() {
  model.value = createDefaultModel();

  if (props.operateType === 'edit' && props.rowData) {
    Object.assign(model.value, props.rowData);
  }
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  // 准备提交数据
  const submitData = isEdit.value
    ? {
        name: model.value.name,
        description: model.value.description,
        status: model.value.status
      }
    : model.value;

  const { error } = isEdit.value
    ? await fetchUpdateRole(roleId.value, submitData)
    : await fetchCreateRole(submitData);

  if (!error) {
    window.$message?.success(isEdit.value ? $t('common.updateSuccess') : '添加成功');
    closeDrawer();
    emit('submitted');
  } else {
    window.$message?.error(isEdit.value ? '更新失败' : '添加失败');
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
      <ElFormItem :label="$t('page.manage.role.roleName')" prop="name">
        <ElInput v-model="model.name" :placeholder="$t('page.manage.role.form.roleName')" />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.role.roleCode')" prop="code">
        <ElInput v-model="model.code" :placeholder="$t('page.manage.role.form.roleCode')" :disabled="isEdit" />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.role.roleStatus')" prop="status">
        <ElRadioGroup v-model="model.status" :disabled="isEdit && model.code === 'admin'">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.role.roleDesc')" prop="description">
        <ElInput v-model="model.description" :placeholder="$t('page.manage.role.form.roleDesc')" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElSpace :size="16">
        <ElButton @click="closeDrawer">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</ElButton>
      </ElSpace>
    </template>
  </ElDrawer>
</template>

<style scoped></style>
