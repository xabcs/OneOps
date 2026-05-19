<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { fetchGetServerRooms, fetchCreateServerRoom, fetchUpdateServerRoom, fetchDeleteServerRoom } from '@/service/api';
import { ElNotification, ElMessageBox, FormInstance, FormRules } from 'element-plus';

defineOptions({ name: 'CmdbRooms' });

const loading = ref(false);
const tableData = ref<CMDB.ServerRoom[]>([]);
const searchKeyword = ref('');

// 对话框
const dialogVisible = ref(false);
const dialogTitle = ref('');
const roomFormRef = ref<FormInstance>();
const roomForm = reactive<{
  id?: number;
  name: string;
  code: string;
  location: string;
  address: string;
  provider: string;
  contact: string;
  phone: string;
  status: number;
  remarks: string;
}>({
  name: '',
  code: '',
  location: '',
  address: '',
  provider: '',
  contact: '',
  phone: '',
  status: 1,
  remarks: ''
});

// 表单验证规则
const roomFormRules: FormRules = {
  name: [
    { required: true, message: '请输入机房名称', trigger: 'blur' },
    { min: 2, max: 100, message: '机房名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入机房代码', trigger: 'blur' },
    { min: 2, max: 50, message: '机房代码长度在 2 到 50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '机房代码只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ]
};

// 服务商选项
const providerOptions = [
  { label: '阿里云', value: 'aliyun' },
  { label: '腾讯云', value: 'tencent' },
  { label: 'AWS', value: 'aws' },
  { label: '华为云', value: 'huawei' },
  { label: '自建', value: 'self' },
  { label: '其他', value: 'other' }
];

// 获取机房列表
async function getTableData() {
  loading.value = true;
  try {
    const { data } = await fetchGetServerRooms();
    if (data) {
      const keyword = searchKeyword.value.trim().toLowerCase();
      tableData.value = keyword
        ? (data || []).filter(item =>
            [item.name, item.code, item.location, item.provider].some(value => value?.toLowerCase().includes(keyword))
          )
        : data || [];
    }
  } catch (err) {
    console.error('获取机房列表失败:', err);
    ElNotification.error('获取机房列表失败');
  } finally {
    loading.value = false;
  }
}

// 打开新增对话框
function handleAdd() {
  dialogTitle.value = '新增机房';
  Object.assign(roomForm, {
    name: '',
    code: '',
    location: '',
    address: '',
    provider: '',
    contact: '',
    phone: '',
    status: 1,
    remarks: ''
  });
  dialogVisible.value = true;
}

// 打开编辑对话框
function handleEdit(row: CMDB.ServerRoom) {
  dialogTitle.value = '编辑机房';
  Object.assign(roomForm, {
    id: row.id,
    name: row.name,
    code: row.code,
    location: row.location,
    address: row.address,
    provider: row.provider,
    contact: row.contact,
    phone: row.phone,
    status: row.status,
    remarks: row.remarks
  });
  dialogVisible.value = true;
}

// 保存
async function handleSave() {
  if (!roomFormRef.value) return;

  try {
    await roomFormRef.value.validate();

    if (roomForm.id) {
      await fetchUpdateServerRoom(roomForm.id, roomForm);
      ElNotification.success('机房更新成功');
    } else {
      await fetchCreateServerRoom(roomForm);
      ElNotification.success('机房创建成功');
    }
    dialogVisible.value = false;
    getTableData();
  } catch (error) {
    console.error('保存失败:', error);
    ElNotification.error('保存失败');
  }
}

// 删除
function handleDelete(row: CMDB.ServerRoom) {
  ElMessageBox.confirm(`确定要删除机房 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
  .then(async () => {
    try {
      await fetchDeleteServerRoom(row.id);
      ElNotification.success('删除成功');
      getTableData();
    } catch (error) {
      console.error('删除失败:', error);
      ElNotification.error('删除失败');
    }
  })
  .catch(() => {});
}

// 获取状态标签
function getStatusTag(status: number) {
  return status === 1
    ? ({ text: '启用', type: 'success' } as const)
    : ({ text: '禁用', type: 'danger' } as const);
}

// 初始化
onMounted(() => {
  getTableData();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <!-- 操作栏 -->
      <div class="mb-16px flex justify-between">
        <ElSpace>
          <ElInput
            v-model="searchKeyword"
            placeholder="搜索机房名称或代码"
            clearable
            style="width: 200px"
            @input="getTableData"
          >
            <template #prefix>
              <icon-mdi-magnify class="align-sub text-icon" />
            </template>
          </ElInput>
        </ElSpace>
        <ElButton type="primary" @click="handleAdd">
          <template #icon>
            <icon-mdi-plus class="align-sub text-icon" />
          </template>
          新增机房
        </ElButton>
      </div>

      <!-- 数据表格 -->
      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="name" label="机房名称" min-width="120" align="center" />
        <ElTableColumn prop="code" label="机房代码" width="120" align="center" />
        <ElTableColumn prop="location" label="位置" min-width="150" align="center" show-overflow-tooltip />
        <ElTableColumn prop="address" label="详细地址" min-width="200" align="center" show-overflow-tooltip />
        <ElTableColumn prop="provider" label="服务商" width="100" align="center" />
        <ElTableColumn prop="contact" label="联系人" width="100" align="center" />
        <ElTableColumn prop="phone" label="联系电话" width="120" align="center" />
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.status).type" size="small">
              {{ getStatusTag(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="remarks" label="备注" min-width="150" align="center" show-overflow-tooltip />
        <ElTableColumn label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
            <ElButton type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <!-- 新增/编辑对话框 -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <ElForm ref="roomFormRef" :model="roomForm" :rules="roomFormRules" label-width="100px">
        <ElFormItem label="机房名称" prop="name">
          <ElInput v-model="roomForm.name" placeholder="请输入机房名称" />
        </ElFormItem>
        <ElFormItem label="机房代码" prop="code">
          <ElInput v-model="roomForm.code" placeholder="请输入机房代码（如：ALIYUN-HUADONG）" />
        </ElFormItem>
        <ElFormItem label="位置">
          <ElInput v-model="roomForm.location" placeholder="请输入位置（如：杭州）" />
        </ElFormItem>
        <ElFormItem label="详细地址">
          <ElInput v-model="roomForm.address" placeholder="请输入详细地址" />
        </ElFormItem>
        <ElFormItem label="服务商">
          <ElSelect v-model="roomForm.provider" placeholder="请选择服务商" style="width: 100%">
            <ElOption v-for="option in providerOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="联系人">
          <ElInput v-model="roomForm.contact" placeholder="请输入联系人" />
        </ElFormItem>
        <ElFormItem label="联系电话">
          <ElInput v-model="roomForm.phone" placeholder="请输入联系电话" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElRadioGroup v-model="roomForm.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput v-model="roomForm.remarks" type="textarea" :rows="3" placeholder="请输入备注信息" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
.card-wrapper {
  @apply flex-col-stretch p-16px;
}
</style>
