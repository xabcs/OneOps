<script setup lang="ts">
    import { computed, reactive, ref, watch } from 'vue';
    import { ElMessage, ElMessageBox } from 'element-plus';
    import { Plus, Refresh, Search } from '@element-plus/icons-vue';
    import {
      fetchCreatePermissionRoute,
      fetchDeletePermissionRoute,
      fetchGetPermissionRoutes,
      fetchPermissionOptions,
      fetchUpdatePermissionRoute
    } from '@/service/api';
    import { useForm, useFormRules } from '@/hooks/common/form';

    defineOptions({ name: 'PermissionRouteModal' });

    const props = defineProps<{
      /** 打开时预置的权限码过滤 */
      initialPermissionCode?: string;
    }>();

    const visible = defineModel<boolean>('visible', { default: false });

    // @ts-expect-error vue-tsc noUnusedLocals: template ref
    const { formRef, validate, restoreValidation } = useForm();
    const { defaultRequiredRule } = useFormRules();

    const loading = ref(false);
    const routes = ref<Api.SystemManage.PermissionRoute[]>([]);
    const permissionOptions = ref<{ id: number; name: string; code: string }[]>([]);

    /** 本地过滤（后端返回全量映射） */
    const filterCode = ref('');
    const filterPath = ref('');

    const filteredRoutes = computed(() =>
      routes.value.filter(
        item =>
          (!filterCode.value || item.permissionCode.includes(filterCode.value.trim())) &&
          (!filterPath.value || item.path.includes(filterPath.value.trim()))
      )
    );

    const methodColors: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
      GET: 'success',
      POST: 'primary',
      PUT: 'warning',
      DELETE: 'danger',
      PATCH: 'info'
    };

    const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];

    const addModel = reactive({
      permissionCode: '',
      method: 'GET',
      path: ''
    });

    const addRules: Record<string, App.Global.FormRule[]> = {
      permissionCode: [defaultRequiredRule],
      method: [defaultRequiredRule],
      path: [
        { required: true, message: '请输入路由路径' },
        { pattern: /^\/api\//, message: '路径必须以 /api/ 开头' }
      ]
    };

    /** 改码弹窗 */
    const editVisible = ref(false);
    const editingRoute = ref<Api.SystemManage.PermissionRoute | null>(null);
    const editModel = reactive({ permissionCode: '' });

    async function loadData() {
      loading.value = true;
      const { error, data } = await fetchGetPermissionRoutes();
      if (!error && data) {
        routes.value = data;
      }
      loading.value = false;
    }

    async function loadPermissionOptions() {
      const { error, data } = await fetchPermissionOptions();
      if (!error && data) {
        permissionOptions.value = data;
      }
    }

    async function handleAdd() {
      await validate();
      const { error } = await fetchCreatePermissionRoute({
        permissionCode: addModel.permissionCode,
        method: addModel.method,
        path: addModel.path.trim()
      });
      if (!error) {
        ElMessage.success('创建成功，即时生效');
        addModel.permissionCode = '';
        addModel.path = '';
        restoreValidation();
        await loadData();
      }
    }

    function openEdit(row: Api.SystemManage.PermissionRoute) {
      editingRoute.value = row;
      editModel.permissionCode = row.permissionCode;
      editVisible.value = true;
    }

    async function handleEditSubmit() {
      if (!editingRoute.value) return;
      if (!editModel.permissionCode) {
        ElMessage.warning('请选择权限编码');
        return;
      }
      const { error } = await fetchUpdatePermissionRoute(editingRoute.value.id, {
        permissionCode: editModel.permissionCode
      });
      if (!error) {
        ElMessage.success('更新成功');
        editVisible.value = false;
        await loadData();
      }
    }

    async function handleDelete(row: Api.SystemManage.PermissionRoute) {
      await ElMessageBox.confirm(
        `删除映射 ${row.method} ${row.path} → ${row.permissionCode}。若该端点无其他权限码，将变为拒绝访问（fail-closed）。确认删除？`,
        '删除路由映射',
        {
          type: 'warning',
          confirmButtonText: '删除',
          cancelButtonText: '取消'
        }
      );
      const { error } = await fetchDeletePermissionRoute(row.id);
      if (!error) {
        ElMessage.success('删除成功');
        await loadData();
      }
    }

    watch(visible, val => {
      if (val) {
        filterCode.value = props.initialPermissionCode || '';
        filterPath.value = '';
        loadData();
        if (permissionOptions.value.length === 0) {
          loadPermissionOptions();
        }
      }
    });
</script>

<template>
    <ElDialog v-model="visible" title="路由映射管理" width="960px" top="6vh" :close-on-click-modal="false">
        <ElAlert type="info" :closable="false" show-icon class="mb-16px" title="每个 API 端点必须配置至少一个权限码，未配置的端点将被拒绝访问（fail-closed）。一个权限码可保护多个端点；同一端点也可配置多个权限码（用户持有任一即可访问），用于实现「单接口权限」与「接口集合权限」两种语义并存。" />

        <!-- 新增表单 -->
        <ElForm ref="formRef" inline :model="addModel" :rules="addRules" class="mb-12px">
            <ElFormItem prop="permissionCode">
                <ElSelect v-model="addModel.permissionCode" placeholder="选择权限编码" filterable style="width: 260px">
                    <ElOption v-for="opt in permissionOptions" :key="opt.id" :label="`${opt.code}（${opt.name}）`" :value="opt.code" />
                </ElSelect>
            </ElFormItem>
            <ElFormItem prop="method">
                <ElSelect v-model="addModel.method" style="width: 110px">
                    <ElOption v-for="m in methodOptions" :key="m" :label="m" :value="m" />
                </ElSelect>
            </ElFormItem>
            <ElFormItem prop="path">
                <ElInput v-model="addModel.path" placeholder="如：/api/system/users" style="width: 320px" @keyup.enter="handleAdd" />
            </ElFormItem>
            <ElFormItem>
                <PermissionButton code="system.permission.create" type="primary" @click="handleAdd">
                    <ElIcon>
                        <Plus />
                    </ElIcon>
                    添加映射
                </PermissionButton>
            </ElFormItem>
        </ElForm>

        <!-- 过滤 -->
        <ElSpace wrap class="mb-12px">
            <ElInput v-model="filterCode" placeholder="按权限编码过滤" clearable :prefix-icon="Search" style="width: 240px" />
            <ElInput v-model="filterPath" placeholder="按路径过滤" clearable :prefix-icon="Search" style="width: 260px" />
            <ElButton :loading="loading" @click="loadData">
                <ElIcon>
                    <Refresh />
                </ElIcon>
                刷新
            </ElButton>
            <span class="text-13px opacity-60">共 {{ filteredRoutes.length }} 条</span>
        </ElSpace>

        <ElTable v-loading="loading" :data="filteredRoutes" border stripe max-height="480" row-key="id">
            <ElTableColumn prop="id" label="ID" width="70" />
            <ElTableColumn prop="permissionCode" label="权限编码" min-width="200">
                <template #default="{ row }">
                    <span style="font-family: monospace; font-size: 12px; color: var(--el-color-primary)">
                        {{ row.permissionCode }}
                    </span>
                </template>
            </ElTableColumn>
            <ElTableColumn prop="method" label="方法" width="90">
                <template #default="{ row }">
                    <ElTag size="small" :type="methodColors[row.method] || 'info'">{{ row.method }}</ElTag>
                </template>
            </ElTableColumn>
            <ElTableColumn prop="path" label="API 路径" min-width="280">
                <template #default="{ row }">
                    <span style="font-family: monospace; font-size: 12px">{{ row.path }}</span>
                </template>
            </ElTableColumn>
            <ElTableColumn prop="isSeed" label="来源" width="80">
                <template #default="{ row }">
                    <ElTag size="small" :type="row.isSeed ? 'info' : 'success'">
                        {{ row.isSeed ? '初始化' : '页面' }}
                    </ElTag>
                </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="150" fixed="right">
                <template #default="{ row }">
                    <ElSpace size="small">
                        <PermissionButton code="system.permission.update" text type="primary" size="small" @click="openEdit(row)">改权限码</PermissionButton>
                        <PermissionButton code="system.permission.delete" text type="danger" size="small" @click="handleDelete(row)">删除</PermissionButton>
                    </ElSpace>
                </template>
            </ElTableColumn>
        </ElTable>

        <!-- 改码弹窗 -->
        <ElDialog v-model="editVisible" title="修改归属权限码" width="440px" append-to-body>
            <ElForm label-position="top">
                <ElFormItem label="端点（不可修改）">
                    <ElInput :model-value="`${editingRoute?.method} ${editingRoute?.path}`" disabled />
                </ElFormItem>
                <ElFormItem label="目标权限编码" required>
                    <ElSelect v-model="editModel.permissionCode" placeholder="选择权限编码" filterable style="width: 100%">
                        <ElOption v-for="opt in permissionOptions" :key="opt.id" :label="`${opt.code}（${opt.name}）`" :value="opt.code" />
                    </ElSelect>
                </ElFormItem>
            </ElForm>
            <template #footer>
                <ElSpace :size="16">
                    <ElButton @click="editVisible = false">取消</ElButton>
                    <ElButton type="primary" @click="handleEditSubmit">确认</ElButton>
                </ElSpace>
            </template>
        </ElDialog>
    </ElDialog>
</template>

<style scoped></style>
