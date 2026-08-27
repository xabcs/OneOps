<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { Delete, Plus } from '@element-plus/icons-vue';
  import { createTicketType, updateTicketType } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'TicketTypeOperateDialog' });

  interface Props {
    /** operation type: add | edit */
    operateType: UI.TableOperateType;
    /** editing row data */
    rowData?: Api.Ticket.TicketType | null;
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

  interface BaseModel {
    name: string;
    code: string;
    icon: string;
    description: string;
    status: number;
  }

  const baseModel = ref<BaseModel>({
    name: '',
    code: '',
    icon: '',
    description: '',
    status: 1
  });

  /** 表单设计器字段模型（编辑态，options 用换行文本） */
  interface FieldModel {
    key: string;
    label: string;
    type: Api.Ticket.FormFieldType;
    required: boolean;
    optionsText: string;
    placeholder: string;
  }

  const fields = ref<FieldModel[]>([]);

  const rules: Record<'name' | 'code', App.Global.FormRule> = {
    name: defaultRequiredRule,
    code: defaultRequiredRule
  };

  const fieldTypeOptions: { label: string; value: Api.Ticket.FormFieldType }[] = [
    { label: '单行文本', value: 'input' },
    { label: '多行文本', value: 'textarea' },
    { label: '数字', value: 'number' },
    { label: '下拉选择', value: 'select' },
    { label: '日期时间', value: 'date' }
  ];

  function createField(): FieldModel {
    return { key: '', label: '', type: 'input', required: false, optionsText: '', placeholder: '' };
  }

  function handleInitModel() {
    baseModel.value = { name: '', code: '', icon: '', description: '', status: 1 };
    fields.value = [createField()];

    if (props.operateType === 'edit' && props.rowData) {
      Object.assign(baseModel.value, {
        name: props.rowData.name,
        code: props.rowData.code,
        icon: props.rowData.icon,
        description: props.rowData.description,
        status: props.rowData.status
      });

      try {
        const parsed = props.rowData.formSchema ? JSON.parse(props.rowData.formSchema) : [];
        if (Array.isArray(parsed) && parsed.length > 0) {
          fields.value = parsed.map((f: Api.Ticket.FormField) => ({
            key: f.key ?? '',
            label: f.label ?? '',
            type: f.type ?? 'input',
            required: Boolean(f.required),
            optionsText: Array.isArray(f.options) ? f.options.join('\n') : '',
            placeholder: f.placeholder ?? ''
          }));
        }
      } catch {
        fields.value = [createField()];
      }
    }
  }

  function addField() {
    fields.value.push(createField());
  }

  function removeField(index: number) {
    fields.value.splice(index, 1);
  }

  function moveField(index: number, dir: -1 | 1) {
    const target = index + dir;
    if (target < 0 || target >= fields.value.length) return;
    const list = fields.value;
    [list[index], list[target]] = [list[target], list[index]];
  }

  /** 校验字段配置 */
  function validateFields(): string | null {
    const keys = new Set<string>();
    for (let i = 0; i < fields.value.length; i += 1) {
      const field = fields.value[i];
      if (!field.key.trim() || !field.label.trim()) {
        return `第 ${i + 1} 个字段需填写字段名与显示名`;
      }
      if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(field.key.trim())) {
        return `字段名「${field.key}」仅支持字母开头的字母/数字/下划线`;
      }
      if (keys.has(field.key.trim())) {
        return `字段名「${field.key}」重复`;
      }
      keys.add(field.key.trim());
      if (field.type === 'select' && !field.optionsText.trim()) {
        return `下拉字段「${field.label}」需填写候选项`;
      }
    }
    return null;
  }

  /** 生成 formSchema JSON */
  function buildFormSchema(): string {
    const schema = fields.value.map(field => {
      const item: Api.Ticket.FormField = {
        key: field.key.trim(),
        label: field.label.trim(),
        type: field.type,
        required: field.required
      };
      if (field.type === 'select') {
        item.options = field.optionsText
          .split('\n')
          .map(s => s.trim())
          .filter(Boolean);
      }
      if (field.placeholder.trim()) {
        item.placeholder = field.placeholder.trim();
      }
      return item;
    });
    return JSON.stringify(schema);
  }

  const submitLoading = ref(false);

  async function handleSubmit() {
    await validate();

    const fieldError = validateFields();
    if (fieldError) {
      ElMessage.warning(fieldError);
      return;
    }

    const payload = {
      ...baseModel.value,
      formSchema: buildFormSchema()
    };

    submitLoading.value = true;
    const { error } =
      props.operateType === 'edit'
        ? await updateTicketType(props.rowData!.id, payload)
        : await createTicketType(payload);
    submitLoading.value = false;

    if (!error) {
      ElMessage.success(props.operateType === 'edit' ? '更新成功' : '创建成功');
      visible.value = false;
      emit('submitted');
    }
  }

  const title = computed(() => (props.operateType === 'add' ? '新建工单类型' : '编辑工单类型'));

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
    }
  });
</script>

<template>
  <ElDialog v-model="visible" :title="title" :width="720" top="6vh" destroy-on-close>
    <ElForm ref="formRef" :model="baseModel" :rules="rules" label-width="100px">
      <ElFormItem label="类型名称" prop="name">
        <ElInput v-model="baseModel.name" placeholder="如：SQL 审核" maxlength="50" />
      </ElFormItem>
      <ElFormItem label="类型编码" prop="code">
        <ElInput v-model="baseModel.code" placeholder="如：sql-audit" :disabled="operateType === 'edit'" maxlength="50" />
      </ElFormItem>
      <ElFormItem label="图标" prop="icon">
        <ElInput v-model="baseModel.icon" placeholder="mdi 图标名，如 mdi:database" maxlength="60" class="w-260px" />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="baseModel.description" type="textarea" :rows="2" placeholder="类型用途说明（发起页展示）" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="baseModel.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
    </ElForm>

    <ElDivider content-position="left">表单字段设计（发起工单时渲染）</ElDivider>

    <div class="flex flex-col gap-12px">
      <div v-for="(field, index) in fields" :key="index" class="rounded-6px border border-gray-200 p-12px">
        <div class="mb-8px flex items-center justify-between">
          <ElTag size="small" type="primary">字段 {{ index + 1 }}</ElTag>
          <div class="flex gap-4px">
            <ElButton size="small" text :disabled="index === 0" @click="moveField(index, -1)">上移</ElButton>
            <ElButton size="small" text :disabled="index === fields.length - 1" @click="moveField(index, 1)">下移</ElButton>
            <ElButton size="small" text type="danger" :icon="Delete" @click="removeField(index)" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-x-12px gap-y-8px">
          <ElFormItem label="字段名" label-width="70px" required>
            <ElInput v-model="field.key" placeholder="英文标识 如 sql_content" />
          </ElFormItem>
          <ElFormItem label="显示名" label-width="70px" required>
            <ElInput v-model="field.label" placeholder="如 SQL 内容" />
          </ElFormItem>
          <ElFormItem label="组件" label-width="70px">
            <ElSelect v-model="field.type">
              <ElOption v-for="opt in fieldTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="占位提示" label-width="70px">
            <ElInput v-model="field.placeholder" placeholder="输入提示（可选）" />
          </ElFormItem>
          <ElFormItem v-if="field.type === 'select'" label="候选项" label-width="70px" class="col-span-2">
            <ElInput
              v-model="field.optionsText"
              type="textarea"
              :rows="3"
              placeholder="每行一个选项"
            />
          </ElFormItem>
          <ElFormItem label="必填" label-width="70px">
            <ElSwitch v-model="field.required" />
          </ElFormItem>
        </div>
      </div>

      <ElButton type="primary" plain :icon="Plus" @click="addField">添加字段</ElButton>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>
