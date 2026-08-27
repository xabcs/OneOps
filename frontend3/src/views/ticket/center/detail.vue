<script setup lang="ts">
  import { computed, onMounted, ref, type Component } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    ArrowLeft,
    ChatDotRound,
    CircleCheckFilled,
    CircleCloseFilled,
    Clock,
    Flag,
    Minus,
    Position
  } from '@element-plus/icons-vue';
  import {
    approveTicket,
    cancelTicket,
    commentTicket,
    fetchTicketDetail,
    rejectTicket
  } from '@/service/api';

  defineOptions({ name: 'TicketDetail' });

  const route = useRoute();
  const router = useRouter();

  const ticketId = computed(() => Number(route.query.id));
  const loading = ref(false);
  const detail = ref<Api.Ticket.TicketDetail | null>(null);
  const actionLoading = ref(false);

  const statusMap: Record<string, { label: string; type: 'primary' | 'success' | 'danger' | 'info' }> = {
    pending: { label: '审批中', type: 'primary' },
    approved: { label: '已通过', type: 'success' },
    rejected: { label: '已驳回', type: 'danger' },
    canceled: { label: '已撤销', type: 'info' }
  };

  const priorityMap: Record<string, { label: string; type: 'info' | 'primary' | 'warning' | 'danger' }> = {
    low: { label: '低', type: 'info' },
    normal: { label: '中', type: 'primary' },
    high: { label: '高', type: 'warning' },
    urgent: { label: '紧急', type: 'danger' }
  };

  const nodeStatusMap: Record<string, { label: string; type: 'wait' | 'process' | 'finish' | 'error' }> = {
    waiting: { label: '未到达', type: 'wait' },
    pending: { label: '审批中', type: 'process' },
    approved: { label: '已通过', type: 'finish' },
    rejected: { label: '已驳回', type: 'error' },
    skipped: { label: '已跳过', type: 'wait' }
  };

  const ticket = computed(() => detail.value?.ticket);
  const nodes = computed(() => detail.value?.nodes ?? []);
  const logs = computed(() => detail.value?.logs ?? []);

  /** 会签进度文本 */
  function signProgress(node: Api.Ticket.TicketNodeRecord): string {
    if (node.multiType !== 'all') return '';
    const total = node.approverIds ? node.approverIds.split(',').filter(Boolean) : [];
    const done = node.approvedIds ? node.approvedIds.split(',').filter(Boolean) : [];
    return `${done.length}/${total.length}`;
  }

  /** 时间格式化：去掉时区尾巴 */
  function fmt(t?: string | null): string {
    if (!t) return '';
    return t.length > 19 ? `${t.slice(0, 10)} ${t.slice(11, 19)}` : t;
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

  /** 节点状态 → 标签颜色 */
  const nodeTagType: Record<string, 'info' | 'primary' | 'success' | 'danger'> = {
    waiting: 'info',
    pending: 'primary',
    approved: 'success',
    rejected: 'danger',
    skipped: 'info'
  };

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
      default:
        return { icon: Clock, color: '#c0c4cc' };
    }
  }

  /** 统一审批时间线条目：发起 → 节点/评论（按时间穿插）→ 结束 */
  interface TimelineItem {
    key: string;
    kind: 'submit' | 'node' | 'comment' | 'finish';
    icon: Component;
    color: string;
    timeText: string;
    node?: Api.Ticket.TicketNodeRecord;
    comment?: { name: string; content: string };
    finishLabel?: string;
  }

  const timelineItems = computed<TimelineItem[]>(() => {
    const t = ticket.value;
    if (!t) return [];

    const items: TimelineItem[] = [
      { key: 'submit', kind: 'submit', icon: Position, color: '#409eff', timeText: fmt(t.createdAt) }
    ];

    // 节点与评论按时间穿插（未开始的节点保持原顺序靠后）
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
      .filter(l => l.action === 'comment')
      .forEach(l => {
        evs.push({
          key: `comment-${l.id}`,
          time: new Date(l.createdAt).getTime(),
          order: 10000,
          item: {
            key: `comment-${l.id}`,
            kind: 'comment',
            icon: ChatDotRound,
            color: '#909399',
            timeText: fmt(l.createdAt),
            comment: { name: l.operatorName, content: l.comment || '' }
          }
        });
      });
    evs.sort((a, b) => a.time - b.time || a.order - b.order);
    items.push(...evs.map(e => e.item));

    if (t.status !== 'pending') {
      const tone = statusMap[t.status];
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

  /** 审批通过 */
  async function handleApprove() {
    const { value } = await ElMessageBox.prompt('审批意见（可留空）', '审批通过', {
      confirmButtonText: '通过',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputValue: '同意'
    }).catch(() => ({ value: null }) as { value: string | null });

    if (value === null) return;
    actionLoading.value = true;
    const { error } = await approveTicket(ticketId.value, value ?? '');
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已通过');
      loadDetail();
    }
  }

  /** 驳回（意见必填） */
  async function handleReject() {
    const { value } = await ElMessageBox.prompt('驳回原因（必填）', '驳回工单', {
      confirmButtonText: '驳回',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputValidator: (v: string) => Boolean(v && v.trim()) || '请填写驳回原因'
    }).catch(() => ({ value: null }) as { value: string | null });

    if (value === null) return;
    actionLoading.value = true;
    const { error } = await rejectTicket(ticketId.value, value.trim());
    actionLoading.value = false;
    if (!error) {
      ElMessage.success('已驳回');
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
            <ElTag v-if="ticket" :type="statusMap[ticket.status]?.type ?? 'info'">
              {{ statusMap[ticket.status]?.label ?? ticket.status }}
            </ElTag>
          </div>
          <div class="flex gap-8px">
            <ElButton v-if="detail?.canApprove" type="success" :loading="actionLoading" @click="handleApprove">
              通过
            </ElButton>
            <ElButton v-if="detail?.canApprove" type="danger" :loading="actionLoading" @click="handleReject">
              驳回
            </ElButton>
            <ElButton v-if="detail?.canCancel" type="warning" :loading="actionLoading" @click="handleCancel">
              撤销
            </ElButton>
            <ElButton @click="handleComment">评论</ElButton>
          </div>
        </div>
      </template>

      <ElDescriptions :column="4" border size="small">
        <ElDescriptionsItem label="工单号">{{ ticket?.ticketNo ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="类型">{{ ticket?.typeName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="优先级">
          <ElTag v-if="ticket" :type="priorityMap[ticket.priority]?.type ?? 'info'" size="small">
            {{ priorityMap[ticket.priority]?.label ?? ticket.priority }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="流程">{{ ticket?.workflowName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="发起人">{{ ticket?.creatorName ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="创建时间">{{ ticket?.createdAt ?? '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="当前节点">
          {{ ticket?.status === 'pending' ? ticket?.currentNodeName || '-' : '已结束' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="结束时间">{{ ticket?.finishedAt ?? '-' }}</ElDescriptionsItem>
      </ElDescriptions>
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

      <!-- 审批记录（发起 → 节点/评论 → 结束，统一时间线） -->
      <ElCard class="card-wrapper" title="审批记录">
        <ElTimeline v-if="timelineItems.length" class="pl-2px">
          <ElTimelineItem
            v-for="item in timelineItems"
            :key="item.key"
            :timestamp="item.timeText"
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
                审批于 {{ fmt(item.node.finishedAt) }}
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

            <!-- 结束 -->
            <div v-else-if="item.kind === 'finish'" class="flex items-center gap-8px">
              <span class="font-medium">{{ item.finishLabel }}</span>
            </div>
          </ElTimelineItem>
        </ElTimeline>
        <ElEmpty v-else description="暂无审批记录" />
      </ElCard>
    </div>
  </div>
</template>
