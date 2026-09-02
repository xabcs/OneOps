<script setup lang="ts">
  import { computed, reactive, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Cpu, DataLine, Link, MagicStick, VideoPlay } from '@element-plus/icons-vue';
  import { executeDiagnosticOneShot, type DiagnosticOneShotResult } from '@/service/api/diagnostic';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import { formatDuration, isHighRisk, riskMeta } from '../shared';

  defineOptions({ name: 'DiagnosticScenarioPanel' });

  const emit = defineEmits<{ openTerminal: [command: string] }>();

  const diagStore = useDiagnosticStore();

  // ========== 场景库（高频诊断场景封装） ==========
  interface ScenarioField {
    key: string;
    label: string;
    placeholder: string;
    defaultValue?: string;
  }

  interface Scenario {
    id: string;
    group: string;
    name: string;
    description: string;
    riskLevel: string;
    streaming: boolean; // 流式命令须在终端中使用
    fields?: ScenarioField[];
    build: (v: Record<string, string>) => string;
  }

  const classMethodFields: ScenarioField[] = [
    { key: 'class', label: '类全名', placeholder: '如 com.demo.OrderService' },
    { key: 'method', label: '方法名', placeholder: '如 createOrder' }
  ];

  const scenarios: Scenario[] = [
    // --- 性能诊断 ---
    {
      id: 'cpu-top',
      group: '性能诊断',
      name: 'CPU 飙升定位',
      description: '查看 CPU 占用最高的 5 个线程，定位热点代码（含 nid 可关联 jstack）',
      riskLevel: 'L1',
      streaming: false,
      fields: [{ key: 'n', label: 'Top N', placeholder: '5', defaultValue: '5' }],
      build: v => `thread -n ${v.n || '5'}`
    },
    {
      id: 'deadlock',
      group: '性能诊断',
      name: '死锁 / 阻塞检测',
      description: '找出阻塞其他线程最多的线程，快速定位死锁与长时间持锁',
      riskLevel: 'L1',
      streaming: false,
      build: () => 'thread --blocked-thread-locks'
    },
    {
      id: 'thread-all',
      group: '性能诊断',
      name: '线程全景',
      description: '列出全部线程状态与 CPU 占比',
      riskLevel: 'L1',
      streaming: false,
      build: () => 'thread'
    },
    {
      id: 'dashboard',
      group: '性能诊断',
      name: '实时运行面板',
      description: '线程/内存/GC 实时刷新面板（流式输出）',
      riskLevel: 'L1',
      streaming: true,
      build: () => 'dashboard'
    },
    {
      id: 'profiler',
      group: '性能诊断',
      name: '生成火焰图',
      description: 'CPU 采样 30 秒生成火焰图（async-profiler）',
      riskLevel: 'L2',
      streaming: true,
      fields: [{ key: 'duration', label: '采样秒数', placeholder: '30', defaultValue: '30' }],
      build: v => `profiler start --duration ${v.duration || '30'}`
    },
    // --- 内存诊断 ---
    {
      id: 'memory',
      group: '内存诊断',
      name: '内存分区占用',
      description: '查看 Eden/Survivor/Old/Metaspace 各区使用情况',
      riskLevel: 'L1',
      streaming: false,
      build: () => 'memory'
    },
    {
      id: 'jvm',
      group: '内存诊断',
      name: 'JVM 全景信息',
      description: '内存池、GC、类加载、线程、操作系统全量信息',
      riskLevel: 'L1',
      streaming: false,
      build: () => 'jvm'
    },
    {
      id: 'heapdump',
      group: '内存诊断',
      name: '堆转储（大对象）',
      description: '导出 live 对象堆转储，用于离线 MAT 分析；会触发 Full GC，生产慎用',
      riskLevel: 'L3',
      streaming: false,
      fields: [{ key: 'file', label: '导出路径', placeholder: '/tmp/heap.hprof' }],
      build: v => `heapdump ${v.file ? `--live ${v.file}` : '/tmp/heap.hprof'}`
    },
    // --- 调用链诊断 ---
    {
      id: 'trace',
      group: '调用链诊断',
      name: '方法耗时追踪',
      description: '追踪方法调用链各级耗时，定位慢调用环节',
      riskLevel: 'L2',
      streaming: true,
      fields: [
        ...classMethodFields,
        { key: 'cost', label: '耗时阈值(ms)', placeholder: '如 100，留空不过滤' }
      ],
      build: v => `trace ${v.class} ${v.method}${v.cost ? ` '#cost > ${v.cost}'` : ''} -n 5`.trim()
    },
    {
      id: 'watch',
      group: '调用链诊断',
      name: '观察方法出入参',
      description: '查看方法调用时的入参、返回值与异常（-n 5 限量）',
      riskLevel: 'L2',
      streaming: true,
      fields: [...classMethodFields, { key: 'expr', label: '观察表达式', placeholder: 'params,returnObj', defaultValue: 'params,returnObj' }],
      build: v => `watch ${v.class} ${v.method} "${v.expr || 'params,returnObj'}" -x 2 -n 5`
    },
    {
      id: 'tt',
      group: '调用链诊断',
      name: '时间隧道（录制回放）',
      description: '录制方法调用，可事后回放与查看入参',
      riskLevel: 'L2',
      streaming: true,
      fields: classMethodFields,
      build: v => `tt -t ${v.class} ${v.method} -n 5`
    },
    {
      id: 'stack',
      group: '调用链诊断',
      name: '调用栈查看',
      description: '输出当前调用该类方法的调用栈，定位调用来源',
      riskLevel: 'L2',
      streaming: true,
      fields: classMethodFields,
      build: v => `stack ${v.class} ${v.method} -n 3`
    },
    // --- 类与日志 ---
    {
      id: 'sc',
      group: '类与加载',
      name: '类加载检索',
      description: '检索 JVM 已加载的类（含子类/接口实现），排查类冲突',
      riskLevel: 'L1',
      streaming: false,
      fields: [{ key: 'pattern', label: '类名模式', placeholder: '如 *OrderService*' }],
      build: v => `sc -d ${v.pattern}`
    },
    {
      id: 'logger',
      group: '类与加载',
      name: '调整日志级别',
      description: '运行时修改 logger 级别（如临时开 DEBUG 排障）',
      riskLevel: 'L4',
      streaming: false,
      fields: [
        { key: 'name', label: 'Logger 名', placeholder: '如 com.demo 或 ROOT' },
        { key: 'level', label: '目标级别', placeholder: 'DEBUG' }
      ],
      build: v => `logger --name ${v.name || 'ROOT'} --level ${v.level || 'DEBUG'}`
    }
  ];

  const scenarioGroups = computed(() => {
    const groups: Record<string, Scenario[]> = {};
    scenarios.forEach(s => {
      groups[s.group] = groups[s.group] ?? [];
      groups[s.group].push(s);
    });
    return Object.entries(groups);
  });

  // ========== 选中场景 + 参数 ==========
  const selectedScenario = ref<Scenario | null>(null);
  const fieldValues = reactive<Record<string, Record<string, string>>>({});

  const scenarioIcons: Record<string, typeof Cpu> = {
    性能诊断: Cpu,
    内存诊断: DataLine,
    调用链诊断: Link,
    类与加载: MagicStick
  };

  function selectScenario(s: Scenario) {
    selectedScenario.value = s;
    fieldValues[s.id] = fieldValues[s.id] ?? {};
    s.fields?.forEach(f => {
      if (fieldValues[s.id][f.key] == null) {
        fieldValues[s.id][f.key] = f.defaultValue ?? '';
      }
    });
  }

  // ========== 执行 ==========
  const executing = ref(false);
  const result = ref<DiagnosticOneShotResult | null>(null);
  const resultCommand = ref('');

  async function run(scenario: Scenario) {
    const agent = diagStore.currentAgent;
    if (!agent) {
      ElMessage.warning('请先在左侧选择诊断实例');
      return;
    }
    if (!agent.online) {
      ElMessage.error('该实例 Agent 离线，无法诊断');
      return;
    }

    const command = scenario.build(fieldValues[scenario.id] ?? {});

    // 流式命令 → 专家终端
    if (scenario.streaming) {
      emit('openTerminal', command);
      return;
    }

    // 高风险强确认
    const meta = riskMeta(scenario.riskLevel);
    if (isHighRisk(scenario.riskLevel)) {
      try {
        await ElMessageBox.confirm(
          `「${scenario.name}」为 ${meta.label}：${meta.desc}。确认在 ${agent.podName || agent.agentId} 上执行？`,
          '高危操作确认',
          { type: 'warning', confirmButtonText: '确认执行', cancelButtonText: '取消' }
        );
      } catch {
        return;
      }
    }

    executing.value = true;
    result.value = null;
    try {
      const { data, error } = await executeDiagnosticOneShot({
        clusterId: agent.clusterId,
        agentId: agent.agentId,
        command,
        timeout: 90
      });
      if (!error && data) {
        result.value = data;
        resultCommand.value = command;
      }
    } finally {
      executing.value = false;
    }
  }

  function copyOutput() {
    if (result.value?.output) {
      navigator.clipboard.writeText(result.value.output);
      ElMessage.success('已复制');
    }
  }
</script>

<template>
  <div class="scenario-panel">
    <!-- 场景分组卡片 -->
    <div class="scenario-groups" :class="{ condensed: selectedScenario }">
      <div v-for="[group, items] in scenarioGroups" :key="group" class="group-block">
        <div class="group-title">
          <ElIcon :size="14"><component :is="scenarioIcons[group] ?? MagicStick" /></ElIcon>
          {{ group }}
        </div>
        <div class="group-cards">
          <div
            v-for="s in items"
            :key="s.id"
            :class="['scenario-card', { selected: selectedScenario?.id === s.id }]"
            @click="selectScenario(s)"
          >
            <div class="card-head">
              <span class="card-name">{{ s.name }}</span>
              <ElTag size="small" :type="riskMeta(s.riskLevel).type">{{ s.riskLevel }}</ElTag>
            </div>
            <div class="card-desc">{{ s.description }}</div>
            <div v-if="s.streaming" class="card-stream">流式 · 在终端执行</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 参数与执行 -->
    <div class="scenario-detail">
      <template v-if="selectedScenario">
        <div class="detail-head">
          <span class="detail-name">{{ selectedScenario.name }}</span>
          <ElTag size="small" :type="riskMeta(selectedScenario.riskLevel).type">
            {{ riskMeta(selectedScenario.riskLevel).label }}
          </ElTag>
        </div>

        <ElForm v-if="selectedScenario.fields?.length" label-position="top" size="small" class="detail-form">
          <ElFormItem v-for="f in selectedScenario.fields" :key="f.key" :label="f.label">
            <ElInput
              v-model="fieldValues[selectedScenario.id][f.key]"
              :placeholder="f.placeholder"
              clearable
            />
          </ElFormItem>
        </ElForm>

        <div class="detail-command">
          <span class="cmd-label">将执行</span>
          <code>{{ selectedScenario.build(fieldValues[selectedScenario.id] ?? {}) }}</code>
        </div>

        <PermissionButton
          code="k8s.diagnostic.execute"
          type="primary"
          :icon="VideoPlay"
          :loading="executing"
          @click="run(selectedScenario)"
        >
          {{ selectedScenario.streaming ? '在专家终端中打开' : '执行诊断' }}
        </PermissionButton>

        <!-- 结果 -->
        <div v-if="result" class="detail-result">
          <div class="result-head">
            <span>执行结果</span>
            <span class="result-meta">
              <ElTag size="small">{{ result.riskLevel }}</ElTag>
              <ElTag size="small" type="info">{{ formatDuration(result.duration) }}</ElTag>
              <ElButton size="small" text type="primary" @click="copyOutput">复制</ElButton>
            </span>
          </div>
          <pre class="result-output">{{ result.output || result.error }}</pre>
        </div>
      </template>

      <ElEmpty v-else description="选择左侧诊断场景" :image-size="80" class="detail-empty" />
    </div>
  </div>
</template>

<style scoped lang="scss">
  .scenario-panel {
    display: flex;
    height: 100%;
    min-height: 0;

    .scenario-groups {
      flex: 1;
      min-width: 0;
      overflow-y: auto;
      padding: 12px;

      &.condensed {
        max-width: 62%;
      }

      .group-block {
        margin-bottom: 14px;

        .group-title {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          font-size: 13px;
          font-weight: 600;
          color: var(--el-text-color-regular);
          margin-bottom: 8px;
        }

        .group-cards {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
          gap: 8px;

          .scenario-card {
            border: 1px solid var(--el-border-color-light);
            border-radius: 6px;
            padding: 10px 12px;
            cursor: pointer;
            transition: all 0.2s;

            &:hover {
              border-color: var(--el-color-primary-light-5);
            }

            &.selected {
              border-color: var(--el-color-primary);
              background: var(--el-color-primary-light-9);
            }

            .card-head {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 4px;

              .card-name {
                font-size: 13px;
                font-weight: 500;
              }
            }

            .card-desc {
              font-size: 12px;
              color: var(--el-text-color-secondary);
              line-height: 1.5;
            }

            .card-stream {
              margin-top: 6px;
              font-size: 11px;
              color: var(--el-color-warning);
            }
          }
        }
      }
    }

    .scenario-detail {
      width: 34%;
      min-width: 300px;
      border-left: 1px solid var(--el-border-color-lighter);
      padding: 12px 14px;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      gap: 12px;

      .detail-head {
        display: flex;
        align-items: center;
        gap: 8px;

        .detail-name {
          font-size: 14px;
          font-weight: 600;
        }
      }

      .detail-form {
        :deep(.el-form-item) {
          margin-bottom: 10px;
        }
      }

      .detail-command {
        display: flex;
        align-items: center;
        gap: 8px;
        background: var(--el-fill-color-darker);
        border-radius: 4px;
        padding: 8px 10px;
        font-size: 12px;
        flex-wrap: wrap;

        .cmd-label {
          color: var(--el-text-color-secondary);
          flex-shrink: 0;
        }

        code {
          color: var(--el-color-primary);
          word-break: break-all;
        }
      }

      .detail-result {
        display: flex;
        flex-direction: column;
        gap: 8px;

        .result-head {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 13px;
          font-weight: 500;

          .result-meta {
            display: inline-flex;
            align-items: center;
            gap: 6px;
          }
        }

        .result-output {
          background: #1e1e1e;
          color: #d4d4d4;
          padding: 12px;
          border-radius: 4px;
          font-size: 12px;
          line-height: 1.6;
          max-height: 320px;
          overflow: auto;
          white-space: pre-wrap;
          word-break: break-all;
          margin: 0;
        }
      }

      .detail-empty {
        margin: auto;
      }
    }
  }
</style>
