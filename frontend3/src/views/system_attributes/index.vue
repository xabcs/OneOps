<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  fetchCreateAttribute,
  fetchDeleteAttribute,
  fetchGetAttributes,
  fetchUpdateAttribute
} from '@/service/api/system-manage';

defineOptions({
  name: 'SystemAttributes'
});

// 属性分类选项
const categoryOptions = [
  { label: '系统分类', value: 'system' },
  { label: '地理位置', value: 'location' },
  { label: '环境信息', value: 'environment' },
  { label: '硬件配置', value: 'hardware' },
  { label: '自定义', value: 'custom' }
];

// 属性类型选项
const typeOptions = [
  { label: '文本输入', value: 'text' },
  { label: '下拉单选', value: 'select' },
  { label: '下拉多选', value: 'multiselect' },
  { label: '数字', value: 'number' },
  { label: '日期', value: 'date' },
  { label: '布尔值', value: 'boolean' }
];

// 状态变量
const loading = ref(false);
const attributes = ref<System.AttributeDefinition[]>([]);
const selectedCategory = ref('');
const dialogVisible = ref(false);
const dialogMode = ref<'create' | 'edit'>('create');
const formRef = ref<FormInstance>();

// 表单数据
const formData = ref<System.AttributeDefinitionForm>({
  name: '',
  key: '',
  category: 'system',
  type: 'text',
  options: '',
  required: false,
  defaultValue: '',
  sortOrder: 0,
  description: ''
});

// 选项列表（用于select/multiselect类型）
const optionsList = ref<System.AttributeOption[]>([]);

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入属性名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  key: [
    { required: true, message: '请输入属性键', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: '只能包含小写字母、数字和下划线，且以字母开头', trigger: 'blur' }
  ],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
};

// 获取属性列表
async function getAttributes() {
  loading.value = true;
  try {
    const params = selectedCategory.value ? { category: selectedCategory.value } : {};
    const { data } = await fetchGetAttributes(params);
    attributes.value = data || [];
  } catch (error) {
    ElMessage.error('获取属性列表失败');
  } finally {
    loading.value = false;
  }
}

// 分类筛选
function handleCategoryChange() {
  getAttributes();
}

// 新增属性
function handleCreate() {
  dialogMode.value = 'create';
  Object.assign(formData.value, {
    name: '',
    key: '',
    category: 'system',
    type: 'text',
    options: '',
    required: false,
    defaultValue: '',
    sortOrder: 0,
    description: ''
  });
  optionsList.value = [];
  dialogVisible.value = true;
}

// 编辑属性
function handleEdit(row: System.AttributeDefinition) {
  dialogMode.value = 'edit';
  Object.assign(formData.value, {
    id: row.id,
    name: row.name,
    key: row.key,
    category: row.category,
    type: row.type,
    options: row.options,
    required: row.required,
    defaultValue: row.defaultValue,
    sortOrder: row.sortOrder,
    description: row.description
  });

  // 解析选项
  if (row.type === 'select' || row.type === 'multiselect') {
    try {
      optionsList.value = JSON.parse(row.options || '[]');
    } catch {
      optionsList.value = [];
    }
  } else {
    optionsList.value = [];
  }

  dialogVisible.value = true;
}

// 删除属性
async function handleDelete(row: System.AttributeDefinition) {
  try {
    await ElMessageBox.confirm(`确定要删除属性"${row.name}"吗？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
  } catch {
    return;
  }

  try {
    await fetchDeleteAttribute(row.id);
    ElMessage.success('删除成功');
    await getAttributes();
  } catch (error: unknown) {
    ElMessage.error((error as { message?: string }).message || '删除失败');
  }
}

// 类型变化处理
function handleTypeChange(type: System.AttributeType) {
  // 清空选项
  if (type !== 'select' && type !== 'multiselect') {
    optionsList.value = [];
    formData.value.options = '';
  }
}

// 添加选项
function handleAddOption() {
  optionsList.value.push({ label: '', value: '' });
}

// 删除选项
function handleRemoveOption(index: number) {
  optionsList.value.splice(index, 1);
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();

    // 如果是select/multiselect类型，生成选项JSON
    const data: System.AttributeDefinitionForm = { ...formData.value };
    if (data.type === 'select' || data.type === 'multiselect') {
      // 验证选项
      const validOptions = optionsList.value.filter(opt => opt.label && opt.value);
      if (validOptions.length === 0) {
        ElMessage.warning('请至少添加一个有效选项');
        return;
      }
      data.options = JSON.stringify(validOptions);
    }

    if (dialogMode.value === 'create') {
      await fetchCreateAttribute(data);
      ElMessage.success('创建成功');
    } else {
      await fetchUpdateAttribute(formData.value.id!, data);
      ElMessage.success('更新成功');
    }

    dialogVisible.value = false;
    await getAttributes();
  } catch (error: unknown) {
    if (error !== false) {
      const message = error instanceof Error ? error.message : '操作失败';
      ElMessage.error(message);
    }
  }
}

// 关闭对话框
function handleCloseDialog() {
  dialogVisible.value = false;
  formRef.value?.resetFields();
}

// 获取分类名称
function getCategoryName(category: System.AttributeCategory): string {
  const option = categoryOptions.find(opt => opt.value === category);
  return option?.label || category;
}

// 获取分类标签类型
function getCategoryTagType(category: System.AttributeCategory) {
  const typeMap: Record<System.AttributeCategory, '' | UI.ThemeColor> = {
    system: 'info',
    location: 'success',
    environment: 'warning',
    hardware: 'primary',
    custom: ''
  };
  return typeMap[category] || 'info';
}

// 获取类型名称
function getTypeName(type: System.AttributeType) {
  const option = typeOptions.find(opt => opt.value === type);
  return option?.label || type;
}

// 解析选项（用于显示）
function parseOptions(optionsStr: string): System.AttributeOption[] {
  if (!optionsStr) return [];
  try {
    return JSON.parse(optionsStr);
  } catch {
    return [];
  }
}

// 初始化
onMounted(() => {
  getAttributes();
});
</script>

<template>
  <div class="attributes-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">属性管理</span>
          <ElButton type="primary" @click="handleCreate">
            <icon-mdi-plus class="text-icon" />
            新增属性
          </ElButton>
        </div>
      </template>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <ElRadioGroup v-model="selectedCategory" @change="handleCategoryChange">
          <ElRadioButton label="">全部</ElRadioButton>
          <ElRadioButton label="system">系统分类</ElRadioButton>
          <ElRadioButton label="location">地理位置</ElRadioButton>
          <ElRadioButton label="environment">环境信息</ElRadioButton>
          <ElRadioButton label="hardware">硬件配置</ElRadioButton>
          <ElRadioButton label="custom">自定义</ElRadioButton>
        </ElRadioGroup>
      </div>

      <!-- 属性列表表格 -->
      <ElTable v-loading="loading" :data="attributes" stripe style="width: 100%">
        <ElTableColumn prop="name" label="属性名称" min-width="120" />
        <ElTableColumn prop="key" label="属性键" min-width="120">
          <template #default="{ row }">
            <ElTag type="info" size="small">{{ row.key }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="分类" width="100">
          <template #default="{ row }">
            <ElTag :type="getCategoryTagType(row.category)" size="small">
              {{ getCategoryName(row.category) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="类型" width="100">
          <template #default="{ row }">
            {{ getTypeName(row.type) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="选项预览" min-width="200">
          <template #default="{ row }">
            <template v-if="row.type === 'select' || row.type === 'multiselect'">
              <ElTag
                v-for="(opt, idx) in parseOptions(row.options)"
                :key="idx"
                type="info"
                size="small"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ opt.label }}
              </ElTag>
            </template>
            <span v-else class="text-gray-400">-</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="必填" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.required ? 'danger' : 'info'" size="small">
              {{ row.required ? '是' : '否' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="sortOrder" label="排序" width="80" align="center" />
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <!-- 属性表单对话框 -->
    <ElDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增属性' : '编辑属性'"
      width="600px"
      :close-on-click-modal="false"
      @close="handleCloseDialog"
    >
      <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
        <ElFormItem label="属性名称" prop="name">
          <ElInput v-model="formData.name" placeholder="如：业务系统、机房、机柜" maxlength="100" show-word-limit />
        </ElFormItem>

        <ElFormItem label="属性键" prop="key">
          <ElInput
            v-model="formData.key"
            placeholder="如：business_system、room、cabinet（英文，唯一）"
            :disabled="dialogMode === 'edit'"
            maxlength="50"
            show-word-limit
          />
          <div class="mt-4px text-12px text-gray-400">只能包含小写字母、数字和下划线，且以字母开头</div>
        </ElFormItem>

        <ElFormItem label="分类" prop="category">
          <ElSelect v-model="formData.category" placeholder="选择分类">
            <ElOption v-for="opt in categoryOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="类型" prop="type">
          <ElSelect v-model="formData.type" placeholder="选择类型" @change="handleTypeChange">
            <ElOption v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
        </ElFormItem>

        <!-- 下拉选择/多选类型的选项配置 -->
        <template v-if="formData.type === 'select' || formData.type === 'multiselect'">
          <ElFormItem label="选项配置">
            <div class="options-config">
              <div v-for="(opt, index) in optionsList" :key="index" class="option-item">
                <ElInput v-model="opt.label" placeholder="显示名称" style="width: 200px" />
                <span class="mx-2">=</span>
                <ElInput v-model="opt.value" placeholder="值" style="width: 150px" />
                <ElButton type="danger" size="small" class="ml-2" @click="handleRemoveOption(index)">删除</ElButton>
              </div>
              <ElButton type="primary" size="small" @click="handleAddOption">+ 添加选项</ElButton>
            </div>
            <div class="mt-4px text-12px text-gray-400">提示：用户选择时看到的是"显示名称"，实际存储的是"值"</div>
          </ElFormItem>
        </template>

        <ElFormItem label="默认值">
          <ElInput v-model="formData.defaultValue" placeholder="属性的默认值（可选）" maxlength="255" />
        </ElFormItem>

        <ElFormItem label="是否必填">
          <ElSwitch v-model="formData.required" />
          <span style="margin-left: 10px">
            {{ formData.required ? '是' : '否' }}
          </span>
        </ElFormItem>

        <ElFormItem label="排序">
          <ElInputNumber v-model="formData.sortOrder" :min="0" :max="9999" placeholder="数字越小越靠前" />
        </ElFormItem>

        <ElFormItem label="说明">
          <ElInput v-model="formData.description" type="textarea" :rows="3" placeholder="属性的详细说明（可选）" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="handleCloseDialog">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.attributes-page {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 500;
}

.filter-bar {
  margin-bottom: 16px;
}

.options-config {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mx-2 {
  margin-left: 8px;
  margin-right: 8px;
}

.ml-2 {
  margin-left: 8px;
}

.text-12px {
  font-size: 12px;
}

.text-gray-400 {
  color: #909399;
}

.mt-4px {
  margin-top: 4px;
}
</style>
