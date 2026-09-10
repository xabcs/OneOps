<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { acknowledgeAlert, fetchAlertStats, fetchAlerts } from '@/service/api/monitoring';
  // 时间格式化与级别标签映射统一收敛到公共工具（formatTime 为兼容旧命名的别名）
  import { formatDateTime as formatTime } from '@/utils/datetime';
  import { createTagMap } from '@/utils/common';

  defineOptions({
    name: 'MonitoringAlerts'
  });

  const loading = ref(false);

  // 过滤条件
  const filterForm = reactive({
    level: '',
    acknowledged: '',
    page: 1,
    pageSize: 20
  });

  // 数据
  const alertList = ref<Alert[]>([]);
  const totalCount = ref(0);
  const stats = ref<Monitoring.AlertStats | null>(null);

  /**
   * 获取告警统计
   */
  async function getStats() {
    const { data } = await fetchAlertStats();
    if (data) {
      stats.value = data;
    }
  }

  /**
   * 获取告警列表
   */
  async function getAlerts() {
    loading.value = true;
    const params: Record<string, unknown> = {
      page: filterForm.page,
      pageSize: filterForm.pageSize
    };
    if (filterForm.level) params.level = filterForm.level;
    if (filterForm.acknowledged !== '') params.acknowledged = filterForm.acknowledged === 'true';

    const { data } = await fetchAlerts(params as Record<string, unknown>);
    if (data) {
      alertList.value = data.list || [];
      totalCount.value = data.total || 0;
    }
    loading.value = false;
  }

  /**
   * 搜索
   */
  function handleSearch() {
    filterForm.page = 1;
    getAlerts();
  }

  /**
   * 重置过滤
   */
  function handleReset() {
    filterForm.level = '';
    filterForm.acknowledged = '';
    filterForm.page = 1;
    getAlerts();
  }

  /**
   * 分页变化
   */
  function handlePageChange(page: number) {
    filterForm.page = page;
    getAlerts();
  }

  function handlePageSizeChange(size: number) {
    filterForm.pageSize = size;
    filterForm.page = 1;
    getAlerts();
  }

  /**
   * 确认告警
   */
  async function handleAcknowledge(alert: Alert) {
    await ElMessageBox.prompt(`确认处理告警？可填写备注（可选）`, '确认告警', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      inputPlaceholder: '备注信息（可选）',
      inputType: 'textarea'
    })
      .then(async ({ value }) => {
        const { error } = await acknowledgeAlert(alert.id, value || undefined);
        if (!error) {
          ElMessage.success('告警已确认');
          getAlerts();
          getStats();
        }
      })
      .catch(() => {});
  }

  /** 告警级别 → ElTag 标签映射 */
  const getLevelTag = createTagMap({
    critical: { text: '严重', type: 'danger' },
    high: { text: '高', type: 'warning' },
    medium: { text: '中', type: 'primary' },
    low: { text: '低', type: 'info' },
    info: { text: '信息', type: 'info' }
  });

  onMounted(() => {
    getStats();
    getAlerts();
  });
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 统计卡片 -->
    <ElRow v-if="stats" :gutter="16">
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-gray-800 font-bold">{{ stats.total }}</div>
          <div class="mt-1 text-sm text-gray-500">告警总数</div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-red-600 font-bold">{{ stats.byLevel.critical }}</div>
          <div class="mt-1 text-sm text-gray-500">严重</div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-orange-500 font-bold">{{ stats.byLevel.high }}</div>
          <div class="mt-1 text-sm text-gray-500">高</div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-blue-500 font-bold">{{ stats.byLevel.medium }}</div>
          <div class="mt-1 text-sm text-gray-500">中</div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-yellow-500 font-bold">{{ stats.active }}</div>
          <div class="mt-1 text-sm text-gray-500">未处理</div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-green-600 font-bold">{{ stats.acknowledged }}</div>
          <div class="mt-1 text-sm text-gray-500">已确认</div>
        </ElCard>
      </ElCol>
    </ElRow>

    <!-- 过滤栏 -->
    <ElCard shadow="never">
      <ElForm :model="filterForm" inline>
        <ElFormItem label="告警级别">
          <ElSelect v-model="filterForm.level" placeholder="全部" clearable style="width: 130px">
            <ElOption label="严重" value="critical" />
            <ElOption label="高" value="high" />
            <ElOption label="中" value="medium" />
            <ElOption label="低" value="low" />
            <ElOption label="信息" value="info" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="处理状态">
          <ElSelect v-model="filterForm.acknowledged" placeholder="全部" clearable style="width: 130px">
            <ElOption label="未处理" value="false" />
            <ElOption label="已确认" value="true" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">搜索</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 告警列表 -->
    <ElCard shadow="never">
      <ElTable v-loading="loading" :data="alertList" border stripe>
        <ElTableColumn label="级别" width="90" align="center">
          <template #default="{ row }">
            <ElTag :type="getLevelTag(row.level).type" size="small">
              {{ getLevelTag(row.level).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="主机名" prop="hostname" min-width="140" />
        <ElTableColumn label="IP地址" prop="ip" width="140" />
        <ElTableColumn label="告警内容" prop="message" min-width="260" show-overflow-tooltip />
        <ElTableColumn label="指标值" width="100" align="right">
          <template #default="{ row }">
            {{ row.metricValue != null ? row.metricValue.toFixed(1) + '%' : '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="阈值" width="90" align="right">
          <template #default="{ row }">
            {{ row.threshold != null ? row.threshold + '%' : '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="首次触发" width="170">
          <template #default="{ row }">{{ formatTime(row.firstSeen) }}</template>
        </ElTableColumn>
        <ElTableColumn label="最后触发" width="170">
          <template #default="{ row }">{{ formatTime(row.lastSeen) }}</template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="100" align="center">
          <template #default="{ row }">
            <ElTag v-if="row.resolvedAt" type="success" size="small">已恢复</ElTag>
            <ElTag v-else-if="row.acknowledged" type="info" size="small">已确认</ElTag>
            <ElTag v-else type="danger" size="small">未处理</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="确认人" prop="acknowledgedBy" width="110">
          <template #default="{ row }">{{ row.acknowledgedBy || '-' }}</template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="100" align="center" fixed="right" class-name="msre-table-actions">
          <template #default="{ row }">
            <PermissionButton
              link
              type="primary"
              size="small"
              v-if="!row.acknowledged && !row.resolvedAt"
              code="monitor.alert.ack"
              @click="handleAcknowledge(row)"
            >
              确认
            </PermissionButton>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="mt-4 flex justify-end">
        <ElPagination
          v-model:current-page="filterForm.page"
          v-model:page-size="filterForm.pageSize"
          :total="totalCount"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>
