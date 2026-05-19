<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import {
  fetchCreateBusinessUnit,
  fetchDeleteBusinessUnit,
  fetchGetBusinessUnits,
  fetchUpdateBusinessUnit
} from '@/service/api';
import { ElMessageBox, ElNotification, type FormInstance, type FormRules } from 'element-plus';

defineOptions({ name: 'CmdbBusiness' });

const loading = ref(false);
const tableData = ref<CMDB.BusinessUnit[]>([]);
const dialogVisible = ref(false);
const dialogTitle = ref('');
const formRef = ref<FormInstance>();

const form = reactive<CMDB.BusinessUnitForm>({
  name: '',
  code: '',
  parentId: 0,
  owner: '',
  phone: '',
  email: '',
  sortOrder: 0,
  status: 1,
  remarks: ''
});

const rules: FormRules = {
  name: [{ required: true, message: '请输入业务名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入业务代码', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '业务代码只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ]
};

const businessTreeOptions = computed(() => [{ id: 0, name: '根业务', children: tableData.value }]);
const treeSelectProps = { label: 'name', value: 'id', children: 'children' } as any;

async function getTableData() {
  loading.value = true;
  try {
    const { data } = await fetchGetBusinessUnits();
    tableData.value = data || [];
  } catch (error) {
    console.error('获取业务列表失败:', error);
    ElNotification.error('获取业务列表失败');
  } finally {
    loading.value = false;
  }
}

function handleAdd() {
  dialogTitle.value = '新增业务';
  Object.assign(form, {
    id: undefined,
    name: '',
    code: '',
    parentId: 0,
    owner: '',
    phone: '',
    email: '',
    sortOrder: 0,
    status: 1,
    remarks: ''
  });
  dialogVisible.value = true;
}

function handleEdit(row: CMDB.BusinessUnit) {
  dialogTitle.value = '编辑业务';
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    parentId: row.parentId,
    owner: row.owner || '',
    phone: row.phone || '',
    email: row.email || '',
    sortOrder: row.sortOrder,
    status: row.status,
    remarks: row.remarks || ''
  });
  dialogVisible.value = true;
}

async function handleSave() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    if (form.id) {
      await fetchUpdateBusinessUnit(form.id, form);
      ElNotification.success('业务更新成功');
    } else {
      await fetchCreateBusinessUnit(form);
      ElNotification.success('业务创建成功');
    }
    dialogVisible.value = false;
    getTableData();
  } catch (error) {
    console.error('保存业务失败:', error);
    ElNotification.error('保存业务失败');
  }
}

function handleDelete(row: CMDB.BusinessUnit) {
  if (row.children && row.children.length > 0) {
    ElNotification.warning('该业务下有子业务，无法删除');
    return;
  }

  ElMessageBox.confirm(`确定要删除业务 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      await fetchDeleteBusinessUnit(row.id);
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
        <ElButton type="primary" @click="handleAdd">新增业务</ElButton>
      </div>

      <ElTable
        v-loading="loading"
        :data="tableData"
        row-key="id"
        border
        stripe
        :tree-props="{ children: 'children' }"
      >
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn prop="name" label="业务名称" min-width="160" show-overflow-tooltip />
        <ElTableColumn prop="code" label="业务代码" width="140" />
        <ElTableColumn prop="owner" label="负责人" width="120" />
        <ElTableColumn prop="phone" label="联系电话" width="140" />
        <ElTableColumn prop="sortOrder" label="排序" width="80" />
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="remarks" label="备注" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="620px">
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
        <ElFormItem label="业务名称" prop="name">
          <ElInput v-model="form.name" placeholder="请输入业务名称" />
        </ElFormItem>
        <ElFormItem label="业务代码" prop="code">
          <ElInput v-model="form.code" placeholder="请输入业务代码" />
        </ElFormItem>
        <ElFormItem label="上级业务">
          <ElTreeSelect
            v-model="form.parentId"
            :data="businessTreeOptions"
            :props="treeSelectProps"
            clearable
            check-strictly
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem label="负责人">
          <ElInput v-model="form.owner" placeholder="请输入负责人" />
        </ElFormItem>
        <ElFormItem label="联系电话">
          <ElInput v-model="form.phone" placeholder="请输入联系电话" />
        </ElFormItem>
        <ElFormItem label="邮箱">
          <ElInput v-model="form.email" placeholder="请输入邮箱" />
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
        <ElFormItem label="备注">
          <ElInput v-model="form.remarks" type="textarea" :rows="3" placeholder="请输入备注" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
