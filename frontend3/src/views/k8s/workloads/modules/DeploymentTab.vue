<script setup lang="ts">
  import { ElButton, ElTableColumn, ElTag } from 'element-plus';
  import WorkloadTable from './WorkloadTable.vue';

  defineProps<{
    data: K8s.Deployment[];
    loading: boolean;
    pagination: { page: number; pageSize: number; itemCount: number };
    selectedCount: number;
  }>();

  const emit = defineEmits<{
    (e: 'go-to-detail', row: K8s.Deployment): void;
    (e: 'page-change', page: number): void;
    (e: 'size-change', pageSize: number): void;
    (e: 'selection-change', selection: K8s.Deployment[]): void;
    (e: 'select-all', selection: K8s.Deployment[]): void;
    (e: 'clear-selection'): void;
    (e: 'scale', row: K8s.Deployment): void;
    (e: 'more-command', command: string, row: K8s.Deployment): void;
    (e: 'batch-restart'): void;
    (e: 'batch-delete'): void;
  }>();
</script>

<template>
  <WorkloadTable
    :data="data"
    :loading="loading"
    :pagination="pagination"
    :selected-count="selectedCount"
    :batch-buttons="[
      { label: '批量重启', disabled: selectedCount === 0, handler: () => emit('batch-restart') },
      { label: '批量删除', type: 'danger', disabled: selectedCount === 0, handler: () => emit('batch-delete') }
    ]"
    @page-change="emit('page-change', $event)"
    @size-change="emit('size-change', $event)"
    @selection-change="emit('selection-change', $event)"
    @select-all="emit('select-all', $event)"
    @clear-selection="emit('clear-selection')"
  >
    <ElTableColumn prop="name" label="名称" min-width="180" align="left">
      <template #default="{ row }">
        <ElButton link type="primary" @click="emit('go-to-detail', row)">{{ row.name }}</ElButton>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
    <ElTableColumn prop="replicas" label="副本数" min-width="100" align="left">
      <template #default="{ row }">{{ row.ready }} / {{ row.replicas }}</template>
    </ElTableColumn>
    <ElTableColumn prop="upToDate" label="最新" min-width="80" align="left" />
    <ElTableColumn prop="available" label="可用" min-width="80" align="left" />
    <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
    <ElTableColumn label="状态" min-width="100" align="left">
      <template #default="{ row }">
        <ElTag v-if="row.ready === row.replicas" type="success">运行中</ElTag>
        <ElTag v-else type="warning">更新中</ElTag>
      </template>
    </ElTableColumn>
    <ElTableColumn label="操作" min-width="200" fixed="right" align="left">
      <template #default="{ row }">
        <span class="operation-buttons">
          <ElButton link type="primary" size="default" @click="emit('go-to-detail', row)">详情</ElButton>
          <PermissionButton code="k8s.resource.update" link type="primary" size="default" @click="emit('scale', row)">
            伸缩
          </PermissionButton>
          <ElDropdown trigger="click" @command="cmd => emit('more-command', cmd, row)">
            <span class="dropdown-link">
              更多
              <icon-mdi-chevron-down class="dropdown-icon" />
            </span>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
                <ElDropdownItem v-permission="'k8s.resource.update'" command="restart">重启</ElDropdownItem>
                <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </span>
      </template>
    </ElTableColumn>
  </WorkloadTable>
</template>
