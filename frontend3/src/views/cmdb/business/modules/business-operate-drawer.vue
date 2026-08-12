<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchCreateBusinessUnit, fetchUpdateBusinessUnit } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'CmdbBusinessOperateDrawer' });

  interface Props {
    operateType: UI.TableOperateType;
    rowData?: CMDB.BusinessUnit | null;
    /** tree options for parent selection */
    treeOptions: Array<{ id: number; name: string; children?: CMDB.BusinessUnit[] }>;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '新增业务' : '编辑业务'));

  type Model = CMDB.BusinessUnitForm;

  function createDefaultModel(): Model {
    return {
      name: '',
      code: '',
      parentId: 0,
      owner: '',
      phone: '',
      email: '',
      sortOrder: 0,
      status: 1,
      remarks: ''
    };
  }

  const model = ref<Model>(createDefaultModel());

  const rules = {
    name: defaultRequiredRule,
    code: defaultRequiredRule
  };

  const treeSelectProps = { label: 'name', value: 'id', children: 'children' };

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      model.value = {
        id: props.rowData.id,
        name: props.rowData.name,
        code: props.rowData.code,
        parentId: props.rowData.parentId,
        owner: props.rowData.owner || '',
        phone: props.rowData.phone || '',
        email: props.rowData.email || '',
        sortOrder: props.rowData.sortOrder,
        status: props.rowData.status,
        remarks: props.rowData.remarks || ''
      };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (model.value.id) {
      const { error } = await fetchUpdateBusinessUnit(model.value.id, model.value);
      if (!error) {
        ElMessage.success('业务更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await fetchCreateBusinessUnit(model.value);
      if (!error) {
        ElMessage.success('业务创建成功');
        closeDrawer();
        emit('submitted');
      }
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="620">
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="100px">
      <ElFormItem label="业务名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入业务名称" />
      </ElFormItem>
      <ElFormItem label="业务代码" prop="code">
        <ElInput v-model="model.code" placeholder="请输入业务代码" />
      </ElFormItem>
      <ElFormItem label="上级业务">
        <ElTreeSelect
          v-model="model.parentId"
          :data="treeOptions"
          :props="treeSelectProps"
          clearable
          check-strictly
          class="w-full"
        />
      </ElFormItem>
      <ElFormItem label="负责人">
        <ElInput v-model="model.owner" placeholder="请输入负责人" />
      </ElFormItem>
      <ElFormItem label="联系电话">
        <ElInput v-model="model.phone" placeholder="请输入联系电话" />
      </ElFormItem>
      <ElFormItem label="邮箱">
        <ElInput v-model="model.email" placeholder="请输入邮箱" />
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber v-model="model.sortOrder" :min="0" class="w-full" />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="备注">
        <ElInput v-model="model.remarks" type="textarea" :rows="3" placeholder="请输入备注" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDrawer>
</template>
