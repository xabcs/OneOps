<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue';
  import { fetchAddPermission, fetchUpdatePermission } from '@/service/api';
  import { useFormRules } from '@/hooks/common/form';

  defineOptions({
    name: 'PermissionOperateDrawer'
  });

  interface Props {
    visible: boolean;
    operateType: 'add' | 'edit';
    rowData?: Api.SystemManage.Permission | null;
  }

  interface Emits {
    (e: 'update:visible', visible: boolean): void;
    (e: 'submitted'): void;
  }

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  const visible = computed({
    get: () => props.visible,
    set: (val: boolean) => emit('update:visible', val)
  });

  const title = computed(() => {
    const titles: Record<typeof props.operateType, string> = {
      add: '新增权限',
      edit: '编辑权限'
    };
    return titles[props.operateType];
  });

  // 表单模型
  const model = reactive<Api.SystemManage.Permission>({
    id: undefined,
    name: '',
    code: '',
    description: '',
    module: 'system',
    resource: '',
    action: '',
    level: 3,
    parentId: null,
    sortOrder: 0,
    status: 1
  });

  // 表单验证规则
  const { formRules } = useFormRules(() => ({
    name: { required: true, message: '请输入权限名称', trigger: 'blur' },
    code: {
      required: true,
      message: '请输入权限编码',
      trigger: 'blur'
    },
    module: { required: true, message: '请选择模块', trigger: 'change' },
    resource: { required: true, message: '请输入资源名称', trigger: 'blur' },
    action: { required: true, message: '请输入操作名称', trigger: 'blur' },
    level: { required: true, message: '请选择权限级别', trigger: 'change' }
  }));

  const formRef = ref<InstanceType<typeof ElForm>>();

  // 权限级别选项
  const levelOptions = [
    { label: '模块级', value: 1 },
    { label: '页面级', value: 2 },
    { label: '按钮级', value: 3 },
    { label: 'API级', value: 4 }
  ];

  // 模块选项
  const moduleOptions = [
    { label: '系统管理', value: 'system' },
    { label: '授权中心', value: 'auth' },
    { label: '资产管理', value: 'cmdb' },
    { label: '监控中心', value: 'monitoring' },
    { label: 'K8s管理', value: 'k8s' }
  ];

  // 操作选项
  const actionOptions = [
    { label: '查看', value: 'view' },
    { label: '创建', value: 'create' },
    { label: '更新', value: 'update' },
    { label: '删除', value: 'delete' },
    { label: '管理', value: 'manage' },
    { label: '导出', value: 'export' },
    { label: '导入', value: 'import' },
    { label: '审核', value: 'audit' }
  ];

  // 生成权限编码预览
  const codePreview = computed(() => {
    if (!model.module || !model.resource || !model.action) {
      return '';
    }
    return `${model.module}.${model.resource}.${model.action}`;
  });

  // 监听操作类型和行数据变化
  watch(
    () => [props.operateType, props.rowData],
    ([type, rowData]) => {
      if (type === 'add') {
        Object.assign(model, {
          id: undefined,
          name: '',
          code: '',
          description: '',
          module: 'system',
          resource: '',
          action: '',
          level: 3,
          parentId: null,
          sortOrder: 0,
          status: 1
        });
      } else if (type === 'edit' && rowData) {
        Object.assign(model, rowData);
      }
    },
    { immediate: true }
  );

  // 提交表单
  async function handleSubmit() {
    const valid = await formRef.value?.validate().catch(() => false);
    if (!valid) return;

    const isEdit = props.operateType === 'edit';
    const data = { ...model };

    if (isEdit) {
      const { error } = await fetchUpdatePermission(data.id!, data);
      if (!error) {
        window.$message?.success('更新成功');
        visible.value = false;
        emit('submitted');
      } else {
        window.$message?.error(error.msg || '更新失败');
      }
    } else {
      const { error } = await fetchAddPermission(data);
      if (!error) {
        window.$message?.success('添加成功');
        visible.value = false;
        emit('submitted');
      } else {
        window.$message?.error(error.msg || '添加失败');
      }
    }
  }
</script>

<template>
  <ElDrawer v-model="visible" :title="title" size="600px" :close-on-click-modal="false" :close-on-press-escape="false">
    <ElForm ref="formRef" :model="model" :rules="formRules" label-position="top" label-width="80px">
      <ElFormItem label="权限名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入权限名称，如：查看用户" maxlength="50" show-word-limit />
      </ElFormItem>

      <ElFormItem label="权限编码" prop="code">
        <ElInput
          v-model="model.code"
          placeholder="自动生成或手动输入，如：system.user.view"
          maxlength="100"
          show-word-limit
        >
          <template #append>
            <ElButton v-if="codePreview" text @click="model.code = codePreview">使用预览</ElButton>
          </template>
        </ElInput>
        <div v-if="codePreview" class="code-preview">预览：{{ codePreview }}</div>
      </ElFormItem>

      <ElFormItem label="所属模块" prop="module">
        <ElSelect v-model="model.module" placeholder="请选择所属模块" style="width: 100%">
          <ElOption v-for="option in moduleOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="权限级别" prop="level">
        <ElSelect v-model="model.level" placeholder="请选择权限级别" style="width: 100%">
          <ElOption v-for="option in levelOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="资源名称" prop="resource">
        <ElInput v-model="model.resource" placeholder="请输入资源名称，如：user、role、menu" maxlength="50" />
      </ElFormItem>

      <ElFormItem label="操作名称" prop="action">
        <ElSelect v-model="model.action" placeholder="请选择或输入操作名称" style="width: 100%" filterable allow-create>
          <ElOption v-for="option in actionOptions" :key="option.value" :label="option.label" :value="option.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="权限描述">
        <ElInput
          v-model="model.description"
          type="textarea"
          placeholder="请输入权限描述"
          :rows="3"
          maxlength="200"
          show-word-limit
        />
      </ElFormItem>

      <ElFormItem label="排序">
        <ElInputNumber
          v-model="model.sortOrder"
          :min="0"
          :max="9999"
          placeholder="数字越小排序越靠前"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="状态">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      </div>
    </template>
  </ElDrawer>
</template>

<style scoped lang="scss">
  .drawer-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .code-preview {
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-color-primary);
    font-weight: 500;
  }

  :deep(.el-drawer__body) {
    padding: 20px;
  }
</style>
