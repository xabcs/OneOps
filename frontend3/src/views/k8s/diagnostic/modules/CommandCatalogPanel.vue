<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Search, VideoPlay } from '@element-plus/icons-vue';
  import { type DiagnosticOneShotResult, executeDiagnosticOneShot } from '@/service/api/diagnostic';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import { formatDuration, isHighRisk, riskMeta } from '../shared';

  defineOptions({ name: 'DiagnosticCommandCatalogPanel' });

  const emit = defineEmits<{ openTerminal: [command: string] }>();

  const diagStore = useDiagnosticStore();

  // ========== 筛选 ==========
  const search = ref('');
  const category = ref<string>(''); // '' 全部
  const riskFilter = ref<string>('');

  const categories = computed(() => {
    const set = new Set(diagStore.commandCatalog.map(c => c.category));
    return Array.from(set).sort();
  });

  const filteredCatalog = computed(() => {
    const kw = search.value.trim().toLowerCase();
    return diagStore.commandCatalog.filter(c => {
      if (category.value && c.category !== category.value) return false;
      if (riskFilter.value && c.riskLevel !== riskFilter.value) return false;
      if (kw) {
        return (
          c.command.toLowerCase().includes(kw) ||
          c.description.toLowerCase().includes(kw) ||
          (c.usage ?? '').toLowerCase().includes(kw)
        );
      }
      return true;
    });
  });

  // ========== 执行 ==========
  const executingCommand = ref('');
  const activeCommand = ref('');
  const result = ref<DiagnosticOneShotResult | null>(null);
  const resultCommand = ref('');
  /** 执行目标快照：结果归属于执行时的实例，切换左侧目标不改变已有结果归属 */
  const resultTarget = ref('');

  async function runCommand(item: { command: string; riskLevel: string; streaming: boolean }) {
    const agent = diagStore.currentAgent;
    if (!agent) {
      ElMessage.warning('请先在左侧选择诊断实例');
      return;
    }
    if (!agent.online) {
      ElMessage.error('该实例 Agent 离线，无法诊断');
      return;
    }
    if (item.riskLevel === 'disabled') {
      ElMessage.error('该命令已被管理员禁用');
      return;
    }

    let command = item.command;
    if (!item.streaming) {
      // 非流式命令允许带参数执行
      const { value } = await ElMessageBox.prompt(
        `输入 ${item.command} 的参数（可留空直接执行）`,
        `执行 ${item.command}`,
        {
          inputPlaceholder: '如：-n 5 / 类名 / 表达式',
          confirmButtonText: '执行',
          cancelButtonText: '取消'
        }
      ).catch(() => ({ value: undefined as string | undefined }));
      if (value === undefined) return;
      command = `${item.command} ${value.trim()}`.trim();
    } else {
      emit('openTerminal', command);
      return;
    }

    if (isHighRisk(item.riskLevel)) {
      const meta = riskMeta(item.riskLevel);
      // stop 的后果与通用风险描述不同，专项提示（agent 永久离线，Pod 重建前不可再诊断）
      const extra =
        item.command === 'stop'
          ? '。stop 会关闭 Arthas：该实例 Agent 将永久离线、无法重连，需重启 Pod 恢复；若只想清理增强请改用 reset'
          : '';
      try {
        await ElMessageBox.confirm(
          `${meta.label}：${meta.desc}${extra}。确认在 ${agent.podName || agent.agentId} 上执行「${command}」？`,
          '高危操作确认',
          { type: 'warning', confirmButtonText: '确认执行', cancelButtonText: '取消' }
        );
      } catch {
        return;
      }
    }

    executingCommand.value = item.command;
    activeCommand.value = command;
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
        resultTarget.value = agent.podName || agent.agentId;
      }
    } finally {
      executingCommand.value = '';
    }
  }
</script>

<template>
  <div class="catalog-panel">
    <!-- 筛选条 -->
    <div class="catalog-filter">
      <ElInput
        v-model="search"
        placeholder="搜索命令 / 说明 / 用法"
        clearable
        :prefix-icon="Search"
        class="filter-search"
      />
      <ElSelect v-model="category" placeholder="全部分类" clearable class="filter-category">
        <ElOption v-for="c in categories" :key="c" :label="c" :value="c" />
      </ElSelect>
      <ElSelect v-model="riskFilter" placeholder="全部风险级" clearable class="filter-risk">
        <ElOption
          v-for="level in ['L0', 'L1', 'L2', 'L3', 'L4', 'L5', 'disabled']"
          :key="level"
          :label="riskMeta(level).label"
          :value="level"
        />
      </ElSelect>
      <span class="filter-count">共 {{ filteredCatalog.length }} 条</span>
    </div>

    <!-- 目录表格 -->
    <ElTable :data="filteredCatalog" size="small" height="100%" class="catalog-table">
      <ElTableColumn prop="command" label="命令" width="130" fixed>
        <template #default="{ row }">
          <code class="cmd-code">{{ row.command }}</code>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="category" label="分类" width="100" />
      <ElTableColumn prop="description" label="说明" min-width="260" show-overflow-tooltip />
      <ElTableColumn label="风险级" width="110">
        <template #default="{ row }">
          <ElTooltip :content="riskMeta(row.riskLevel).desc" placement="top">
            <ElTag size="small" :type="riskMeta(row.riskLevel).type">{{ riskMeta(row.riskLevel).label }}</ElTag>
          </ElTooltip>
        </template>
      </ElTableColumn>
      <ElTableColumn label="输出" width="80">
        <template #default="{ row }">
          <ElTag v-if="row.streaming" size="small" type="warning" effect="plain">流式</ElTag>
          <ElTag v-else size="small" type="info" effect="plain">一次性</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="usage" label="用法示例" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="usage-code">{{ row.usage }}</code>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" align="center" width="110" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <PermissionButton
            link
            type="primary"
            size="small"
            code="k8s.diagnostic.execute"
            :icon="VideoPlay"
            :loading="executingCommand === row.command"
            @click="runCommand(row)"
          >
            {{ row.streaming ? '终端' : '执行' }}
          </PermissionButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 最近执行结果 -->
    <div v-if="result" class="catalog-result">
      <div class="result-head">
        <code>{{ resultCommand }}</code>
        <span class="result-meta">
          <ElTag v-if="resultTarget" size="small" type="warning">目标: {{ resultTarget }}</ElTag>
          <ElTag size="small" :type="result.status === 'success' ? 'success' : 'danger'">{{ result.status }}</ElTag>
          <ElTag size="small" type="info">{{ formatDuration(result.duration) }}</ElTag>
          <ElButton size="small" text type="primary" @click="result = null">关闭</ElButton>
        </span>
      </div>
      <pre class="result-output">{{ result.output || result.error }}</pre>
    </div>
  </div>
</template>

<style scoped lang="scss">
  .catalog-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    padding: 12px;
    gap: 10px;

    .catalog-filter {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-shrink: 0;

      .filter-search {
        width: 260px;
      }

      .filter-category,
      .filter-risk {
        width: 140px;
      }

      .filter-count {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        margin-left: auto;
      }
    }

    .catalog-table {
      flex: 1;
      min-height: 0;

      .cmd-code {
        color: var(--el-color-primary);
        font-weight: 600;
      }

      .usage-code {
        color: var(--el-text-color-secondary);
      }
    }

    .catalog-result {
      flex-shrink: 0;
      max-height: 280px;
      display: flex;
      flex-direction: column;
      gap: 6px;

      .result-head {
        display: flex;
        justify-content: space-between;
        align-items: center;

        code {
          color: var(--el-color-primary);
          font-size: 12px;
        }

        .result-meta {
          display: inline-flex;
          align-items: center;
          gap: 6px;
        }
      }

      .result-output {
        flex: 1;
        overflow: auto;
        background: #1e1e1e;
        color: #d4d4d4;
        padding: 12px;
        border-radius: 4px;
        font-size: 12px;
        line-height: 1.6;
        white-space: pre-wrap;
        word-break: break-all;
        margin: 0;
      }
    }
  }
</style>
