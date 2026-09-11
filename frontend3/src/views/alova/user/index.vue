<script setup lang="tsx">
  import { ref } from 'vue';
  import { usePagination } from '@sa/alova/client';
  import { enableStatusOptions, enableStatusRecord, userGenderOptions, userGenderRecord } from '@/constants/business';
  import { batchDeleteUser, deleteUser, fetchGetUserList } from '@/service-alova/api';
  import { translateOptions } from '@/utils/common';
  import { $t } from '@/locales';
  import useCheckedColumns from './hooks/use-checked-columns';
  import useTableOperate from './hooks/use-table-operate';
  import UserOperateDrawer from './modules/user-operate-drawer.vue';

  const searchParams = ref({
    status: undefined,
    userName: undefined,
    userGender: undefined,
    nickName: undefined,
    userPhone: undefined,
    userEmail: undefined
  });
  const { loading, data, refresh, reload, page, pageSize, pageCount, send, remove, total } = usePagination(
    (pageNum, size) =>
      fetchGetUserList({
        ...searchParams.value,
        current: pageNum,
        size
      }),
    {
      data: ({ records }) => records,

      // trigger reload when states in `searchParams` changed
      watchingStates: [searchParams.value],

      // debounce of `searchParams`
      debounce: [1000]
    }
  );
  const getDataByPage = (newPage = 1) => {
    page.value = newPage;
    send(page.value, pageSize.value);
  };

  const handleSizeChange = (newSize: number) => {
    pageSize.value = newSize;
    send(page.value, newSize);
  };

  // 布局重置按钮：恢复初始筛选（Object.assign 保持引用，alova watching 自动刷新）
  function handleReset() {
    Object.assign(searchParams.value, {
      status: undefined,
      userName: undefined,
      userGender: undefined,
      nickName: undefined,
      userPhone: undefined,
      userEmail: undefined
    });
  }

  const {
    drawerVisible,
    operateType,
    editingData,
    handleAdd,
    handleEdit,
    handleDelete,
    handleBatchDelete,
    checkedRowKeys
    // batchDeleting
    // closeDrawer
  } = useTableOperate(data, {
    async delete(row) {
      await deleteUser(row.id);
      remove(row);
    },
    async batchDelete(rows) {
      await batchDeleteUser(rows.map(({ id }) => id));
      remove(...rows);
    }
  });

  function edit(id: number) {
    handleEdit(id);
  }

  const { columnChecks, columns } = useCheckedColumns<typeof fetchGetUserList>(() => [
    { type: 'selection', width: 48 },
    { prop: 'userName', label: $t('page.manage.user.userName'), minWidth: 100 },
    {
      prop: 'userGender',
      label: $t('page.manage.user.userGender'),
      width: 100,
      formatter: row => {
        if (row.userGender === undefined) {
          return '';
        }

        const tagMap: Record<Api.SystemManage.UserGender, UI.ThemeColor> = {
          1: 'primary',
          2: 'danger'
        };

        const label = $t(userGenderRecord[row.userGender]);

        return <ElTag type={tagMap[row.userGender]}>{label}</ElTag>;
      }
    },
    { prop: 'nickName', label: $t('page.manage.user.nickName'), minWidth: 100 },
    { prop: 'userPhone', label: $t('page.manage.user.userPhone'), width: 120 },
    { prop: 'userEmail', label: $t('page.manage.user.userEmail'), minWidth: 200 },
    {
      prop: 'status',
      label: $t('page.manage.user.userStatus'),
      width: 100,
      formatter: row => {
        if (row.status === undefined) {
          return '';
        }

        const tagMap: Record<Api.Common.EnableStatus, UI.ThemeColor> = {
          1: 'success',
          2: 'warning'
        };

        const label = $t(enableStatusRecord[row.status]);

        return <ElTag type={tagMap[row.status]}>{label}</ElTag>;
      }
    },
    {
      prop: 'operate',
      label: $t('common.operate'),
      className: 'msre-table-actions',
      width: 130,
      align: 'center',
      fixed: 'right',
      formatter: row => (
        <div>
          <ElButton link type="primary" size="small" onClick={() => edit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton link type="danger" size="small">
                  {$t('common.delete')}
                </ElButton>
              )
            }}
          </ElPopconfirm>
        </div>
      )
    }
  ]);
</script>

<template>
  <ListPageLayout
    :title="$t('page.manage.user.title')"
    :pagination="
      total
        ? {
            total,
            currentPage: page,
            pageSize,
            pageCount,
            pageSizes: [10, 15, 20, 25, 30],
            'current-change': getDataByPage,
            'size-change': handleSizeChange
          }
        : null
    "
    @search="getDataByPage"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput v-model="searchParams.userName" :placeholder="$t('page.manage.user.form.userName')" clearable class="w-180px" />
      <ElSelect
        v-model="searchParams.userGender"
        :placeholder="$t('page.manage.user.form.userGender')"
        clearable
        class="w-120px"
      >
        <ElOption
          v-for="{ label, value } in translateOptions(userGenderOptions)"
          :key="value"
          :label="label"
          :value="value"
        />
      </ElSelect>
      <ElInput v-model="searchParams.nickName" :placeholder="$t('page.manage.user.form.nickName')" clearable class="w-180px" />
      <ElInput v-model="searchParams.userPhone" :placeholder="$t('page.manage.user.form.userPhone')" clearable class="w-160px" />
      <ElInput v-model="searchParams.userEmail" :placeholder="$t('page.manage.user.form.userEmail')" clearable class="w-200px" />
      <ElSelect
        v-model="searchParams.status"
        :placeholder="$t('page.manage.user.form.userStatus')"
        clearable
        class="w-120px"
      >
        <ElOption
          v-for="{ label, value } in translateOptions(enableStatusOptions)"
          :key="value"
          :label="label"
          :value="value"
        />
      </ElSelect>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <TableHeaderOperation
        v-model:columns="columnChecks"
        :disabled-delete="checkedRowKeys.length === 0"
        :loading="loading"
        @add="handleAdd"
        @delete="handleBatchDelete"
        @refresh="refresh"
      />
    </template>

    <!-- 表格 -->
    <ElTable
      v-loading="loading"
      height="100%"
      border
      :data="data"
      row-key="id"
      @selection-change="checkedRowKeys = $event"
    >
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>

    <!-- 用户新增/编辑抽屉 -->
    <UserOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="reload"
    />
  </ListPageLayout>
</template>
