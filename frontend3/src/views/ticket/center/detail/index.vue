<script setup lang="ts">
  import { computed, onMounted, ref, type Component } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useIntervalFn, useWindowSize } from '@vueuse/core';
  import {
    ArrowLeft,
    Bell,
    ChatDotRound,
    CircleCheckFilled,
    CircleCloseFilled,
    Clock,
    Flag,
    Minus,
    Position,
    Refresh,
    RefreshRight,
    Switch
  } from '@element-plus/icons-vue';
  import {
    approveTicket,
    cancelTicket,
    commentTicket,
    fetchTicketDetail,
    rejectTicket,
    urgeTicket
  } from '@/service/api';
  import { useAuthStore } from '@/store/modules/auth';
  import {
    URGE_COOLDOWN_MS,
    nodeStatusMap,
    nodeTagType,
    ticketPriorityMap,
    ticketStatusMap
  } from '../../constants';
  import TicketResubmitDialog from '../modules/ticket-resubmit-dialog.vue';
  import TicketReassignDialog from '../modules/ticket-reassign-dialog.vue';

  defineOptions({ name: 'TicketDetail' });

  const route = useRoute();
  const router = useRouter();
  const authStore = useAuthStore();

  const ticketId = computed(() => Number(route.query.id));
  const loading = ref(false);
  const detail = ref<Api.Ticket.TicketDetail | null>(null);
  const actionLoading = ref(false);
  /** 行内审批意见（审批操作条） */
  const actionComment = ref('');

  const resubmitVisible = ref(false);
  const reassignVisible = ref(false);

  const ticket = computed(() => detail.value?.ticket);
  const nodes = computed(() => detail.value?.nodes ?? []);
  const logs = computed(() => detail.value?.logs ?? []);

  /** 改派权限：有权限码且工单审批中 */
  const canReassign = computed(
    () => authStore.hasPermission('ticket.ticket.reassign') && ticket.value?.status === 'pending'
  );

  /** 发起人视角（审批进度条/催办入口）：审批中且当前用户是发起人（与撤销权限 canCancel 解耦） */
  const isCreatorPending = computed(
    () => ticket.value?.status === 'pending' && ticket.value.creatorId === authStore.userInfo.id
  );

  /** 催办冷却倒计时驱动：computed 无法感知时间流逝，用 30s tick 驱动重算 */
  const nowTick = ref(Date.now());
  useIntervalFn(() => {
    nowTick.value = Date.now();
  }, 30 * 1000);

  /** 头部基本信息响应式列数 */
  const { width: vw } = useWindowSize();
  const descColumn = computed(() => (vw.value < 768 ? 1 : vw.value < 1200 ? 2 : 4));

  /** 当前待审批节点（审批操作条/改派目标） */
  const pendingNode = computed(() => nodes.value.find(n => n.status === 'pending'));

  /** 最近一次驳回意见（重提弹窗提示用） */
  const lastRejectComment = computed(() => {
    const rejects = logs.value.filter(l => l.action === 'reject');
    return rejects.length ? rejects[rejects.length - 1].comment : '';
  });

  /** 催办冷却剩余提示（空串 = 可立即催办） */
  const urgeCooldownText = computed(() => {
    const at = ticket.value?.lastUrgeAt;
    if (!at) return '';
    const remain = URGE_COOLDOWN_MS - (nowTick.value - new Date(at).getTime());
    if (remain <= 0) return '';
    return `${Math.ceil(remain / 60000)} 分钟后可再次催办`;
  });

  /** 会签进度文本 */
  function signProgress(node?: Api.Ticket.TicketNodeRecord): string {
    if (!node || node.multiType !== 'all') return '';
    const total = node.approverIds ? node.approverIds.split(',').filter(Boolean) : [];
    const done = node.approvedIds ? node.approvedIds.split(',').filter(Boolean) : [];
    return `${done.length}/${total.length}`;
  }

  /** 节点记录状态 → 步骤条状态 */
  const stepStatusMap: Record<string, 'wait' | 'process' | 'finish' | 'error' | 'success'> = {
    waiting: 'wait',
    pending: 'process',
    approved: 'finish',
    rejected: 'error',
    skipped: 'wait',
    canceled: 'wait'
  };

  /** 已等待时长（当前节点 startedAt → now，由 nowTick 每 30s 驱动刷新） */
  function waitedText(startedAt?: string | null): string {
    if (!startedAt) return '';
    const min = Math.floor((nowTick.value - new Date(startedAt).getTime()) / 60000);
    if (min < 1) return '刚刚';
    if (min < 60) return `已等待 ${min} 分钟`;
    const hour = Math.floor(min / 60);
    if (hour < 24) return `已等待 ${hour} 小时 ${min % 60} 分`;
    return `已等待 ${Math.floor(hour / 24)} 天 ${hour % 24} 小时`;
  }

  interface FlowStep {
    key: string;
    title: string;
    status: 'wait' | 'process' | 'finish' | 'error' | 'success';
    desc: string;
  }

  /** 流程进度步骤条：发起 → 各审批节点 →（工单结束后）终点；细节历史仍由时间线承载 */
  const flowSteps = computed<FlowStep[]>(() => {
    const t = ticket.value;
    if (!t) return [];
    const steps: FlowStep[] = [{ key: 'submit', title: '发起', status: 'finish', desc: fmt(t.createdAt) }];
    nodes.value.forEach(n => {
      const segs: string[] = [];
      const names = approverList(n);
      if (names.length) segs.push(names.join('、'));
      const sp = signProgress(n);
      if (sp) segs.push(`会签 ${sp}`);
      if (n.status === 'pending') {
        const w = waitedText(n.startedAt);
        if (w) segs.push(w);
      }
      steps.push({
        key: `node-${n.id}`,
        title: n.nodeName,
        status: stepStatusMap[n.status] ?? 'wait',
        desc: segs.join(' · ')
      });
    });
    if (t.status !== 'pending') {
      const tone = ticketStatusMap[t.status];
      steps.push({
        key: 'finish',
        title: tone ? `工单${tone.label}` : '结束',
        status: t.status === 'approved' ? 'success' : t.status === 'rejected' ? 'error' : 'wait',
        desc: fmt(t.finishedAt)
      });
    }
    return steps;
  });

  /** 时间格式化：统一用 new Date 解析一次（正确处理后端时区偏移），输出浏览器本地时间。
   *  与时间线排序（new Date().getTime()）保持同一解析口径，避免显示与穿插顺序矛盾 */
  function fmt(t?: string | null): string {
    if (!t) return '';
    const d = new Date(t);
    if (Number.isNaN(d.getTime())) return t;
    const pad = (n: number) => `${n}`.padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  }

  /** 审批人姓名列表 */
  function approverList(node: Api.Ticket.TicketNodeRecord): string[] {
    return node.approverNames ? node.approverNames.split(',').map(s => s.trim()).filter(Boolean) : [];
  }

  /** 头像底色（按姓名散列取色板） */
  const avatarPalette = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#9a6fe0', '#23b7c4', '#5a7dc4'];
  function avatarStyle(name: string) {
    const code = [...name].reduce((s, c) => s + c.charCodeAt(0), 0);
    return { backgroundColor: avatarPalette[code % avatarPalette.length] };
  }

  /** 节点状态 → 时间线圆点图标/颜色 */
  function nodeDot(status: string): { icon: Component; color: string } {
    switch (status) {
      case 'approved':
        return { icon: CircleCheckFilled, color: '#67c23a' };
      case 'rejected':
        return { icon: CircleCloseFilled, color: '#f56c6c' };
      case 'pending':
        return { icon: Clock, color: '#409eff' };
      case 'skipped':
        return { icon: Minus, color: '#c0c4cc' };
      case 'canceled':
        return { icon: Minus, color: '#c0c4cc' };
      default:
        return { icon: Clock, color: '#c0c4cc' };
    }
  }

  /** 统一审批时间线条目：发起 → 节点/评论/事件（按时间穿插）→ 结束 */
  interface TimelineItem {
    key: string;
    kind: 'submit' | 'node' | 'comment' | 'event' | 'finish';
    icon: Component;
    color: string;
    timeText: string;
    node?: Api.Ticket.TicketNodeRecord;
    comment?: { name: string; content: string };
    event?: { name: string; content: string };
    finishLabel?: string;
  }

  /** 事件类日志（resubmit/reassign/urge）的时间线展示配置 */
  function logEventMeta(action: string): { icon: Component; color: string; text: string } | null {
    switch (action) {
      case 'resubmit':
        return { icon: RefreshRight, color: '#e6a23c', text: '驳回后修改并重新提交，流程重新流转' };
      case 'reassign':
        return { icon: Switch, color: '#9a6fe0', text: '改派审批人' };
      case 'urge':
        return { icon: Bell, color: '#e6a23c', text: '发送催办提醒' };
      default:
        return null;
    }
  }

  /** 事件条目文案：改派/催办带节点名前缀，重提用固定文案 */
  function logEventContent(log: Api.Ticket.TicketFlowLog, meta: { text: string }): string {
    const nodePrefix = log.nodeName ? `「${log.nodeName}」` : '';
    switch (log.action) {
      case 'reassign':
        return `${nodePrefix}${log.comment || meta.text}`;
      case 'urge':
        return `${nodePrefix}催办提醒`;
      default:
        return meta.text;
    }
  }

  /** 同刻排序权重：节点（按流程定义序，0 起）< 事件 < 评论 */
  const ORDER_EVENT = 9000;
  const ORDER_COMMENT = 10000;

  const timelineItems = computed<TimelineItem[]>(() => {
    const t = ticket.value;
    if (!t) return [];

    const items: TimelineItem[] = [
      { key: 'submit', kind: 'submit', icon: Position, color: '#409eff', timeText: fmt(t.createdAt) }
    ];

    // 节点/评论/事件按时间穿插（未开始的节点保持原顺序靠后）
    type Ev = { key: string; time: number; order: number; item: TimelineItem };
    const evs: Ev[] = [];
    nodes.value.forEach((n, i) => {
      const dot = nodeDot(n.status);
      evs.push({
        key: `node-${n.id}`,
        time: n.startedAt ? new Date(n.startedAt).getTime() : Number.MAX_SAFE_INTEGER,
        order: i,
        item: { key: `node-${n.id}`, kind: 'node', icon: dot.icon, color: dot.color, timeText: fmt(n.startedAt), node: n }
      });
    });
    logs.value
      .filter(l => ['comment', 'resubmit', 'reassign', 'urge'].includes(l.action))
      .forEach(l => {
        if (l.action === 'comment') {
          evs.push({
            key: `comment-${l.id}`,
            time: new Date(l.createdAt).getTime(),
            order: ORDER_COMMENT,
            item: {
              key: `comment-${l.id}`,
              kind: 'comment',
              icon: ChatDotRound,
              color: '#909399',
              timeText: fmt(l.createdAt),
              comment: { name: l.operatorName, content: l.comment || '' }
            }
          });
          return;
        }
        const meta = logEventMeta(l.action);
        if (!meta) return;
        evs.push({
          key: `event-${l.id}`,
          time: new Date(l.createdAt).getTime(),
          order: ORDER_EVENT,
          item: {
            key: `event-${l.id}`,
            kind: 'event',
            icon: meta.icon,
            color: meta.color,
            timeText: fmt(l.createdAt),
            event: { name: l.operatorName, content: logEventContent(l, meta) }
          }
        });
      });
    evs.sort((a, b) => a.time - b.time || a.order - b.order);
    items.push(...evs.map(e => e.item));

    if (t.status !== 'pending') {
      const tone = ticketStatusMap[t.status];
      const finishColor = t.status === 'approved' ? '#67c23a' : t.status === 'rejected' ? '#f56c6c' : '#909399';
      items.push({
        key: 'finish',
        kind: 'finish',
        icon: Flag,
        color: finishColor,
        timeText: fmt(t.finishedAt),
        finishLabel: tone ? `工单${tone.label}` : '工单已结束'
      });
    }
    return items;
  });

  async function loadDetail() {
    if (!ticketId.value) return;
    loading.value = true;
    const { data, error } = await fetchTicketDetail(ticketId.value);
    loading.value = false;
    if (!error && data) {
      detail.value = data;
    }
  }

  function handleBack() {
    router.push({ name: 'ticket_center' });
  }

  /** 回车快捷通过：仅在已填写意见时生效，空意见回车不触发（避免误触终审） */
  function handleEnterApprove() {
    if (actionComment.value.trim()) handleApprove();
  }

  /** 审批通过（操作条行内意见） */
  async function handleApprove() {
    actionLoading.value = true;
    const { error } = await approveTicket(ticketId.value, actionComment.value.trim() || '同意');
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已通过');
      actionComment.value = '';
      loadDetail();
    }
  }

  /** 驳回（操作条行内意见，必填；驳回会终止流程，需二次确认） */
  async function handleReject() {
    const comment = actionComment.value.trim();
    if (!comment) {
      ElMessage.warning('请填写驳回原因');
      return;
    }
    try {
      await ElMessageBox.confirm('驳回后流程立即终止，后续节点不再审批。确认驳回该工单？', '驳回工单', {
        confirmButtonText: '确认驳回',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
    actionLoading.value = true;
    const { error } = await rejectTicket(ticketId.value, comment);
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已驳回');
      actionComment.value = '';
      loadDetail();
    }
  }

  /** 发起人催办当前节点审批人 */
  async function handleUrge() {
    actionLoading.value = true;
    const { error } = await urgeTicket(ticketId.value);
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已向当前审批人发送催办提醒');
      loadDetail();
    }
  }

  /** 撤销 */
  async function handleCancel() {
    try {
      await ElMessageBox.confirm('确定撤销该工单吗？撤销后流程立即终止。', '撤销工单', {
        confirmButtonText: '撤销',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
    actionLoading.value = true;
    const { error } = await cancelTicket(ticketId.value, '');
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已撤销');
      loadDetail();
    }
  }

  /** 评论 */
  async function handleComment() {
    const { value } = await ElMessageBox.prompt('评论内容', '添加评论', {
      confirmButtonText: '提交',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputValidator: (v: string) => Boolean(v && v.trim()) || '请填写评论内容'
    }).catch(() => ({ value: null }) as { value: string | null });

    if (value === null) return;
    const { error } = await commentTicket(ticketId.value, value.trim());
    if (!error) {
      ElMessage.success('评论成功');
      loadDetail();
    }
  }

  /** 重提成功：关弹窗并刷新详情 */
  function handleResubmitSuccess() {
    resubmitVisible.value = false;
    loadDetail();
  }

  /** 改派成功：关弹窗并刷新详情 */
  function handleReassignSuccess() {
    reassignVisible.value = false;
    loadDetail();
  }

  onMounted(loadDetail);
</script>

<template>
  <div v-loading="loading" class="min-h-500px flex-col-stretch gap-16px">
    <!-- 头部：基本信息 + 操作 -->
    <ElCard class="card-wrapper">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-12px">
            <ElButton :icon="ArrowLeft" text @click="handleBack">返回</ElButton>
            <span class="text-lg font-medium">{{ ticket?.title ?? '工单详情' }}</span>
            <ElTag v-if="ticket" :type="ticketStatusMap[ticket.status]?.type ?? 'info'">
              {{ ticketStatusMap[ticket.status]?.label ?? ticket.status }}
            </ElTag>
          </div>
          <div class="flex gap-8px">
            <ElButton v-if="detail?.canCancel" type="warning" :loading="actionLoading" @click="handleCancel">
              撤销
            </ElButton>
            <ElButton
              v-if="detail?.canResubmit"
              type="warning"
              plain
              :icon="RefreshRight"
              @click="resubmitVisible = true"
            >
              重新提交
            </ElButton>
            <ElButton v-if="canReassign && pendingNode" :icon="Switch" @click="reassignVisible = true">改派</ElButton>
            <ElButton @click="handleComment">评论</ElButton>
            <ElButton :icon="Refresh" :loading="loading" title="刷新" @click="loadDetail" />
          </div>
        </div>
      </template>

      <ElDescriptions :column="descColumn" border size="small">
        <ElDescriptionsItem label="工单号">{{ ticket?.ticketNo ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="类型">{{ ticket?.typeName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="优先级">
          <ElTag v-if="ticket" :type="ticketPriorityMap[ticket.priority]?.type ?? 'info'" size="small">
            {{ ticketPriorityMap[ticket.priority]?.label ?? ticket.priority }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="流程">{{ ticket?.workflowName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="发起人">{{ ticket?.creatorName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="创建时间">{{ fmt(ticket?.createdAt) || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="当前节点">
          {{ ticket?.status === 'pending' ? ticket?.currentNodeName || '-' : '已结束' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="结束时间">{{ fmt(ticket?.finishedAt) || '-' }}</ElDescriptionsItem>
      </ElDescriptions>
    </ElCard>

    <!-- 流程进度概览：一眼看清走到哪一步（完整历史见下方时间线），窄屏转纵向 -->
    <ElCard v-if="flowSteps.length" class="card-wrapper" title="流程进度">
      <ElSteps :active="flowSteps.length" :direction="vw < 768 ? 'vertical' : 'horizontal'" align-center>
        <ElStep v-for="s in flowSteps" :key="s.key" :title="s.title" :status="s.status" :description="s.desc" />
      </ElSteps>
    </ElCard>

    <!-- 审批操作条：待我审批（行内意见 + 通过/驳回），吸顶保证长表单下始终可达 -->
    <ElCard v-if="detail?.canApprove" class="card-wrapper sticky top-0 z-10" shadow="never">
      <div class="flex flex-wrap items-center justify-between gap-12px">
        <div class="flex flex-wrap items-center gap-8px">
          <ElIcon :size="18" color="#e6a23c"><Bell /></ElIcon>
          <span class="font-medium">节点「{{ pendingNode?.nodeName }}」等待你的审批</span>
          <span v-if="signProgress(pendingNode)" class="text-13px text-primary">
            会签 {{ signProgress(pendingNode) }}
          </span>
        </div>
        <div class="flex min-w-360px flex-1 items-center justify-end gap-8px">
          <ElInput
            v-model="actionComment"
            placeholder="审批意见（驳回时必填，填写后可回车通过）"
            clearable
            maxlength="500"
            class="max-w-360px"
            @keyup.enter="handleEnterApprove"
          />
          <ElButton type="success" :loading="actionLoading" @click="handleApprove">通过</ElButton>
          <ElButton type="danger" :loading="actionLoading" @click="handleReject">驳回</ElButton>
        </div>
      </div>
    </ElCard>

    <!-- 发起人视角条：审批进度 + 催办，吸顶 -->
    <ElCard v-else-if="isCreatorPending" class="card-wrapper sticky top-0 z-10" shadow="never">
      <div class="flex flex-wrap items-center justify-between gap-12px">
        <div class="flex items-center gap-8px text-13px text-gray-600">
          <ElIcon :size="16" color="#409eff"><Clock /></ElIcon>
          <span>
            审批中：当前节点「{{ ticket?.currentNodeName }}」，审批人 {{ pendingNode?.approverNames || '-' }}
          </span>
        </div>
        <div class="flex items-center gap-8px">
          <span v-if="urgeCooldownText" class="text-12px text-gray-400">{{ urgeCooldownText }}</span>
          <ElButton
            type="warning"
            plain
            :icon="Bell"
            :disabled="!detail?.canUrge"
            :loading="actionLoading"
            @click="handleUrge"
          >
            催办
          </ElButton>
        </div>
      </div>
    </ElCard>

    <div class="grid grid-cols-[1fr_480px] gap-16px lt-xl:grid-cols-1">
      <!-- 申请内容 -->
      <ElCard class="card-wrapper" title="申请内容">
        <ElDescriptions v-if="detail" :column="1" border size="small">
          <ElDescriptionsItem
            v-for="field in detail.formSchema"
            :key="field.key"
            :label="field.label"
            :label-width="140"
          >
            <pre v-if="field.type === 'textarea'" class="m-0 whitespace-pre-wrap break-all font-mono text-13px">{{
              (detail.formData[field.key] as string) || '-'
            }}</pre>
            <template v-else>{{ detail.formData[field.key] ?? '-' }}</template>
          </ElDescriptionsItem>
        </ElDescriptions>
        <ElEmpty v-else-if="!loading" description="暂无表单数据" />
      </ElCard>

      <!-- 审批记录（发起 → 节点/评论/事件 → 结束，统一时间线） -->
      <ElCard class="card-wrapper" title="审批记录">
        <ElTimeline v-if="timelineItems.length" class="pl-2px">
          <ElTimelineItem
            v-for="item in timelineItems"
            :key="item.key"
            :timestamp="item.timeText"
            :hide-timestamp="!item.timeText"
            placement="top"
          >
            <template #dot>
              <ElIcon :size="16" :color="item.color" class="bg-white">
                <component :is="item.icon" />
              </ElIcon>
            </template>

            <!-- 发起 -->
            <div v-if="item.kind === 'submit'" class="flex items-center gap-8px pb-4px">
              <span class="font-medium">发起申请</span>
              <span class="text-13px text-gray-500">{{ ticket?.creatorName }}</span>
            </div>

            <!-- 审批节点 -->
            <div v-else-if="item.kind === 'node' && item.node" class="pb-4px">
              <div class="flex flex-wrap items-center gap-8px">
                <span class="font-medium">{{ item.node.nodeName }}</span>
                <ElTag size="small" :type="nodeTagType[item.node.status] ?? 'info'">
                  {{ nodeStatusMap[item.node.status]?.label ?? item.node.status }}
                </ElTag>
                <span v-if="signProgress(item.node)" class="text-12px text-primary">
                  会签 {{ signProgress(item.node) }}
                </span>
              </div>
              <div class="mt-6px flex flex-wrap items-center gap-10px">
                <div v-for="name in approverList(item.node)" :key="name" class="flex items-center gap-4px">
                  <span
                    class="flex h-22px w-22px items-center justify-center rounded-full text-12px text-white"
                    :style="avatarStyle(name)"
                  >
                    {{ name.slice(0, 1) }}
                  </span>
                  <span class="text-13px text-gray-600">{{ name }}</span>
                </div>
                <span v-if="!approverList(item.node).length" class="text-13px text-gray-400">-</span>
              </div>
              <div v-if="item.node.finishedAt" class="mt-4px text-12px text-gray-400">
                {{ item.node.status === 'canceled' ? '终止于' : '审批于' }} {{ fmt(item.node.finishedAt) }}
              </div>
              <div
                v-if="item.node.comment"
                class="mt-6px rounded-6px bg-gray-100 px-10px py-6px text-13px text-gray-700"
              >
                {{ item.node.comment }}
              </div>
            </div>

            <!-- 评论 -->
            <div v-else-if="item.kind === 'comment' && item.comment" class="pb-4px">
              <div class="flex items-center gap-8px">
                <span class="text-13px text-gray-500">{{ item.comment.name }}</span>
                <span class="text-12px text-gray-400">添加了评论</span>
              </div>
              <div class="mt-4px rounded-6px bg-gray-100 px-10px py-6px text-13px text-gray-700">
                {{ item.comment.content }}
              </div>
            </div>

            <!-- 事件（重新提交/改派/催办） -->
            <div v-else-if="item.kind === 'event' && item.event" class="flex flex-wrap items-center gap-8px pb-4px">
              <span class="text-13px text-gray-500">{{ item.event.name }}</span>
              <span class="text-13px font-medium" :style="{ color: item.color }">{{ item.event.content }}</span>
            </div>

            <!-- 结束 -->
            <div v-else-if="item.kind === 'finish'" class="flex items-center gap-8px">
              <span class="font-medium">{{ item.finishLabel }}</span>
            </div>
          </ElTimelineItem>
        </ElTimeline>
        <ElEmpty v-else description="暂无审批记录" />
      </ElCard>
    </div>

    <!-- 重新提交弹窗（驳回后发起人） -->
    <TicketResubmitDialog
      v-if="ticket"
      v-model:visible="resubmitVisible"
      :ticket-id="ticketId"
      :title="ticket.title"
      :priority="ticket.priority"
      :form-schema="detail?.formSchema ?? []"
      :form-data="detail?.formData ?? {}"
      :reject-comment="lastRejectComment"
      @success="handleResubmitSuccess"
    />

    <!-- 改派弹窗（管理员） -->
    <TicketReassignDialog
      v-model:visible="reassignVisible"
      :ticket-id="ticketId"
      :node-name="pendingNode?.nodeName ?? ''"
      :current-approvers="pendingNode?.approverNames ?? ''"
      @success="handleReassignSuccess"
    />
  </div>
</template>
