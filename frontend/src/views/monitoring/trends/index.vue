<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Refresh } from '@element-plus/icons-vue';
import * as echarts from 'echarts';
import { fetchGetServers, fetchServerMetricsHistory } from '@/service/api';
import type { MetricsDatapoint } from '@/service/api/monitoring';

defineOptions({
  name: 'MonitoringTrends'
});

const route = useRoute();
const loading = ref(false);
const chartLoading = ref(false);

// 时间范围选项
const timeRangeOptions = [
  { label: '近1小时', value: '1h' },
  { label: '近6小时', value: '6h' },
  { label: '近24小时', value: '24h' },
  { label: '近7天', value: '7d' },
  { label: '近30天', value: '30d' }
];

const filterForm = reactive({
  serverId: null as number | null,
  metricType: 'cpu',
  timeRange: '24h',
  customStartTime: '',
  customEndTime: ''
});

const servers = ref<CMDB.Server[]>([]);
const chartData = ref<MetricsDatapoint[]>([]);
const timeRange = ref('24h');
const customDateRange = ref<[Date, Date] | null>(null);

// ECharts 图表相关
const chartRef = ref<HTMLElement>();
let chartInstance: echarts.ECharts | null = null;

// 计算时间范围
function calculateTimeRange(range: string): { start: Date; end: Date } {
  const end = new Date();
  const start = new Date();

  switch (range) {
    case '1h':
      start.setHours(start.getHours() - 1);
      break;
    case '6h':
      start.setHours(start.getHours() - 6);
      break;
    case '24h':
      start.setHours(start.getHours() - 24);
      break;
    case '7d':
      start.setDate(start.getDate() - 7);
      break;
    case '30d':
      start.setDate(start.getDate() - 30);
      break;
    default:
      start.setHours(start.getHours() - 24);
  }

  return { start, end };
}

// 格式化时间为ISO字符串
function formatToISOString(date: Date): string {
  return date.toISOString();
}

// 获取主机列表
async function getServerList() {
  loading.value = true;
  const { data } = await fetchGetServers({ page: 1, pageSize: 1000, agentStatus: 'running' });
  if (data) {
    servers.value = data.list || [];
    // 如果URL中有serverId，则选中该主机
    const queryServerId = route.query.serverId;
    if (queryServerId) {
      filterForm.serverId = Number(queryServerId);
    } else if (servers.value.length > 0) {
      filterForm.serverId = servers.value[0].id;
    }
  }
  loading.value = false;
}

// 加载趋势数据
async function loadTrendData() {
  if (!filterForm.serverId) {
    return;
  }

  chartLoading.value = true;
  try {
    const { start, end } = calculateTimeRange(timeRange.value);

    const { data } = await fetchServerMetricsHistory(filterForm.serverId, {
      metricType: filterForm.metricType,
      startTime: formatToISOString(start),
      endTime: formatToISOString(end),
      interval:
        timeRange.value === '1h' ? '5m' : timeRange.value === '6h' ? '10m' : timeRange.value === '24h' ? '30m' : '1h'
    });

    if (data) {
      chartData.value = data.datapoints || [];
    }
  } catch (error) {
    console.error('加载趋势数据失败:', error);
    chartData.value = [];
  }
  chartLoading.value = false;
}

// 加载自定义时间范围数据
async function loadCustomRangeData() {
  if (!filterForm.serverId || !customDateRange.value) {
    return;
  }

  chartLoading.value = true;
  try {
    const { data } = await fetchServerMetricsHistory(filterForm.serverId, {
      metricType: filterForm.metricType,
      startTime: formatToISOString(customDateRange.value[0]),
      endTime: formatToISOString(customDateRange.value[1]),
      interval: '1h'
    });

    if (data) {
      chartData.value = data.datapoints || [];
    }
  } catch (error) {
    console.error('加载趋势数据失败:', error);
    chartData.value = [];
  }
  chartLoading.value = false;
}

// 时间范围变化
function handleTimeRangeChange(range: string) {
  timeRange.value = range;
  customDateRange.value = null;
  loadTrendData();
}

// 自定义时间范围变化
function handleCustomDateChange() {
  if (customDateRange.value && customDateRange.value.length === 2) {
    loadCustomRangeData();
  }
}

// 指标类型变化
function handleMetricTypeChange() {
  if (customDateRange.value) {
    loadCustomRangeData();
  } else {
    loadTrendData();
  }
}

// 主机变化
function handleServerChange() {
  if (customDateRange.value) {
    loadCustomRangeData();
  } else {
    loadTrendData();
  }
}

// 获取指标类型名称
function getMetricTypeName(type: string): string {
  const map: Record<string, string> = {
    cpu: 'CPU使用率',
    memory: '内存使用率',
    disk: '磁盘使用率',
    load1: '1分钟负载',
    load5: '5分钟负载'
  };
  return map[type] || type;
}

// 获取指标单位
function getMetricUnit(type: string): string {
  return ['cpu', 'memory', 'disk'].includes(type) ? '%' : '';
}

// 格式化图表数据
const formattedChartData = computed(() => {
  return chartData.value.map(item => ({
    time: item.timestamp,
    value: typeof item.value === 'number' ? item.value : Number.parseFloat(String(item.value)) || 0
  }));
});

// 计算统计信息
const statistics = computed(() => {
  if (chartData.value.length === 0) {
    return { min: 0, max: 0, avg: 0, current: 0 };
  }

  const values = chartData.value
    .map(item => (typeof item.value === 'number' ? item.value : Number.parseFloat(String(item.value)) || 0))
    .filter(v => !isNaN(v));

  if (values.length === 0) {
    return { min: 0, max: 0, avg: 0, current: 0 };
  }

  const min = Math.min(...values);
  const max = Math.max(...values);
  const sum = values.reduce((a, b) => a + b, 0);
  const avg = sum / values.length;
  const current = values[values.length - 1];

  return { min, max, avg, current };
});

// 刷新数据
function handleRefresh() {
  if (customDateRange.value) {
    loadCustomRangeData();
  } else {
    loadTrendData();
  }
}

// 初始化图表
function initChart() {
  if (!chartRef.value) return;

  chartInstance = echarts.init(chartRef.value);
  updateChart();

  // 响应式调整
  window.addEventListener('resize', () => {
    chartInstance?.resize();
  });
}

// 更新图表
function updateChart() {
  if (!chartInstance) return;

  const times = chartData.value.map(item => {
    const date = new Date(item.timestamp);
    return `${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
  });

  const values = chartData.value.map(item =>
    typeof item.value === 'number' ? item.value : Number.parseFloat(String(item.value)) || 0
  );

  const option: echarts.EChartsOption = {
    title: {
      text: getMetricTypeName(filterForm.metricType),
      left: 'center',
      textStyle: {
        fontSize: 16,
        fontWeight: 600
      }
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        if (!params || params.length === 0) return '';
        const param = params[0];
        return `${param.name}<br/>${param.seriesName}: ${param.value.toFixed(2)}${getMetricUnit(filterForm.metricType)}`;
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        rotate: times.length > 20 ? 45 : 0,
        formatter: (value: string) => {
          // 根据数据量调整显示格式
          if (times.length > 100) {
            return value.split(' ')[0]; // 只显示日期
          }
          return value;
        }
      }
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: ['cpu', 'memory', 'disk'].includes(filterForm.metricType) ? 100 : undefined,
      axisLabel: {
        formatter: `{value}${getMetricUnit(filterForm.metricType)}`
      }
    },
    series: [
      {
        name: getMetricTypeName(filterForm.metricType),
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: values,
        lineStyle: {
          width: 2,
          color: '#409EFF'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        },
        markLine: {
          silent: true,
          lineStyle: {
            color: '#ff4d4f',
            type: 'dashed'
          },
          data: [
            {
              yAxis: 80,
              label: {
                formatter: '警告线 (80%)'
              }
            }
          ].filter(item => ['cpu', 'memory', 'disk'].includes(filterForm.metricType))
        }
      }
    ]
  };

  chartInstance.setOption(option, true);
}

let refreshTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  getServerList().then(() => {
    loadTrendData();
  });

  // 初始化图表
  nextTick(() => {
    initChart();
  });

  // 每30秒自动刷新
  refreshTimer = setInterval(() => {
    if (!customDateRange.value) {
      loadTrendData();
    }
  }, 30000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  if (chartInstance) {
    chartInstance.dispose();
  }
  window.removeEventListener('resize', () => {
    chartInstance?.resize();
  });
});

// 监听图表数据变化，自动更新图表
watch(
  [chartData, () => filterForm.metricType],
  () => {
    nextTick(() => {
      if (chartInstance) {
        updateChart();
      } else {
        initChart();
      }
    });
  },
  { deep: true }
);
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 标题栏 -->
    <ElCard shadow="never">
      <div class="flex items-center justify-between">
        <span class="text-lg font-semibold">趋势分析</span>
        <ElButton type="primary" :loading="chartLoading" @click="handleRefresh">
          <ElIcon :size="16"><Refresh /></ElIcon>
          刷新
        </ElButton>
      </div>
    </ElCard>

    <!-- 过滤栏 -->
    <ElCard shadow="never">
      <ElForm :model="filterForm" inline>
        <ElFormItem label="主机">
          <ElSelect
            v-model="filterForm.serverId"
            placeholder="选择主机"
            filterable
            style="width: 200px"
            @change="handleServerChange"
          >
            <ElOption
              v-for="server in servers"
              :key="server.id"
              :label="`${server.hostname} (${server.ip})`"
              :value="server.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="指标类型">
          <ElSelect v-model="filterForm.metricType" style="width: 130px" @change="handleMetricTypeChange">
            <ElOption label="CPU使用率" value="cpu" />
            <ElOption label="内存使用率" value="memory" />
            <ElOption label="磁盘使用率" value="disk" />
            <ElOption label="1分钟负载" value="load1" />
            <ElOption label="5分钟负载" value="load5" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="时间范围">
          <ElRadioGroup v-model="timeRange" @change="handleTimeRangeChange">
            <ElRadioButton label="1h">近1小时</ElRadioButton>
            <ElRadioButton label="6h">近6小时</ElRadioButton>
            <ElRadioButton label="24h">近24小时</ElRadioButton>
            <ElRadioButton label="7d">近7天</ElRadioButton>
            <ElRadioButton label="30d">近30天</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem label="自定义时间">
          <ElDatePicker
            v-model="customDateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DD HH:mm:ss"
            @change="handleCustomDateChange"
          />
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 统计信息 -->
    <ElRow v-if="chartData.length > 0" :gutter="16">
      <ElCol :xs="12" :sm="6">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-blue-600 font-bold">
            {{ statistics.current.toFixed(1) }}{{ getMetricUnit(filterForm.metricType) }}
          </div>
          <div class="mt-1 text-sm text-gray-500">当前值</div>
        </ElCard>
      </ElCol>
      <ElCol :xs="12" :sm="6">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-green-600 font-bold">
            {{ statistics.min.toFixed(1) }}{{ getMetricUnit(filterForm.metricType) }}
          </div>
          <div class="mt-1 text-sm text-gray-500">最小值</div>
        </ElCard>
      </ElCol>
      <ElCol :xs="12" :sm="6">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-red-600 font-bold">
            {{ statistics.max.toFixed(1) }}{{ getMetricUnit(filterForm.metricType) }}
          </div>
          <div class="mt-1 text-sm text-gray-500">最大值</div>
        </ElCard>
      </ElCol>
      <ElCol :xs="12" :sm="6">
        <ElCard shadow="never" class="text-center">
          <div class="text-2xl text-orange-600 font-bold">
            {{ statistics.avg.toFixed(1) }}{{ getMetricUnit(filterForm.metricType) }}
          </div>
          <div class="mt-1 text-sm text-gray-500">平均值</div>
        </ElCard>
      </ElCol>
    </ElRow>

    <!-- 图表 -->
    <ElCard shadow="never">
      <template #header>
        <span>{{ getMetricTypeName(filterForm.metricType) }}趋势图</span>
      </template>

      <ElSkeleton v-if="chartLoading" :rows="8" animated />

      <div v-else-if="chartData.length === 0" class="py-12 text-center">
        <ElEmpty description="暂无数据" />
      </div>

      <div v-else class="chart-container">
        <!-- ECharts 图表 -->
        <div ref="chartRef" class="echarts-chart"></div>

        <!-- 数据表格 -->
        <div class="mt-4">
          <ElTable :data="formattedChartData.slice(-10)" border size="small" max-height="300">
            <ElTableColumn label="时间" prop="time" width="180" />
            <ElTableColumn label="值" width="120">
              <template #default="{ row }">
                {{ row.value.toFixed(2) }}{{ getMetricUnit(filterForm.metricType) }}
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="100">
              <template #default="{ row }">
                <ElTag :type="row.value > 80 ? 'danger' : row.value > 60 ? 'warning' : 'success'" size="small">
                  {{ row.value > 80 ? '高' : row.value > 60 ? '中' : '正常' }}
                </ElTag>
              </template>
            </ElTableColumn>
          </ElTable>
          <div class="mt-2 text-center text-sm text-gray-400">仅显示最近10条数据，完整数据请查看上方图表</div>
        </div>
      </div>
    </ElCard>
  </div>
</template>

<style scoped>
.chart-container {
  min-height: 400px;
}

.echarts-chart {
  width: 100%;
  height: 400px;
}
</style>
