<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue';
import {
  createAuthGroup,
  deleteAuthGroup,
  fetchAuthGroups,
  updateAuthGroup
} from '@/service/api/application-permission';

defineOptions({ name: 'AuthCenterGroups' });

const loading = ref(false);
const tableData = ref<Api.ApplicationPermission.AuthGroup[]>([]);
const total = ref(0);
const searchName = ref('');

const pagination = ref({
  page: 1,
  pageSize: 10
});

const drawerVisible = ref(false);
const isEdit = ref(false);

const formData = ref<Api.ApplicationPermission.AuthGroup>({
  name: '',
  code: '',
  description: '',
  status: 1
});

async function getData() {
  loading.value = true;
  try {
    const { data, error } = await fetchAuthGroups({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      name: searchName.value
    });
    if (!error && data) {
      tableData.value = data.list || [];
      total.value = data.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.value.page = 1;
  getData();
}

function handleReset() {
  searchName.value = '';
  pagination.value.page = 1;
  getData();
}

function handleAdd() {
  isEdit.value = false;
  formData.value = {
    name: '',
    code: '',
    description: '',
    status: 1
  };
  drawerVisible.value = true;
}

function handleEdit(row: any) {
  isEdit.value = true;
  formData.value = { ...row };
  drawerVisible.value = true;
}

async function handleSubmit() {
  const api = isEdit.value ? updateAuthGroup(formData.value.id!, formData.value) : createAuthGroup(formData.value);

  const { error } = (await api) as any;
  if (!error) {
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功');
    drawerVisible.value = false;
    getData();
  }
}

async function handleDelete(id: number) {
  const { error } = await deleteAuthGroup(id);
  if (!error) {
    ElMessage.success('删除成功');
    getData();
  }
}

function handlePageChange(page: number) {
  pagination.value.page = page;
  getData();
}

function handleSizeChange(size: number) {
  pagination.value.pageSize = size;
  pagination.value.page = 1;
  getData();
}

onMounted(() => {
  getData();
});
</script>

<template>
    <div class="min-h-500px flex-col gap-4">
        <!-- 搜索栏 -->
        <ElCard shadow="never">
            <ElForm :inline="true">
                <ElFormItem label="用户组名称">
                    <ElInput v-model="searchName" placeholder="请输入用户组名称" clearable class="w-200px" @keyup.enter="handleSearch" />
                </ElFormItem>
                <ElFormItem>
                    <ElButton type="primary" :icon="Search" @click="handleSearch">搜索</ElButton>
                    <ElButton :icon="Refresh" @click="handleReset">重置</ElButton>
                </ElFormItem>
            </ElForm>
        </ElCard>

        <!-- 表格 -->
        <ElCard shadow="never" class="flex-1">
            <template #header>
                <div class="flex items-center justify-between">
                    <span class="text-lg font-medium">授权中心用户组列表</span>
                    <ElButton type="primary" :icon="Plus" @click="handleAdd">添加用户组</ElButton>
                </div>
            </template>

            <ElTable v-loading="loading" :data="tableData" :border="false">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="name" label="用户组名称" align="center" min-width="120" />
                <ElTableColumn prop="code" label="用户组代码" align="center" min-width="120" />
                <ElTableColumn prop="description" label="描述" align="center" min-width="200" />
                <ElTableColumn prop="status" label="状态" align="center" width="80">
                    <template #default="{ row }">
                        <ElTag :type="row.status === 1 ? 'success' : 'info'">
                            {{ row.status === 1 ? '启用' : '禁用' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" align="center" width="150" fixed="right">
                    <template #default="{ row }">
                        <ElSpace>
                            <ElButton size="small" type="warning" :icon="Edit" @click="handleEdit(row)">编辑</ElButton>
                            <ElButton size="small" type="danger" :icon="Delete" @click="handleDelete(row.id)">删除</ElButton>
                        </ElSpace>
                    </template>
                </ElTableColumn>
            </ElTable>

            <div class="mt-4 flex justify-end">
                <ElPagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" @size-change="handleSizeChange" @current-change="handlePageChange" />
            </div>
        </ElCard>

        <!-- 添加/编辑抽屉 -->
        <ElDrawer v-model="drawerVisible" :title="isEdit ? '编辑用户组' : '添加用户组'" :width="500">
            <ElForm :model="formData" label-width="100px">
                <ElFormItem label="用户组名称" required>
                    <ElInput v-model="formData.name" placeholder="请输入用户组名称" />
                </ElFormItem>
                <ElFormItem label="用户组代码" required>
                    <ElInput v-model="formData.code" placeholder="请输入用户组代码" :disabled="isEdit" />
                </ElFormItem>
                <ElFormItem label="描述">
                    <ElInput v-model="formData.description" type="textarea" :rows="3" placeholder="请输入描述" />
                </ElFormItem>
                <ElFormItem label="状态">
                    <ElRadioGroup v-model="formData.status">
                        <ElRadio :value="1">启用</ElRadio>
                        <ElRadio :value="0">禁用</ElRadio>
                    </ElRadioGroup>
                </ElFormItem>
            </ElForm>
            <template #footer>
                <ElButton @click="drawerVisible = false">取消</ElButton>
                <ElButton type="primary" @click="handleSubmit">确定</ElButton>
            </template>
        </ElDrawer>
    </div>
</template>
