<script setup lang="ts">
    import { ElButton, ElTableColumn, ElTag } from 'element-plus';
    import WorkloadTable from './WorkloadTable.vue';

    defineProps<{
      data: K8s.Pod[];
      loading: boolean;
      pagination: { page: number; pageSize: number; itemCount: number };
      selectedCount: number;
    }>();

    const emit = defineEmits<{
      (e: 'go-to-detail', row: K8s.Pod): void;
      (e: 'page-change', page: number): void;
      (e: 'size-change', pageSize: number): void;
      (e: 'selection-change', selection: K8s.Pod[]): void;
      (e: 'select-all', selection: K8s.Pod[]): void;
      (e: 'clear-selection'): void;
      (e: 'pod-logs', row: K8s.Pod): void;
      (e: 'pod-more-command', command: string, row: K8s.Pod): void;
      (e: 'batch-delete-pods'): void;
    }>();
</script>

<template>
    <WorkloadTable :data="data" :loading="loading" :pagination="pagination" :selected-count="selectedCount" :batch-buttons="[
      { label: '批量删除', type: 'danger', disabled: selectedCount === 0, handler: () => emit('batch-delete-pods') }
    ]" @page-change="emit('page-change', $event)" @size-change="emit('size-change', $event)" @selection-change="emit('selection-change', $event)" @select-all="emit('select-all', $event)" @clear-selection="emit('clear-selection')">
        <ElTableColumn prop="name" label="名称" min-width="200" align="left">
            <template #default="{ row }">
                <ElButton link type="primary" @click="emit('go-to-detail', row)">{{ row.name }}</ElButton>
            </template>
        </ElTableColumn>
        <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
        <ElTableColumn label="状态" min-width="100" align="left">
            <template #default="{ row }">
                <ElTag :type="row.status === 'Running' ? 'success' : 'warning'">{{ row.status }}</ElTag>
            </template>
        </ElTableColumn>
        <ElTableColumn prop="ip" label="IP地址" min-width="140" align="left" />
        <ElTableColumn prop="node" label="节点" min-width="150" align="left" />
        <ElTableColumn prop="restarts" label="重启次数" min-width="100" align="left" />
        <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
        <ElTableColumn label="操作" min-width="150" fixed="right" align="left">
            <template #default="{ row }">
                <span class="operation-buttons">
                    <ElButton link type="primary" size="default" @click="emit('go-to-detail', row)">详情</ElButton>
                    <ElButton link type="primary" size="default" @click="emit('pod-logs', row)">日志</ElButton>
                    <ElDropdown trigger="click" @command="cmd => emit('pod-more-command', cmd, row)">
                        <span class="dropdown-link">
                            更多
                            <icon-mdi-chevron-down class="dropdown-icon" />
                        </span>
                        <template #dropdown>
                            <ElDropdownMenu>
                                <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                                <ElDropdownItem command="terminal">终端</ElDropdownItem>
                                <ElDropdownItem command="delete">删除</ElDropdownItem>
                            </ElDropdownMenu>
                        </template>
                    </ElDropdown>
                </span>
            </template>
        </ElTableColumn>
    </WorkloadTable>
</template>
