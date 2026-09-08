<script setup lang="ts">
  import { ElButton, ElTableColumn, ElTag } from 'element-plus';
  import WorkloadTable from './WorkloadTable.vue';

  defineProps<{
    data: K8s.WorkloadRow[];
    loading: boolean;
    pagination: { page: number; pageSize: number; itemCount: number };
    selectedCount: number;
    resourceType: 'statefulset' | 'daemonset' | 'job' | 'cronjob';
    columns: 'statefulset' | 'daemonset' | 'job' | 'cronjob';
  }>();

  const emit = defineEmits<{
    (e: 'go-to-detail', row: K8s.WorkloadRow): void;
    (e: 'page-change', page: number): void;
    (e: 'size-change', pageSize: number): void;
    (e: 'selection-change', selection: K8s.WorkloadRow[]): void;
    (e: 'select-all', selection: K8s.WorkloadRow[]): void;
    (e: 'clear-selection'): void;
    (e: 'workload-command', command: string, row: K8s.WorkloadRow): void;
    (e: 'batch-delete-by-type', resourceType: string): void;
  }>();
</script>

<template>
  <WorkloadTable
    :data="data"
    :loading="loading"
    :pagination="pagination"
    :selected-count="selectedCount"
    :batch-buttons="[
      {
        label: '批量删除',
        type: 'danger',
        disabled: selectedCount === 0,
        handler: () => emit('batch-delete-by-type', resourceType)
      }
    ]"
    @page-change="emit('page-change', $event)"
    @size-change="emit('size-change', $event)"
    @selection-change="emit('selection-change', $event)"
    @select-all="emit('select-all', $event)"
    @clear-selection="emit('clear-selection')"
  >
    <!-- 通用名称列 -->
    <ElTableColumn prop="name" label="名称" min-width="180" align="left">
      <template #default="{ row }">
        <ElButton link type="primary" @click="emit('go-to-detail', row)">{{ row.name }}</ElButton>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />

    <!-- StatefulSet 特有列 -->
    <template v-if="columns === 'statefulset'">
      <ElTableColumn prop="replicas" label="副本数" min-width="100" align="left">
        <template #default="{ row }">{{ row.ready }} / {{ row.replicas }}</template>
      </ElTableColumn>
      <ElTableColumn prop="current" label="当前" min-width="80" align="left" />
      <ElTableColumn prop="updated" label="已更新" min-width="80" align="left" />
    </template>

    <!-- DaemonSet 特有列 -->
    <template v-if="columns === 'daemonset'">
      <ElTableColumn label="节点数" min-width="120" align="left">
        <template #default="{ row }">{{ row.current }} / {{ row.desired }}</template>
      </ElTableColumn>
      <ElTableColumn prop="ready" label="就绪" min-width="80" align="left" />
      <ElTableColumn prop="available" label="可用" min-width="80" align="left" />
    </template>

    <!-- Job 特有列 -->
    <template v-if="columns === 'job'">
      <ElTableColumn label="完成数" min-width="100" align="left">
        <template #default="{ row }">{{ row.succeeded || 0 }} / {{ row.completions || '-' }}</template>
      </ElTableColumn>
      <ElTableColumn prop="duration" label="时长" min-width="100" align="left" />
      <ElTableColumn label="状态" min-width="100" align="left">
        <template #default="{ row }">
          <ElTag
            :type="
              row.status === '完成'
                ? 'success'
                : row.status === '失败'
                  ? 'danger'
                  : row.status === '运行中'
                    ? 'primary'
                    : 'info'
            "
          >
            {{ row.status }}
          </ElTag>
        </template>
      </ElTableColumn>
    </template>

    <!-- CronJob 特有列 -->
    <template v-if="columns === 'cronjob'">
      <ElTableColumn prop="schedule" label="调度规则" min-width="160" align="left" />
      <ElTableColumn label="挂起" min-width="80" align="left">
        <template #default="{ row }">
          <ElTag :type="row.suspend ? 'warning' : 'success'">{{ row.suspend ? '是' : '否' }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="上次执行" min-width="160" align="left">
        <template #default="{ row }">{{ row.lastSchedule || '-' }}</template>
      </ElTableColumn>
    </template>

    <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />

    <!-- 操作列 -->
    <ElTableColumn
      label="操作"
      :min-width="columns === 'cronjob' ? 200 : 180"
      fixed="right"
      align="center"
      class-name="msre-table-actions"
    >
      <template #default="{ row }">
        <ElButton link type="primary" size="small" @click="emit('go-to-detail', row)">详情</ElButton>
        <ElDropdown trigger="click" @command="cmd => emit('workload-command', cmd, row)">
          <ElButton link type="primary" size="small" class="table-dropdown-trigger">
            更多
            <icon-mdi-chevron-down class="dropdown-icon" />
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
              <ElDropdownItem
                v-if="columns === 'statefulset' || columns === 'daemonset'"
                v-permission="'k8s.resource.update'"
                command="restart"
              >
                重启
              </ElDropdownItem>
              <ElDropdownItem v-if="columns === 'cronjob'" v-permission="'k8s.resource.update'" command="suspend">
                {{ row.suspend ? '恢复' : '暂停' }}
              </ElDropdownItem>
              <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </template>
    </ElTableColumn>
  </WorkloadTable>
</template>
