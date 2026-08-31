<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue';
  import { ElMessage } from 'element-plus';
  import { useAuthStore } from '@/store/modules/auth';
  import { fetchNotificationChannels, fetchNotifyLogs, fetchNotifyPolicies, updateNotifyPolicy } from '@/service/api';

  defineOptions({
    name: 'TicketNotifySettings'
  });

  const authStore = useAuthStore();
  const hasUpdatePerm = computed(() => authStore.hasPermission('ticket.notify.update'));
  const activeTab = ref('matrix');

  // ============================================
  // Tab 1：事件矩阵（事件 × 渠道绑定）
  // ============================================

  const policies = ref<Api.Ticket.NotifyPolicy[]>([]);
  const loading = ref(false);
  // 渠道选项（系统管理-通知渠道维护的平台级渠道池）
  const channelOptions = ref<Monitoring.NotificationChannel[]>([]);
  const savingEvent = ref('');

  async function loadPolicies() {
    loading.value = true;
    try {
      const { data } = await fetchNotifyPolicies();
      policies.value = data || [];
    } catch {
      ElMessage.error('加载通知设置失败');
      policies.value = [];
    } finally {
      loading.value = false;
    }
  }

  async function loadChannels() {
    try {
      const { data } = await fetchNotificationChannels();
      channelOptions.value = data || [];
    } catch {
      channelOptions.value = [];
    }
  }

  // 渠道下拉选项文案：类型 + 名称（未启用渠道可被选中但发送时跳过，给出提示）
  function channelLabel(channel: Monitoring.NotificationChannel) {
    const typeText: Record<string, string> = {
      email: '邮件',
      wechat: '企业微信',
      dingtalk: '钉钉',
      feishu: '飞书'
    };
    const enabledText = channel.enabled ? '' : '（已停用）';
    return `${typeText[channel.channelType] || channel.channelType} · ${channel.channelName}${enabledText}`;
  }

  // 保存单个事件（渠道/启用变更；模板经独立弹窗保存，此处保留原值不覆盖）
  async function handleSave(row: Api.Ticket.NotifyPolicy) {
    savingEvent.value = row.event;
    try {
      const { error } = await updateNotifyPolicy(row.event, {
        channels: row.channels || [],
        enabled: row.enabled,
        titleTpl: row.titleTpl || '',
        bodyTpl: row.bodyTpl || ''
      });
      if (!error) {
        ElMessage.success(`「${row.name}」已保存`);
        await loadPolicies();
      } else {
        ElMessage.error('保存失败');
      }
    } catch {
      ElMessage.error('保存失败');
    } finally {
      savingEvent.value = '';
    }
  }

  // ============================================
  // 通知模板（蓝图⑤：按事件自定义标题/正文，空=内置默认）
  // ============================================

  const tplDialog = reactive({
    visible: false,
    event: '',
    name: '',
    titleTpl: '',
    bodyTpl: '',
    // 保存时回传当前行的渠道绑定与启用状态（API 为整行 upsert）
    channels: [] as number[],
    enabled: 1
  });
  const savingTpl = ref(false);

  // 模板变量说明（通用 + 事件特有）
  const tplVars: Array<{ name: string; desc: string; events?: string }> = [
    { name: '{{ticketNo}}', desc: '工单号' },
    { name: '{{title}}', desc: '工单标题' },
    { name: '{{typeName}}', desc: '工单场景名称' },
    { name: '{{creator}}', desc: '发起人昵称' },
    { name: '{{priority}}', desc: '优先级（低/普通/高/紧急）' },
    { name: '{{time}}', desc: '通知生成时间' },
    { name: '{{node}}', desc: '节点名称', events: '待审批/改派/撤销/催办/超时/升级' },
    { name: '{{operator}}', desc: '操作人（催办人/改派人等）', events: '审批结果/改派/撤销/催办' },
    { name: '{{status}}', desc: '审批结果（已审批通过/已被驳回）', events: '审批结果' },
    { name: '{{comment}}', desc: '审批意见', events: '审批结果' },
    { name: '{{stay}}', desc: '节点已停留时长', events: '超时/升级' },
    { name: '{{threshold}}', desc: '超时阈值（小时）', events: '超时/升级' },
    { name: '{{approvers}}', desc: '待审批人名单', events: '升级' }
  ];

  function handleEditTpl(row: Api.Ticket.NotifyPolicy) {
    tplDialog.event = row.event;
    tplDialog.name = row.name;
    tplDialog.titleTpl = row.titleTpl || '';
    tplDialog.bodyTpl = row.bodyTpl || '';
    tplDialog.channels = row.channels ? [...row.channels] : [];
    tplDialog.enabled = row.enabled;
    tplDialog.visible = true;
  }

  async function handleSaveTpl() {
    savingTpl.value = true;
    try {
      const { error } = await updateNotifyPolicy(tplDialog.event, {
        channels: tplDialog.channels,
        enabled: tplDialog.enabled,
        titleTpl: tplDialog.titleTpl.trim(),
        bodyTpl: tplDialog.bodyTpl.trim()
      });
      if (!error) {
        ElMessage.success('模板已保存（5 分钟内生效）');
        tplDialog.visible = false;
        await loadPolicies();
      } else {
        ElMessage.error('保存失败');
      }
    } catch {
      ElMessage.error('保存失败');
    } finally {
      savingTpl.value = false;
    }
  }

  function handleResetTpl() {
    tplDialog.titleTpl = '';
    tplDialog.bodyTpl = '';
  }

  // 行内变更（选择渠道/开关）标记为"已定制"，提示需保存
  function markDirty(row: Api.Ticket.NotifyPolicy) {
    row.configured = true;
  }

  // 事件英文码 → 中文名（发送记录表格复用）
  const eventTextMap: Record<string, string> = {
    pending: '待审批',
    result: '审批结果',
    reassign: '改派',
    cancel: '撤销',
    urge: '催办',
    timeout: '审批超时',
    escalation: '超时升级'
  };

  function eventText(event: string) {
    return eventTextMap[event] || event;
  }

  // ============================================
  // Tab 2：发送记录（外部渠道投递结果，排障用）
  // ============================================

  const logs = ref<Api.Ticket.NotifyLog[]>([]);
  const logsLoading = ref(false);
  const logQuery = reactive({
    page: 1,
    pageSize: 20,
    event: '',
    status: -1 // -1=全部 0失败 1成功
  });
  const logTotal = ref(0);

  async function loadLogs() {
    logsLoading.value = true;
    try {
      const { data, error } = await fetchNotifyLogs(logQuery);
      if (!error) {
        // 后端返回统一分页结构 { list, total, ... }
        const wrapped = (data as unknown as { list?: Api.Ticket.NotifyLog[]; total?: number } | null) ?? {};
        logs.value = Array.isArray(wrapped.list) ? wrapped.list : [];
        logTotal.value = wrapped.total || 0;
      }
    } catch {
      ElMessage.error('加载发送记录失败');
    } finally {
      logsLoading.value = false;
    }
  }

  function handleLogSearch() {
    logQuery.page = 1;
    loadLogs();
  }

  onMounted(() => {
    loadPolicies();
    loadChannels();
  });
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 标题栏 -->
    <ElCard shadow="never">
      <div class="flex items-center justify-between">
        <div>
          <span class="text-lg font-semibold">通知设置</span>
          <div class="mt-1 text-xs text-gray-400">
            事件矩阵：为各审批事件选择通知渠道（渠道在「系统管理 → 通知渠道」维护）。未定制的事件默认走全部启用渠道；站内消息为兜底渠道，事件启用即写入
          </div>
        </div>
      </div>
    </ElCard>

    <ElCard shadow="never">
      <ElTabs v-model="activeTab" @tab-change="(name: string | number) => name === 'logs' && loadLogs()">
        <!-- 事件矩阵 -->
        <ElTabPane label="事件矩阵" name="matrix">
          <ElTable v-loading="loading" :data="policies" border stripe>
            <ElTableColumn label="通知事件" width="130">
              <template #default="{ row }">
                <span class="font-medium">{{ row.name }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="说明" min-width="180">
              <template #default="{ row }">
                <span class="text-sm text-gray-500">{{ row.desc }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="通知渠道" min-width="300">
              <template #default="{ row }">
                <div class="flex items-center gap-2">
                  <ElSelect
                    v-model="row.channels"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    clearable
                    placeholder="不选则该事件不发送"
                    :disabled="!hasUpdatePerm || row.enabled === 0"
                    class="w-full"
                    @change="markDirty(row)"
                  >
                    <ElOption
                      v-for="ch in channelOptions"
                      :key="ch.id"
                      :value="ch.id"
                      :label="channelLabel(ch)"
                      :disabled="!ch.enabled"
                    />
                  </ElSelect>
                  <ElTag v-if="!row.configured" type="info" size="small" class="shrink-0">默认</ElTag>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn label="通知模板" width="110" align="center">
              <template #default="{ row }">
                <div class="flex items-center justify-center gap-1">
                  <ElTag v-if="row.hasTpl" type="warning" size="small">自定义</ElTag>
                  <ElTag v-else type="info" size="small">默认</ElTag>
                  <PermissionButton code="ticket.notify.update" type="primary" link size="small" @click="handleEditTpl(row)">
                    编辑
                  </PermissionButton>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn label="启用" width="90" align="center">
              <template #default="{ row }">
                <ElSwitch
                  v-model="row.enabled"
                  :active-value="1"
                  :inactive-value="0"
                  :disabled="!hasUpdatePerm"
                  @change="markDirty(row)"
                />
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="100" align="center" fixed="right">
              <template #default="{ row }">
                <PermissionButton
                  code="ticket.notify.update"
                  type="primary"
                  link
                  size="small"
                  :loading="savingEvent === row.event"
                  @click="handleSave(row)"
                >
                  保存
                </PermissionButton>
              </template>
            </ElTableColumn>
          </ElTable>

          <ElEmpty v-if="!loading && policies.length === 0" description="暂无通知事件" />
        </ElTabPane>

        <!-- 发送记录 -->
        <ElTabPane label="发送记录" name="logs">
          <div class="mb-4 flex items-center gap-3">
            <ElSelect v-model="logQuery.event" clearable placeholder="事件" class="w-40" @change="handleLogSearch">
              <ElOption label="待审批" value="pending" />
              <ElOption label="审批结果" value="result" />
              <ElOption label="改派" value="reassign" />
              <ElOption label="撤销" value="cancel" />
              <ElOption label="催办" value="urge" />
              <ElOption label="审批超时" value="timeout" />
              <ElOption label="超时升级" value="escalation" />
            </ElSelect>
            <ElSelect v-model="logQuery.status" clearable placeholder="状态" class="w-32" @change="handleLogSearch">
              <ElOption label="成功" :value="1" />
              <ElOption label="失败" :value="0" />
            </ElSelect>
            <PermissionButton code="ticket.notify.list" type="primary" @click="loadLogs">查询</PermissionButton>
            <span class="text-xs text-gray-400">外部渠道投递结果（保留 30 天）；站内消息不在此列，见头部通知中心</span>
          </div>

          <ElTable v-loading="logsLoading" :data="logs" border stripe>
            <ElTableColumn label="时间" width="170">
              <template #default="{ row }">
                {{ new Date(row.createdAt).toLocaleString('zh-CN', { hour12: false }) }}
              </template>
            </ElTableColumn>
            <ElTableColumn label="事件" width="100" align="center">
              <template #default="{ row }">
                <ElTag size="small">{{ eventText(row.event) }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="工单" width="140">
              <template #default="{ row }">
                <span class="text-xs">{{ row.ticketNo || '-' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="渠道" width="150">
              <template #default="{ row }">
                {{ row.channelName || row.channelType }}
              </template>
            </ElTableColumn>
            <ElTableColumn label="接收者" min-width="220" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="text-xs">{{ row.recipient }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="80" align="center">
              <template #default="{ row }">
                <ElTag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                  {{ row.status === 1 ? '成功' : '失败' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="错误信息" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="text-xs" :class="row.error ? 'text-red-500' : 'text-gray-400'">{{ row.error || '—' }}</span>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="mt-4 flex justify-end">
            <ElPagination
              v-model:current-page="logQuery.page"
              v-model:page-size="logQuery.pageSize"
              :total="logTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              @current-change="loadLogs"
              @size-change="handleLogSearch"
            />
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 通知模板编辑弹窗（蓝图⑤） -->
    <ElDialog v-model="tplDialog.visible" :title="`通知模板 · ${tplDialog.name}`" width="640px" destroy-on-close>
      <div class="space-y-4">
        <ElAlert type="info" :closable="false" show-icon>
          <template #title>
            留空使用系统内置默认文案；未定义变量将原样保留。保存后约 5 分钟内生效（渠道缓存刷新）
          </template>
        </ElAlert>

        <div>
          <div class="mb-1 text-sm font-medium">标题模板</div>
          <ElInput v-model="tplDialog.titleTpl" placeholder="例：【OneOps 催办】{{title}}（{{ticketNo}}）" maxlength="500" show-word-limit />
        </div>

        <div>
          <div class="mb-1 text-sm font-medium">正文模板</div>
          <ElInput
            v-model="tplDialog.bodyTpl"
            type="textarea"
            :rows="7"
            placeholder="例：{{operator}} 提醒您：工单「{{title}}」的节点「{{node}}」等待审批，请尽快处理。"
            maxlength="5000"
            show-word-limit
          />
        </div>

        <div class="rounded bg-gray-50 p-3 dark:bg-gray-800">
          <div class="mb-2 text-xs font-medium text-gray-500">可用变量（点击插入正文）</div>
          <div class="flex flex-wrap gap-2">
            <ElTag
              v-for="v in tplVars"
              :key="v.name"
              :title="v.events ? `${v.desc}（适用：${v.events}）` : v.desc"
              class="cursor-pointer"
              @click="tplDialog.bodyTpl += tplDialog.bodyTpl ? `\n${v.name}` : v.name"
            >
              {{ v.name }}
            </ElTag>
          </div>
          <div class="mt-2 text-xs text-gray-400">通用变量所有事件可用；其余仅标注事件可用，其他事件中会原样保留</div>
        </div>
      </div>

      <template #footer>
        <ElButton @click="tplDialog.visible = false">取消</ElButton>
        <ElButton @click="handleResetTpl">恢复默认</ElButton>
        <ElButton type="primary" :loading="savingTpl" @click="handleSaveTpl">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
