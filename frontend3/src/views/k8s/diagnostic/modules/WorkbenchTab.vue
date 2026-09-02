<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { Grid, Monitor, Sort } from '@element-plus/icons-vue';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import ScenarioPanel from './ScenarioPanel.vue';
  import CommandCatalogPanel from './CommandCatalogPanel.vue';
  import ExpertTerminalPanel from './ExpertTerminalPanel.vue';

  defineOptions({ name: 'DiagnosticWorkbenchTab' });

  const diagStore = useDiagnosticStore();

  const subTab = ref('scenario');

  // 流式场景 → 切专家终端并自动下发命令
  const terminalCommand = ref('');
  function openTerminal(command: string) {
    terminalCommand.value = command;
    subTab.value = 'terminal';
  }

  // 左侧实例列表（来自上下文选择的应用）
  const agentSearch = ref('');
  const filteredAgents = computed(() => {
    const kw = agentSearch.value.trim().toLowerCase();
    if (!kw) return diagStore.agents;
    return diagStore.agents.filter(
      a => (a.podName || a.agentId).toLowerCase().includes(kw) || a.namespace.toLowerCase().includes(kw)
    );
  });

  const onlineCount = computed(() => diagStore.onlineAgents.length);
</script>

<template>
  <div class="workbench">
    <!-- 左侧：实例列表 -->
    <div class="agent-panel">
      <div class="panel-header">
        <span class="panel-title">实例列表</span>
        <ElTag size="small" type="success">{{ onlineCount }}/{{ diagStore.agents.length }} 在线</ElTag>
      </div>

      <div class="panel-search">
        <ElInput v-model="agentSearch" placeholder="搜索 Pod / 命名空间" clearable size="small" />
      </div>

      <div v-loading="diagStore.agentsLoading" class="agent-list">
        <div
          v-for="agent in filteredAgents"
          :key="agent.agentId"
          :class="['agent-item', { active: diagStore.currentAgent?.agentId === agent.agentId }]"
          @click="diagStore.selectAgent(agent)"
        >
          <div class="agent-main">
            <div class="agent-name" :title="agent.podName || agent.agentId">
              {{ agent.podName || agent.agentId }}
            </div>
            <div class="agent-meta">{{ agent.namespace }} · {{ agent.agentVersion || 'arthas' }}</div>
          </div>
          <span :class="['status-dot', { online: agent.online }]" />
        </div>

        <ElEmpty
          v-if="!diagStore.agentsLoading && filteredAgents.length === 0"
          :description="diagStore.currentApp ? '该应用暂无实例' : '请先在上方选择应用'"
          :image-size="70"
        />
      </div>
    </div>

    <!-- 右侧：诊断能力面板（场景 / 目录 / 终端） -->
    <div class="capability-panel">
      <ElTabs v-model="subTab" class="sub-tabs">
        <ElTabPane name="scenario">
          <template #label>
            <span class="sub-tab-label"><ElIcon :size="14"><Monitor /></ElIcon> 场景诊断</span>
          </template>
        </ElTabPane>
        <ElTabPane name="catalog">
          <template #label>
            <span class="sub-tab-label"><ElIcon :size="14"><Sort /></ElIcon> 命令目录</span>
          </template>
        </ElTabPane>
        <ElTabPane name="terminal">
          <template #label>
            <span class="sub-tab-label"><ElIcon :size="14"><Grid /></ElIcon> 专家终端</span>
          </template>
        </ElTabPane>
      </ElTabs>

      <div class="sub-tab-body">
        <ScenarioPanel v-show="subTab === 'scenario'" @open-terminal="openTerminal" />
        <CommandCatalogPanel v-show="subTab === 'catalog'" @open-terminal="openTerminal" />
        <ExpertTerminalPanel v-show="subTab === 'terminal'" v-model:auto-command="terminalCommand" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
  .workbench {
    display: flex;
    gap: 12px;
    height: 100%;
    min-height: 0;

    .agent-panel {
      width: 260px;
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      overflow: hidden;

      .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 12px;
        border-bottom: 1px solid var(--el-border-color-lighter);

        .panel-title {
          font-weight: 600;
          font-size: 13px;
        }
      }

      .panel-search {
        padding: 8px 10px;
      }

      .agent-list {
        flex: 1;
        overflow-y: auto;
        padding: 4px 8px 8px;

        .agent-item {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 10px;
          margin-bottom: 4px;
          border: 1px solid transparent;
          border-radius: 5px;
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            background: var(--el-fill-color-light);
          }

          &.active {
            background: var(--el-color-primary-light-9);
            border-color: var(--el-color-primary-light-7);
          }

          .agent-main {
            flex: 1;
            min-width: 0;

            .agent-name {
              font-size: 13px;
              font-weight: 500;
              overflow: hidden;
              text-overflow: ellipsis;
              white-space: nowrap;
            }

            .agent-meta {
              font-size: 12px;
              color: var(--el-text-color-secondary);
              margin-top: 2px;
            }
          }

          .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: var(--el-color-danger);
            flex-shrink: 0;

            &.online {
              background: var(--el-color-success);
            }
          }
        }
      }
    }

    .capability-panel {
      flex: 1;
      min-width: 0;
      display: flex;
      flex-direction: column;
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      overflow: hidden;

      .sub-tabs {
        :deep(.el-tabs__header) {
          margin-bottom: 0;
          padding: 0 12px;
        }

        .sub-tab-label {
          display: inline-flex;
          align-items: center;
          gap: 4px;
        }
      }

      .sub-tab-body {
        flex: 1;
        min-height: 0;
        overflow: hidden;
      }
    }
  }
</style>
