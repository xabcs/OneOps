<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import {
  fetchCreateServerTag,
  fetchDeleteServerTag,
  fetchGetServerTags,
  fetchUpdateServerTag
} from '@/service/api';
import { ElMessageBox, ElNotification, type FormInstance, type FormRules } from 'element-plus';

defineOptions({ name: 'CmdbTags' });

const loading = ref(false);
const tableData = ref<CMDB.ServerTag[]>([]);
const dialogVisible = ref(false);
const dialogTitle = ref('');
const formRef = ref<FormInstance>();

const form = reactive<CMDB.ServerTagForm>({
  name: '',
  color: '#409EFF',
  description: '',
  sortOrder: 0,
  status: 1
});

const rules: FormRules = {
  name: [{ required: true, message: '请输入标签名称', trigger: 'blur' }],
  color: [{ required: true, message: '请选择标签颜色', trigger: 'change' }]
};

async function getTableData() {
  loading.value = true;
  try {
    const { data } = await fetchGetServerTags();
    tableData.value = data || [];
  } catch (error) {
    console.error('获取标签列表失败:', error);
    ElNotification.error('获取标签列表失败');
  } finally {
    loading.value = false;
  }
}

function handleAdd() {
  dialogTitle.value = '新增标签';
  Object.assign(form, {
    id: undefined,
    name: '',
    color: '#409EFF',
    description: '',
    sortOrder: 0,
    status: 1
  });
  dialogVisible.value = true;
}

function handleEdit(row: CMDB.ServerTag) {
  dialogTitle.value = '编辑标签';
  Object.assign(form, {
    id: row.id,
    name: row.name,
    color: row.color,
    description: row.description || '',
    sortOrder: row.sortOrder,
    status: row.status
  });
  dialogVisible.value = true;
}

async function handleSave() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    if (form.id) {
      await fetchUpdateServerTag(form.id, form);
      ElNotification.success('标签更新成功');
    } else {
      await fetchCreateServerTag(form);
      ElNotification.success('标签创建成功');
    }
    dialogVisible.value = false;
    getTableData();
  } catch (error) {
    console.error('保存标签失败:', error);
    ElNotification.error('保存标签失败');
  }
}

function handleDelete(row: CMDB.ServerTag) {
  ElMessageBox.confirm(`确定要删除标签 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await fetchDeleteServerTag(row.id);
      ElNotification.success('删除成功');
      getTableData();
    })
    .catch(() => {});
}

onMounted(() => {
  getTableData();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <div class="mb-16px flex justify-between">
        <ElButton @click="getTableData">刷新</ElButton>
        <ElButton type="primary" @click="handleAdd">新增标签</ElButton>
      </div>

      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn prop="name" label="标签名称" min-width="150" />
        <ElTableColumn label="颜色" width="100">
          <template #default="{ row }">
            <span
              v-if="row.color"
              :style="{ backgroundColor: row.color }"
              class="inline-block w-40px h-22px rounded-4px align-middle"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="sortOrder" label="排序" width="80" />
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
        <ElFormItem label="标签名称" prop="name">
          <ElInput v-model="form.name" placeholder="请输入标签名称" />
        </ElFormItem>
        <ElFormItem label="颜色" prop="color">
          <ElColorPicker v-model="form.color" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="form.sortOrder" :min="0" style="width: 100%" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElRadioGroup v-model="form.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入标签描述" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
