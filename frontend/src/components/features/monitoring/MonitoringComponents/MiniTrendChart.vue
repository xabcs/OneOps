<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import * as echarts from 'echarts';

interface Props {
  data?: number[];
  height?: number;
  color?: string;
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  height: 30,
  color: 'rgb(0, 82, 217)' // 腾讯云蓝
});

const chartRef = ref<HTMLElement>();
let chartInstance: echarts.ECharts | null = null;

function initChart() {
  if (!chartRef.value) return;

  try {
    chartInstance = echarts.init(chartRef.value, undefined, {
      renderer: 'svg',
      width: chartRef.value.clientWidth,
      height: props.height
    });

    updateChart();
  } catch (error) {
    console.error('[MiniTrendChart] 初始化图表失败:', error);
  }
}

function updateChart() {
  if (!chartInstance) return;

  try {
    const option: echarts.EChartsOption = {
      grid: {
        left: 0,
        right: 0,
        top: 0,
        bottom: 0
      },
      xAxis: {
        type: 'category',
        show: false,
        data: props.data.map((_, i) => i)
      },
      yAxis: {
        type: 'value',
        show: false,
        min: 0,
        max: 100
      },
      series: [
        {
          type: 'line',
          data: props.data,
          smooth: true,
          symbol: 'none',
          lineStyle: {
            color: props.color,
            width: 1.5
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: `${props.color}40` },
              { offset: 1, color: `${props.color}05` }
            ])
          }
        }
      ]
    };

    chartInstance.setOption(option);
  } catch (error) {
    console.error('[MiniTrendChart] 更新图表失败:', error);
  }
}

function handleResize() {
  chartInstance?.resize();
}

onMounted(() => {
  initChart();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  if (chartInstance) {
    chartInstance.dispose();
    chartInstance = null;
  }
});

watch(
  () => [props.data, props.color],
  () => {
    updateChart();
  },
  { deep: true }
);
</script>

<template>
  <div ref="chartRef" class="mini-trend-chart" :style="{ height: height + 'px' }"></div>
</template>

<style scoped>
.mini-trend-chart {
  width: 100%;
}
</style>
