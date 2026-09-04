<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchAddPermission, fetchUpdatePermission } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'PermissionOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: Api.SystemManage.Permission | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', { default: false });

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => {
    const titles: Record<UI.TableOperateType, string> = {
      add: '新增权限',
      edit: '编辑权限'
    };
    return titles[props.operateType];
  });

  type Model = Pick<
    Api.SystemManage.Permission,
    'name' | 'code' | 'description' | 'module' | 'resource' | 'action' | 'level' | 'sortOrder' | 'status'
  >;

  function createDefaultModel(): Model {
    return {
      name: '',
      code: '',
      description: '',
      module: '',
      resource: '',
      action: '',
      level: 3,
      sortOrder: 0,
      status: 1
    };
  }

  const model = ref(createDefaultModel());

  const isEdit = computed(() => props.operateType === 'edit');

  const rules = computed(() => {
    const baseRules: Record<string, App.Global.FormRule> = {
      name: defaultRequiredRule,
      module: defaultRequiredRule,
      resource: defaultRequiredRule,
      action: defaultRequiredRule
    };
    if (!isEdit.value) {
      baseRules.code = defaultRequiredRule;
    }
    return baseRules;
  });

  const permissionId = computed(() => props.rowData?.id || -1);

  /** 权限级别选项 */
  const levelOptions = [
    { label: '模块', value: 1 },
    { label: '资源', value: 2 },
    { label: '操作', value: 3 }
  ];

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

    const { error } = isEdit.value
      ? await fetchUpdatePermission(permissionId.value, model.value)
      : await fetchAddPermission(model.value);

    if (!error) {
      window.$message?.success(isEdit.value ? '更新成功' : '添加成功');
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
  <ElDrawer v-model="visible" :title="title" :size="460">
    <ElForm ref="formRef" :model="model" :rules="rules" label-position="top">
      <ElFormItem label="权限名称" prop="name">
        <ElInput v-model="model.name" placeholder="如：创建用户" />
      </ElFormItem>
      <ElFormItem label="权限编码" prop="code">
        <ElInput v-model="model.code" placeholder="如：system.user.create" :disabled="isEdit" />
      </ElFormItem>
      <ElFormItem label="所属模块" prop="module">
        <ElInput v-model="model.module" placeholder="如：system" />
      </ElFormItem>
      <ElFormItem label="资源名称" prop="resource">
        <ElInput v-model="model.resource" placeholder="如：user" />
      </ElFormItem>
      <ElFormItem label="操作名称" prop="action">
        <ElInput v-model="model.action" placeholder="如：create" />
      </ElFormItem>
      <ElFormItem label="权限级别" prop="level">
        <ElRadioGroup v-model="model.level">
          <ElRadio v-for="opt in levelOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="排序" prop="sortOrder">
        <ElInputNumber v-model="model.sortOrder" :min="0" style="width: 100%" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="model.description" type="textarea" :rows="2" placeholder="权限描述" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElSpace :size="16">
        <ElButton @click="closeDrawer">取消</ElButton>
        <PermissionButton
          :code="isEdit ? 'system.permission.update' : 'system.permission.create'"
          type="primary"
          @click="handleSubmit"
        >
          确认
        </PermissionButton>
      </ElSpace>
    </template>
  </ElDrawer>
</template>

<style scoped></style>
