<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Delete, Edit, Plus, Refresh } from '@element-plus/icons-vue';
  import {
    type DiagnosticCommandOverride,
    type DiagnosticWebhookStatus,
    deleteDiagnosticCommandOverride,
    disableDiagnosticWebhook,
    enableDiagnosticWebhook,
    fetchDiagnosticCommandOverrides,
    fetchDiagnosticWebhookStatus,
    saveDiagnosticCommandOverride,
    saveDiagnosticWebhookConfig
  } from '@/service/api/diagnostic';
  import { fetchK8sClusterOptions } from '@/service/api/k8s';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import { RISK_LEVEL_META, formatDateTime } from '../shared';

  defineOptions({ name: 'DiagnosticGovernanceTab' });

  const diagStore = useDiagnosticStore();

  // ========== 命令风险覆盖 ==========
  const loading = ref(false);
  const overrides = ref<DiagnosticCommandOverride[]>([]);

  const dialogVisible = ref(false);
  const saving = ref(false);
  const form = reactive({ command: '', riskLevel: 'L5', description: '' });
  const editingId = ref<number | null>(null);

  async function load() {
    loading.value = true;
    try {
      const { data, error } = await fetchDiagnosticCommandOverrides();
      if (!error && data) {
        overrides.value = data;
      }
    } finally {
      loading.value = false;
    }
  }

  function openCreate() {
    editingId.value = null;
    form.command = '';
    form.riskLevel = 'L2';
    form.description = '';
    dialogVisible.value = true;
  }

  function openEdit(row: DiagnosticCommandOverride) {
    editingId.value = row.id;
    form.command = row.command;
    form.riskLevel = row.riskLevel;
    form.description = row.description;
    dialogVisible.value = true;
  }

  async function save() {
    if (!form.command.trim()) {
      ElMessage.warning('请输入命令名');
      return;
    }
    saving.value = true;
    try {
      const { data, error } = await saveDiagnosticCommandOverride({
        command: form.command.trim(),
        riskLevel: form.riskLevel,
        description: form.description.trim()
      });
      if (!error) {
        ElMessage.success('已保存，命令目录已即时生效');
        dialogVisible.value = false;
        load();
        // 刷新共享目录缓存
        if (data) {
          diagStore.commandCatalog = data;
        }
      }
    } finally {
      saving.value = false;
    }
  }

  async function remove(row: DiagnosticCommandOverride) {
    try {
      await ElMessageBox.confirm(`删除后「${row.command}」将恢复内置默认分级。确认删除？`, '删除覆盖', {
        type: 'warning'
      });
    } catch {
      return;
    }
    const { data, error } = await deleteDiagnosticCommandOverride(row.id);
    if (!error) {
      ElMessage.success('已删除');
      load();
      if (data) {
        diagStore.commandCatalog = data;
      }
    }
  }

  // ========== Webhook 自动注入管理（Pod label 触发） ==========
  interface ClusterOption {
    id: number;
    name: string;
    status: string;
  }

  const clusters = ref<ClusterOption[]>([]);
  const clusterId = ref<number | null>(null);
  const webhookStatus = ref<DiagnosticWebhookStatus | null>(null);
  const webhookLoading = ref(false);
  const webhookSaving = ref(false);
  const webhookForm = reactive({ webhookURL: '', tunnelURL: '', initImage: '' });

  async function loadClusters() {
    const { data, error } = await fetchK8sClusterOptions();
    if (!error && data && data.length > 0) {
      clusters.value = data;
      clusterId.value = data[0].id;
      loadWebhookStatus();
    }
  }

  async function loadWebhookStatus() {
    if (!clusterId.value) return;
    webhookLoading.value = true;
    try {
      const { data, error } = await fetchDiagnosticWebhookStatus(clusterId.value);
      if (!error && data) {
        webhookStatus.value = data;
        // 回填已保存配置，避免"保存后刷新变空、以为没保存"
        webhookForm.tunnelURL = data.tunnelURL ?? '';
        webhookForm.initImage = data.initImage ?? '';
        webhookForm.webhookURL = data.webhookURL ?? '';
      }
    } finally {
      webhookLoading.value = false;
    }
  }

  /** 由 tunnel HTTP 地址推导 WS 会话地址：http://x:8080 → ws://x:7777 */
  function deriveSessionURL(tunnelURL: string): string {
    if (!tunnelURL) return '';
    try {
      const u = new URL(tunnelURL);
      return `${u.protocol === 'https:' ? 'wss' : 'ws'}://${u.hostname}:7777`;
    } catch {
      return '';
    }
  }

  const switchWebhook = async (enabled: boolean) => {
    if (!clusterId.value || webhookLoading.value) return;
    const action = enabled ? enableDiagnosticWebhook : disableDiagnosticWebhook;
    // 启用/禁用为意图下发 + 等待集群内控制器收敛（最长约 20s），期间锁定防重复点击
    webhookLoading.value = true;
    try {
      const { error } = await action(clusterId.value);
      if (!error) {
        ElMessage.success(enabled ? '已启用自动注入' : '已禁用自动注入');
      }
    } finally {
      webhookLoading.value = false;
      loadWebhookStatus();
    }
  };

  const saveWebhookConfig = async () => {
    if (!clusterId.value) return;
    if (!webhookForm.tunnelURL.trim()) {
      ElMessage.warning('请填写 tunnel-server 地址（如 http://47.111.184.180:8080）');
      return;
    }
    webhookSaving.value = true;
    try {
      const { error } = await saveDiagnosticWebhookConfig({
        clusterId: clusterId.value,
        tunnelURL: webhookForm.tunnelURL.trim(),
        initImage: webhookForm.initImage.trim() || undefined,
        webhookURL: webhookForm.webhookURL.trim() || undefined
      });
      if (!error) {
        ElMessage.success('配置已保存（对后续新建 Pod 生效）');
        loadWebhookStatus();
      }
    } finally {
      webhookSaving.value = false;
    }
  };

  // ========== 接入指引 ==========
  const accessMethods = [
    {
      title: '方式二：Starter 主动接入',
      desc: '应用引入 arthas-spring-boot-starter，配置 tunnel 地址；适合自建镜像或无法使用 Webhook 的场景。',
      steps: [
        '依赖：com.taobao.arthas:arthas-spring-boot-starter',
        '配置：arthas.tunnel-server=ws://<tunnel-host>:7777/ws',
        '配置：arthas.agent-id=${POD_NAME}-${POD_NAMESPACE}'
      ]
    },
    {
      title: '方式三：基础镜像预置',
      desc: '在统一 Java 基础镜像中预置 -javaagent 启动参数，全量应用无差别接入。',
      steps: [
        '下载 arthas-boot.jar / arthas-agent 至基础镜像',
        'JAVA_TOOL_OPTIONS=-javaagent:/opt/arthas/arthas-agent.jar=tunnel_server=...',
        '重建并滚动更新所有应用镜像'
      ]
    }
  ];

  const injectionUsage = `为应用接入（管理员启用上方开关后，应用负责人自助操作）：
1. 在工作负载 spec.template.metadata.labels 添加 oneops-arthas-injection: enabled（切勿加进 spec.selector）
2. 触发滚动重启：kubectl rollout restart deploy/<app>（Webhook 仅对新建 Pod 生效）
3. 平台自动注入 initContainer + JAVA_TOOL_OPTIONS，agent 以 POD_NAME-POD_NAMESPACE 注册到 tunnel-server
4. 摘除接入：删除该 label 并再次滚动重启；单 Pod 显式排除可加 annotation oneops-arthas.injection/disabled: "true"`;

  const tunnelConfigTip = `tunnel-server 地址说明：
填入平台访问 tunnel-server 的 HTTP 地址（如 http://47.111.184.180:8080）。
平台会话地址自动推导为 ws://同主机:7777，agent 反连地址用集群内默认值。`;

  onMounted(() => {
    load();
    loadClusters();
  });
</script>

<template>
  <div class="governance-tab">
    <!-- Webhook 自动注入管理 -->
    <div class="webhook-mgmt">
      <div class="section-head">
        <div>
          <div class="section-title">Webhook 自动注入（推荐）</div>
          <div class="section-sub">
            按 Pod 粒度触发：仅注入带 label
            <code>oneops-arthas-injection=enabled</code>
            的 Pod，无需对命名空间打标签。
          </div>
        </div>
        <div class="section-actions">
          <ElSelect
            v-model="clusterId"
            placeholder="选择集群"
            size="small"
            style="width: 180px"
            @change="loadWebhookStatus"
          >
            <ElOption v-for="c in clusters" :key="c.id" :label="c.name" :value="c.id" />
          </ElSelect>
          <ElSwitch
            :model-value="webhookStatus?.enabled ?? false"
            :loading="webhookLoading"
            active-text="启用"
            inactive-text="禁用"
            inline-prompt
            @change="(val: any) => switchWebhook(Boolean(val))"
          />
        </div>
      </div>

      <div v-loading="webhookLoading" class="webhook-body">
        <!-- 当前生效配置（保存后此处回显） -->
        <ElDescriptions v-if="webhookStatus" :column="1" border size="small" class="webhook-desc">
          <ElDescriptionsItem label="注入状态">
            <ElTag :type="webhookStatus.enabled ? 'success' : 'info'" size="small">
              {{ webhookStatus.enabled ? '已启用' : '未启用' }}
            </ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="tunnel 探测地址">
            {{ webhookStatus.tunnelURL || '未配置' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="WS 会话地址">
            {{ webhookStatus.tunnelSess || deriveSessionURL(webhookStatus.tunnelURL) || '未配置' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="agent 反连地址">
            {{ webhookStatus.tunnelWS || '未配置' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="apiserver 回调地址">
            {{ webhookStatus.webhookURL || '未配置（启用前必须填写）' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="init 镜像">
            {{ webhookStatus.initImage || '未配置（用内置默认）' }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <ElAlert
          v-if="webhookStatus?.enabled"
          :title="`本集群已启用：${webhookStatus.objectRule}`"
          type="success"
          :closable="false"
          show-icon
          class="webhook-alert"
        />

        <ElForm label-width="130px" size="small" class="webhook-form">
          <ElFormItem label="回调地址(全局)">
            <ElInput
              v-model="webhookForm.webhookURL"
              placeholder="集群内部署填 svc://<ns>/msre-pilot:9443（如 svc://test/...）；平台直连填 https://<OneOps地址>:9443/webhook/msre-pilot"
              clearable
            />
          </ElFormItem>
          <ElFormItem label="tunnel 地址">
            <ElInput
              v-model="webhookForm.tunnelURL"
              placeholder="http://47.111.184.180:8080（平台访问 tunnel-server）"
              clearable
            />
            <div class="form-tip">会话地址自动推导为 ws://同主机:7777；agent 反连用集群内默认地址</div>
          </ElFormItem>
          <ElFormItem label="init 镜像">
            <ElInput
              v-model="webhookForm.initImage"
              placeholder="含 /opt/arthas/arthas-agent.jar 的镜像（默认 registry.cn-hangzhou.aliyuncs.com/oneops/arthas-agent）"
              clearable
            />
          </ElFormItem>
          <ElFormItem>
            <PermissionButton
              code="k8s.diagnostic.execute"
              type="primary"
              size="small"
              :loading="webhookSaving"
              @click="saveWebhookConfig"
            >
              保存配置
            </PermissionButton>
            <span class="form-tip">保存后点上方开关"启用"，证书与 TLS 服务自动就绪，无需配置任何环境变量</span>
          </ElFormItem>
        </ElForm>

        <ElAlert :title="injectionUsage" type="info" :closable="false" show-icon class="webhook-alert usage" />
      </div>
    </div>

    <!-- Pod 接入指引 -->
    <div class="access-guide">
      <div class="section-title">其他接入方式</div>
      <div class="guide-cards">
        <div v-for="m in accessMethods" :key="m.title" class="guide-card">
          <div class="guide-card-title">{{ m.title }}</div>
          <div class="guide-card-desc">{{ m.desc }}</div>
          <ol class="guide-steps">
            <li v-for="s in m.steps" :key="s">
              <code>{{ s }}</code>
            </li>
          </ol>
        </div>
      </div>
      <ElAlert :title="tunnelConfigTip" type="info" :closable="false" show-icon class="config-tip" />
    </div>

    <!-- 命令风险管控 -->
    <div class="command-governance">
      <div class="section-head">
        <div>
          <div class="section-title">命令风险管控</div>
          <div class="section-sub">
            覆盖内置命令的默认风险分级；设为「已禁用」即拉黑（工作台与终端入口均拒绝执行）。
          </div>
        </div>
        <div class="section-actions">
          <ElButton :icon="Refresh" size="small" :loading="loading" @click="load">刷新</ElButton>
          <PermissionButton code="k8s.diagnostic.execute" type="primary" size="small" :icon="Plus" @click="openCreate">
            新增覆盖
          </PermissionButton>
        </div>
      </div>

      <ElTable v-loading="loading" :data="overrides" size="small">
        <ElTableColumn prop="command" label="命令" width="140">
          <template #default="{ row }">
            <code class="cmd-text">{{ row.command }}</code>
          </template>
        </ElTableColumn>
        <ElTableColumn label="风险级" width="140">
          <template #default="{ row }">
            <ElTag size="small" :type="RISK_LEVEL_META[row.riskLevel]?.type ?? 'info'">
              {{ RISK_LEVEL_META[row.riskLevel]?.label ?? row.riskLevel }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="说明" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="updatedBy" label="更新人" width="110" />
        <ElTableColumn label="更新时间" width="165">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <PermissionButton
              code="k8s.diagnostic.execute"
              size="small"
              link
              type="primary"
              :icon="Edit"
              @click="openEdit(row)"
            >
              编辑
            </PermissionButton>
            <PermissionButton
              code="k8s.diagnostic.execute"
              size="small"
              link
              type="danger"
              :icon="Delete"
              @click="remove(row)"
            >
              删除
            </PermissionButton>
          </template>
        </ElTableColumn>
        <template #empty>
          <ElEmpty description="暂无覆盖记录，全部命令使用内置默认分级" :image-size="70" />
        </template>
      </ElTable>

      <!-- 风险分级说明 -->
      <div class="risk-legend">
        <div v-for="(meta, level) in RISK_LEVEL_META" :key="level" class="risk-item">
          <ElTag size="small" :type="meta.type">{{ meta.label }}</ElTag>
          <span class="risk-desc">{{ meta.desc }}</span>
        </div>
      </div>
    </div>

    <!-- 覆盖编辑 -->
    <ElDialog v-model="dialogVisible" :title="editingId ? '编辑命令覆盖' : '新增命令覆盖'" width="480px">
      <ElForm :model="form" label-width="90px">
        <ElFormItem label="命令名">
          <ElInput v-model="form.command" placeholder="如 ognl / vmtool / shutdown" :disabled="!!editingId" />
        </ElFormItem>
        <ElFormItem label="风险级">
          <ElSelect v-model="form.riskLevel" style="width: 100%">
            <ElOption v-for="(meta, level) in RISK_LEVEL_META" :key="level" :label="meta.label" :value="level" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="form.description" type="textarea" :rows="2" placeholder="覆盖原因/风险说明（可选）" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <PermissionButton code="k8s.diagnostic.execute" type="primary" :loading="saving" @click="save">
          保存
        </PermissionButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
  .governance-tab {
    display: flex;
    flex-direction: column;
    gap: 14px;

    .section-head {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;

      .section-title {
        font-size: 14px;
        font-weight: 600;
      }

      .section-sub {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        margin-top: 4px;

        code {
          background: var(--el-fill-color);
          padding: 0 4px;
          border-radius: 3px;
        }
      }

      .section-actions {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-shrink: 0;
      }
    }

    .webhook-mgmt {
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      padding: 14px;

      .webhook-body {
        .webhook-form {
          margin-top: 10px;
          max-width: 760px;

          .form-tip {
            margin-left: 12px;
            font-size: 12px;
            color: var(--el-text-color-secondary);
          }
        }

        .webhook-desc {
          margin-bottom: 10px;
        }

        .webhook-alert {
          margin-top: 10px;

          &.usage {
            white-space: pre-line;
          }
        }
      }
    }

    .access-guide {
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      padding: 14px;

      .section-title {
        font-size: 14px;
        font-weight: 600;
        margin-bottom: 10px;
      }

      .guide-cards {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 10px;
        margin-bottom: 10px;

        .guide-card {
          border: 1px solid var(--el-border-color-lighter);
          border-radius: 6px;
          padding: 12px;

          .guide-card-title {
            font-size: 13px;
            font-weight: 600;
            margin-bottom: 6px;
          }

          .guide-card-desc {
            font-size: 12px;
            color: var(--el-text-color-secondary);
            line-height: 1.6;
            margin-bottom: 8px;
          }

          .guide-steps {
            margin: 0;
            padding-left: 18px;
            font-size: 12px;

            li {
              margin-bottom: 4px;

              code {
                background: var(--el-fill-color);
                padding: 1px 5px;
                border-radius: 3px;
                word-break: break-all;
              }
            }
          }
        }
      }

      .config-tip {
        white-space: pre-line;
      }
    }

    .command-governance {
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      padding: 14px;

      .section-head {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 10px;

        .section-title {
          font-size: 14px;
          font-weight: 600;
        }

        .section-sub {
          font-size: 12px;
          color: var(--el-text-color-secondary);
          margin-top: 4px;
        }
      }

      .cmd-text {
        color: var(--el-color-primary);
        font-weight: 600;
      }

      .risk-legend {
        display: flex;
        flex-wrap: wrap;
        gap: 8px 20px;
        margin-top: 12px;

        .risk-item {
          display: inline-flex;
          align-items: center;
          gap: 6px;

          .risk-desc {
            font-size: 12px;
            color: var(--el-text-color-secondary);
          }
        }
      }
    }
  }
</style>
