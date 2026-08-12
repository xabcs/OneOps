<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue';
  import { ElMessage } from 'element-plus';
  import { Document } from '@element-plus/icons-vue';
  import { fetchGetServers } from '@/service/api';

  defineOptions({
    name: 'MonitoringReports'
  });

  const loading = ref(false);
  const generating = ref(false);

  const servers = ref<CMDB.Server[]>([]);

  // 报告类型选项
  const reportTypes = [
    { label: '主机巡检报告', value: 'host' },
    { label: '安全合规报告', value: 'security' },
    { label: '容量分析报告', value: 'capacity' }
  ];

  // 报告表单
  const reportForm = reactive({
    reportType: 'host',
    serverIds: [] as number[],
    title: ''
  });

  // 报告历史
  const reportHistory = ref<
    Array<{
      id: number;
      type: string;
      title: string;
      status: string;
      createdAt: string;
      completedAt?: string;
    }>
  >([]);

  // 获取主机列表
  async function getServerList() {
    loading.value = true;
    const { data } = await fetchGetServers({ page: 1, pageSize: 1000, agentStatus: 'running' });
    if (data) {
      servers.value = data.list || [];
    }
    loading.value = false;
  }

  // 生成报告
  async function generateReport() {
    if (reportForm.serverIds.length === 0) {
      ElMessage.warning('请至少选择一台主机');
      return;
    }

    if (!reportForm.title) {
      const typeText = reportTypes.find(t => t.value === reportForm.reportType)?.label || '';
      reportForm.title = `${typeText}-${new Date().toLocaleString('zh-CN')}`;
    }

    generating.value = true;

    // 模拟生成报告
    setTimeout(() => {
      const newReport = {
        id: Date.now(),
        type: reportForm.reportType,
        title: reportForm.title,
        status: 'completed',
        createdAt: new Date().toISOString(),
        completedAt: new Date().toISOString()
      };

      reportHistory.value.unshift(newReport);
      generating.value = false;
      ElMessage.success('报告生成成功');
      reportForm.title = '';
    }, 2000);
  }

  // 查看报告
  function viewReport(report: { id: number; status: string }) {
    if (report.status !== 'completed') {
      ElMessage.warning('报告尚未生成完成');
      return;
    }

    // 这里应该跳转到报告详情页或打开PDF预览
    ElMessage.info(`查看报告功能开发中，报告ID: ${report.id}`);
  }

  // 导出报告
  function exportReport(report: { id: number; status: string }) {
    if (report.status !== 'completed') {
      ElMessage.warning('报告尚未生成完成');
      return;
    }

    ElMessage.success(`导出报告功能开发中，报告ID: ${report.id}`);
  }

  // 删除报告
  function deleteReport(id: number) {
    const index = reportHistory.value.findIndex(r => r.id === id);
    if (index > -1) {
      reportHistory.value.splice(index, 1);
      ElMessage.success('删除成功');
    }
  }

  // 获取报告状态标签类型
  function getStatusTagType(status: string): 'success' | 'warning' | 'danger' | 'info' {
    const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
      completed: 'success',
      generating: 'warning',
      failed: 'danger',
      pending: 'info'
    };
    return map[status] || 'info';
  }

  // 获取报告状态文本
  function getStatusText(status: string): string {
    const map: Record<string, string> = {
      completed: '已完成',
      generating: '生成中',
      failed: '失败',
      pending: '等待中'
    };
    return map[status] || status;
  }

  // 格式化时间
  function formatTime(time: string): string {
    if (!time) return '-';
    return new Date(time).toLocaleString('zh-CN');
  }

  onMounted(() => {
    getServerList();

    // 模拟加载历史报告
    reportHistory.value = [
      {
        id: 1,
        type: 'host',
        title: '主机巡检报告-2026-05-26',
        status: 'completed',
        createdAt: '2026-05-26T10:00:00',
        completedAt: '2026-05-26T10:05:00'
      },
      {
        id: 2,
        type: 'security',
        title: '安全合规报告-2026-05-25',
        status: 'completed',
        createdAt: '2026-05-25T15:00:00',
        completedAt: '2026-05-25T15:03:00'
      }
    ];
  });
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 标题栏 -->
    <ElCard shadow="never">
      <div class="flex items-center justify-between">
        <span class="text-lg font-semibold">巡检报告</span>
      </div>
    </ElCard>

    <!-- 创建报告 -->
    <ElCard shadow="never" header="生成报告">
      <ElForm :model="reportForm" label-width="100px">
        <ElRow :gutter="16">
          <ElCol :xs="24" :sm="12">
            <ElFormItem label="报告类型">
              <ElSelect v-model="reportForm.reportType" style="width: 100%">
                <ElOption v-for="type in reportTypes" :key="type.value" :label="type.label" :value="type.value" />
              </ElSelect>
            </ElFormItem>
          </ElCol>

          <ElCol :xs="24" :sm="12">
            <ElFormItem label="报告标题">
              <ElInput v-model="reportForm.title" placeholder="留空自动生成" clearable />
            </ElFormItem>
          </ElCol>

          <ElCol :xs="24">
            <ElFormItem label="选择主机">
              <ElSelect
                v-model="reportForm.serverIds"
                multiple
                filterable
                placeholder="请选择要生成报告的主机"
                style="width: 100%"
              >
                <ElOption
                  v-for="server in servers"
                  :key="server.id"
                  :label="`${server.hostname} (${server.ip})`"
                  :value="server.id"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>

          <ElCol :xs="24">
            <ElFormItem>
              <ElButton
                type="primary"
                :loading="generating"
                :disabled="reportForm.serverIds.length === 0"
                @click="generateReport"
              >
                <ElIcon :size="16"><Document /></ElIcon>
                {{ generating ? '生成中...' : '生成报告' }}
              </ElButton>
            </ElFormItem>
          </ElCol>
        </ElRow>
      </ElForm>
    </ElCard>

    <!-- 报告历史 -->
    <ElCard shadow="never" header="报告历史">
      <ElTable v-loading="loading" :data="reportHistory" border stripe>
        <ElTableColumn label="报告类型" width="140">
          <template #default="{ row }">
            {{ reportTypes.find(t => t.value === row.type)?.label || row.type }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="报告标题" prop="title" min-width="260" show-overflow-tooltip />
        <ElTableColumn label="状态" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </ElTableColumn>
        <ElTableColumn label="完成时间" width="180">
          <template #default="{ row }">{{ formatTime(row.completedAt) }}</template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" link size="small" @click="viewReport(row)">查看</ElButton>
            <ElButton type="success" link size="small" @click="exportReport(row)">导出</ElButton>
            <ElButton type="danger" link size="small" @click="deleteReport(row.id)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="reportHistory.length === 0" description="暂无报告历史" />
    </ElCard>

    <!-- 报告说明 -->
    <ElCard shadow="never" header="报告说明">
      <ElDescriptions :column="2" border>
        <ElDescriptionsItem label="主机巡检报告">
          包含主机基本信息、性能指标、服务状态、进程信息、硬件资产等全面检查
        </ElDescriptionsItem>
        <ElDescriptionsItem label="安全合规报告">
          检查SSH配置、防火墙状态、用户权限、登录历史等安全相关配置
        </ElDescriptionsItem>
        <ElDescriptionsItem label="容量分析报告">
          分析CPU、内存、磁盘使用趋势，预测容量需求，提供扩容建议
        </ElDescriptionsItem>
        <ElDescriptionsItem label="导出格式">支持 PDF、Excel、HTML 三种格式导出</ElDescriptionsItem>
      </ElDescriptions>
    </ElCard>
  </div>
</template>
