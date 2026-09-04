<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { DiagnosticFinding } from '../finding/types';

  defineOptions({ name: 'DiagnosticFindingResult' });

  const props = defineProps<{
    finding: DiagnosticFinding;
    raw: string;
  }>();

  const emit = defineEmits<{ openTerminal: [command: string] }>();

  const showRaw = ref(false);

  const LEVEL_TEXT: Record<DiagnosticFinding['level'], string> = {
    critical: '发现异常',
    warning: '需要关注',
    ok: '未见异常',
    info: '信息',
    unknown: '未判定'
  };
  const KIND_TEXT: Record<DiagnosticFinding['kind'], string> = {
    finding: '判定',
    describe: '信息',
    export: '产物',
    mutate: '变更'
  };

  const levelText = computed(() => LEVEL_TEXT[props.finding.level]);
  const kindText = computed(() => KIND_TEXT[props.finding.kind]);

  /** 堆栈帧列使用等宽字体 */
  function isFrameCol(sec: { columns?: string[] }, idx: number): boolean {
    const name = sec.columns?.[idx] ?? '';
    return /帧|栈|class|loader/i.test(name);
  }
</script>

<template>
  <div class="finding-result">
    <!-- 状态条（severity banner） -->
    <div class="f-banner" :class="[`level-${finding.level}`]">
      <span class="f-dot" />
      <span class="f-level-text">{{ levelText }}</span>
      <span class="f-kind">{{ kindText }}</span>
      <span class="f-symptom">{{ finding.parseFailed ? '结果未能结构化，请查看原始输出' : finding.symptom }}</span>
    </div>

    <!-- 叙事区：症状/根因/影响（判定型专用，Dynatrace 分诊流） -->
    <div v-if="finding.kind === 'finding' && !finding.parseFailed" class="f-narrative">
      <div v-if="finding.symptom" class="narr-row">
        <span class="narr-label">症状</span>
        <span class="narr-value">{{ finding.symptom }}</span>
      </div>
      <div v-if="finding.rootCause" class="narr-row root">
        <span class="narr-label">根因候选</span>
        <span class="narr-value mono">{{ finding.rootCause }}</span>
        <ElTag size="small" type="info" effect="plain">启发式 · 需人确认</ElTag>
      </div>
      <div v-if="finding.impact" class="narr-row">
        <span class="narr-label">影响</span>
        <span class="narr-value">{{ finding.impact }}</span>
      </div>
    </div>

    <!-- 证据/信息分区 -->
    <div v-for="(sec, i) in finding.sections" :key="i" class="f-section">
      <div v-if="sec.title" class="sec-title">{{ sec.title }}</div>

      <!-- 指标 + 水位进度条 -->
      <div v-if="sec.type === 'metrics'" class="metrics-grid">
        <div
          v-for="m in sec.metrics"
          :key="m.label"
          class="metric-item"
          :class="[{ warn: m.level === 'warning', crit: m.level === 'critical' }]"
        >
          <div class="metric-head">
            <span class="metric-label">{{ m.label }}</span>
            <span class="metric-value">{{ m.value }}</span>
          </div>
          <div v-if="m.percent !== undefined" class="metric-bar">
            <div
              class="metric-bar-fill"
              :style="{ width: `${Math.min(100, m.percent)}%` }"
              :class="{ warn: m.percent >= 70 && m.percent < 85, crit: m.percent >= 85 }"
            />
          </div>
        </div>
      </div>

      <!-- 表格（行级动作送终端；异常行高亮） -->
      <div v-else-if="sec.type === 'table'" class="table-wrap">
        <table class="f-table">
          <thead>
            <tr>
              <th v-for="c in sec.columns" :key="c">{{ c }}</th>
              <th v-if="sec.rows?.some(r => r.actions?.length)" class="op-col">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(r, ri) in sec.rows" :key="ri" :class="{ highlight: r.highlight }">
              <td v-for="(c, ci) in r.cells" :key="ci" :class="{ mono: isFrameCol(sec, ci) }">{{ c }}</td>
              <td v-if="sec.rows?.some(x => x.actions?.length)" class="op-col">
                <template v-if="r.note" class="row-note">{{ r.note }}</template>
                <ElButton
                  v-for="a in r.actions"
                  :key="a.command"
                  size="small"
                  text
                  type="primary"
                  @click="emit('openTerminal', a.command)"
                >
                  {{ a.label }}
                </ElButton>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分组 key-value（describe 风格） -->
      <div v-else-if="sec.type === 'kv'" class="kv-list">
        <div v-for="item in sec.kv" :key="item.key" class="kv-row">
          <span class="kv-key">{{ item.key }}</span>
          <span class="kv-value mono">{{ item.value || '-' }}</span>
        </div>
      </div>

      <!-- 代码块（堆栈等） -->
      <pre v-else-if="sec.type === 'code'" class="sec-code">{{ sec.code }}</pre>
    </div>

    <!-- 下一步动作（联动专家终端） -->
    <div v-if="finding.nextActions?.length" class="f-next">
      <span class="next-label">建议下一步</span>
      <ElButton
        v-for="a in finding.nextActions"
        :key="a.command"
        size="small"
        type="primary"
        plain
        @click="emit('openTerminal', a.command)"
      >
        {{ a.label }}
      </ElButton>
    </div>

    <!-- 原始输出（折叠兜底，永远可查可复制） -->
    <div class="f-raw-toggle">
      <ElButton text size="small" @click="showRaw = !showRaw">
        {{ showRaw ? '收起原始输出' : '展开原始输出' }}
      </ElButton>
    </div>
    <pre v-if="showRaw" class="f-raw">{{ raw }}</pre>
  </div>
</template>

<style scoped lang="scss">
  .finding-result {
    display: flex;
    flex-direction: column;
    gap: 10px;
    font-size: 12px;

    .f-banner {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 10px;
      border-radius: 6px;
      border: 1px solid transparent;

      .f-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        flex-shrink: 0;
      }

      .f-level-text {
        font-weight: 600;
        flex-shrink: 0;
      }

      .f-kind {
        flex-shrink: 0;
        font-size: 11px;
        padding: 0 6px;
        border-radius: 8px;
        background: var(--el-fill-color);
        color: var(--el-text-color-secondary);
      }

      .f-symptom {
        color: var(--el-text-color-regular);
        word-break: break-all;
      }

      &.level-critical {
        background: var(--el-color-danger-light-9);
        border-color: var(--el-color-danger-light-7);
        .f-dot {
          background: var(--el-color-danger);
        }
        .f-level-text {
          color: var(--el-color-danger);
        }
      }

      &.level-warning {
        background: var(--el-color-warning-light-9);
        border-color: var(--el-color-warning-light-7);
        .f-dot {
          background: var(--el-color-warning);
        }
        .f-level-text {
          color: var(--el-color-warning);
        }
      }

      &.level-ok {
        background: var(--el-color-success-light-9);
        border-color: var(--el-color-success-light-7);
        .f-dot {
          background: var(--el-color-success);
        }
        .f-level-text {
          color: var(--el-color-success);
        }
      }

      &.level-info,
      &.level-unknown {
        background: var(--el-fill-color-light);
        .f-dot {
          background: var(--el-text-color-secondary);
        }
        .f-level-text {
          color: var(--el-text-color-regular);
        }
      }
    }

    .f-narrative {
      display: flex;
      flex-direction: column;
      gap: 6px;

      .narr-row {
        display: flex;
        align-items: flex-start;
        gap: 8px;
        flex-wrap: wrap;

        .narr-label {
          flex-shrink: 0;
          width: 52px;
          color: var(--el-text-color-secondary);
          text-align: right;
        }

        .narr-value {
          color: var(--el-text-color-regular);
          word-break: break-all;

          &.mono {
            font-family: var(--el-font-family-mono, monospace);
            color: var(--el-color-primary);
          }
        }

        &.root .narr-value {
          color: var(--el-color-danger);
          font-weight: 600;
        }
      }
    }

    .f-section {
      .sec-title {
        font-weight: 600;
        color: var(--el-text-color-regular);
        margin-bottom: 6px;
        padding-left: 8px;
        border-left: 3px solid var(--el-color-primary-light-5);
      }
    }

    .metrics-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      gap: 8px;

      .metric-item {
        border: 1px solid var(--el-border-color-lighter);
        border-radius: 6px;
        padding: 8px 10px;

        &.warn {
          border-color: var(--el-color-warning-light-5);
        }
        &.crit {
          border-color: var(--el-color-danger-light-5);
        }

        .metric-head {
          display: flex;
          justify-content: space-between;
          gap: 6px;
          margin-bottom: 6px;

          .metric-label {
            color: var(--el-text-color-secondary);
          }
          .metric-value {
            font-weight: 500;
          }
        }

        .metric-bar {
          height: 6px;
          border-radius: 3px;
          background: var(--el-fill-color-darker);
          overflow: hidden;

          .metric-bar-fill {
            height: 100%;
            border-radius: 3px;
            background: var(--el-color-success);
            transition: width 0.3s;

            &.warn {
              background: var(--el-color-warning);
            }
            &.crit {
              background: var(--el-color-danger);
            }
          }
        }
      }
    }

    .table-wrap {
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 6px;
      overflow: auto;
      max-height: 300px;

      .f-table {
        width: 100%;
        border-collapse: collapse;
        table-layout: auto;
        white-space: nowrap;

        th,
        td {
          padding: 6px 10px;
          text-align: left;
          border-bottom: 1px solid var(--el-border-color-extra-light);
          max-width: 320px;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        th {
          background: var(--el-fill-color-light);
          color: var(--el-text-color-secondary);
          font-weight: 500;
          position: sticky;
          top: 0;
        }

        td.mono {
          font-family: var(--el-font-family-mono, monospace);
          font-size: 11px;
        }

        tr.highlight td {
          background: var(--el-color-danger-light-9);
        }

        .op-col {
          text-align: right;
          white-space: nowrap;
        }
      }
    }

    .kv-list {
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 6px;
      overflow: hidden;

      .kv-row {
        display: flex;
        gap: 10px;
        padding: 6px 10px;
        border-bottom: 1px solid var(--el-border-color-extra-light);

        &:nth-child(odd) {
          background: var(--el-fill-color-lighter);
        }
        &:last-child {
          border-bottom: none;
        }

        .kv-key {
          flex-shrink: 0;
          width: 140px;
          color: var(--el-text-color-secondary);
        }

        .kv-value {
          color: var(--el-text-color-regular);
          word-break: break-all;

          &.mono {
            font-size: 11px;
          }
        }
      }
    }

    .sec-code {
      background: #1e1e1e;
      color: #d4d4d4;
      padding: 10px;
      border-radius: 4px;
      font-size: 11px;
      line-height: 1.6;
      max-height: 260px;
      overflow: auto;
      white-space: pre-wrap;
      word-break: break-all;
      margin: 0;
    }

    .f-next {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
      padding: 8px 10px;
      background: var(--el-color-primary-light-9);
      border-radius: 6px;

      .next-label {
        color: var(--el-text-color-secondary);
        flex-shrink: 0;
      }
    }

    .f-raw-toggle {
      display: flex;

      :deep(.el-button) {
        padding: 4px 0;
      }
    }

    .f-raw {
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
</style>
