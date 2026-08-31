import { request } from '../request';

// ==================== 工单系统接口 ====================

/**
 * 工单列表
 * @param params scope: created=我发起的 / todo=待我审批 / done=我已审批 / all=全部
 */
export function fetchTickets(params: Api.Ticket.TicketSearchParams) {
  return request<Api.Common.PaginatingQueryRecord<Api.Ticket.TicketItem>>({
    url: '/ticket/tickets',
    method: 'get',
    params
  });
}

/** 发起工单 */
export function createTicket(data: Api.Ticket.TicketCreateRequest) {
  return request<{ id: number; ticketNo: string }>({
    url: '/ticket/tickets',
    method: 'post',
    data
  });
}

/** 工单详情 */
export function fetchTicketDetail(id: number) {
  return request<Api.Ticket.TicketDetail>({
    url: `/ticket/tickets/${id}`,
    method: 'get'
  });
}

/** 审批通过 */
export function approveTicket(id: number, comment: string) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/approve`,
    method: 'post',
    data: { comment }
  });
}

/** 驳回工单 */
export function rejectTicket(id: number, comment: string) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/reject`,
    method: 'post',
    data: { comment }
  });
}

/** 撤销工单（仅发起人） */
export function cancelTicket(id: number, comment: string) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/cancel`,
    method: 'post',
    data: { comment }
  });
}

/** 驳回后重新提交（仅发起人，可修改标题/优先级/表单） */
export function resubmitTicket(id: number, data: Api.Ticket.TicketResubmitRequest) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/resubmit`,
    method: 'post',
    data
  });
}

/** 改派当前节点审批人（需 ticket.ticket.reassign 权限） */
export function reassignTicket(id: number, approverIds: number[]) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/reassign`,
    method: 'post',
    data: { approverIds }
  });
}

/** 发起人催办（30 分钟冷却） */
export function urgeTicket(id: number) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/urge`,
    method: 'post'
  });
}

/** 工单评论 */
export function commentTicket(id: number, comment: string) {
  return request<boolean>({
    url: `/ticket/tickets/${id}/comment`,
    method: 'post',
    data: { comment }
  });
}

/** 启用的工单类型选项（发起工单用，仅需登录） */
export function fetchTicketTypeOptions() {
  return request<Api.Ticket.TicketTypeOption[]>({
    url: '/ticket/types/options',
    method: 'get'
  });
}

/** 某场景下启用的流程选项（发起工单用，仅需登录） */
export function fetchWorkflowsByType(typeId: number) {
  return request<Api.Ticket.WorkflowOption[]>({
    url: `/ticket/workflows/by-type/${typeId}`,
    method: 'get'
  });
}

// ─────────────── 工单类型管理（RBAC） ───────────────

/** 工单类型分页列表 */
export function fetchTicketTypes(params: { page: number; pageSize: number; keyword?: string; status?: number }) {
  return request<Api.Common.PaginatingQueryRecord<Api.Ticket.TicketType>>({
    url: '/ticket/types',
    method: 'get',
    params
  });
}

/** 创建工单类型 */
export function createTicketType(data: Partial<Api.Ticket.TicketType>) {
  return request<boolean>({
    url: '/ticket/types',
    method: 'post',
    data
  });
}

/** 更新工单类型 */
export function updateTicketType(id: number, data: Partial<Api.Ticket.TicketType>) {
  return request<boolean>({
    url: `/ticket/types/${id}`,
    method: 'put',
    data
  });
}

/** 删除工单类型 */
export function deleteTicketType(id: number) {
  return request<boolean>({
    url: `/ticket/types/${id}`,
    method: 'delete'
  });
}

// ─────────────── 流程定义管理（RBAC） ───────────────

/** 流程定义分页列表（typeId 可按场景过滤） */
export function fetchWorkflows(params: {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: number;
  typeId?: number;
}) {
  return request<Api.Common.PaginatingQueryRecord<Api.Ticket.Workflow>>({
    url: '/ticket/workflows',
    method: 'get',
    params
  });
}

/** 流程详情（含节点） */
export function fetchWorkflowDetail(id: number) {
  return request<Api.Ticket.Workflow>({
    url: `/ticket/workflows/${id}`,
    method: 'get'
  });
}

/** 创建流程 */
export function createWorkflow(data: Api.Ticket.WorkflowSaveRequest) {
  return request<boolean>({
    url: '/ticket/workflows',
    method: 'post',
    data
  });
}

/** 更新流程 */
export function updateWorkflow(id: number, data: Api.Ticket.WorkflowSaveRequest) {
  return request<boolean>({
    url: `/ticket/workflows/${id}`,
    method: 'put',
    data
  });
}

/** 删除流程 */
export function deleteWorkflow(id: number) {
  return request<boolean>({
    url: `/ticket/workflows/${id}`,
    method: 'delete'
  });
}

// ─────────────── 通知设置（事件矩阵，RBAC） ───────────────

/** 获取工单通知事件矩阵（事件 × 渠道绑定） */
export function fetchNotifyPolicies() {
  return request<Api.Ticket.NotifyPolicy[]>({
    url: '/ticket/notify-policies',
    method: 'get'
  });
}

/** 保存某通知事件的渠道绑定与启用状态 */
export function updateNotifyPolicy(
  event: string,
  data: { channels: number[]; enabled: number; titleTpl?: string; bodyTpl?: string }
) {
  return request<boolean>({
    url: `/ticket/notify-policies/${event}`,
    method: 'put',
    data
  });
}

/** 分页查询通知发送记录（排障：谁、哪个渠道、成功/失败） */
export function fetchNotifyLogs(params: PaginationParam & { event?: string; status?: number }) {
  return request<Api.Ticket.NotifyLog[]>({
    url: '/ticket/notify-logs',
    method: 'get',
    params
  });
}

// ─────────────── 站内消息（登录用户本人） ───────────────

/** 分页查询本人站内消息 */
export function fetchTicketMessages(params: PaginationParam & { unreadOnly?: number }) {
  return request<Api.Ticket.TicketMessage[]>({
    url: '/ticket/messages',
    method: 'get',
    params
  });
}

/** 本人未读消息数 */
export function fetchUnreadMessageCount() {
  return request<number>({
    url: '/ticket/messages/unread-count',
    method: 'get'
  });
}

/** 标记消息已读（ids 为空=全部已读） */
export function markTicketMessagesRead(ids: number[]) {
  return request<boolean>({
    url: '/ticket/messages/read',
    method: 'put',
    data: { ids }
  });
}

// ─────────────── 用户通知偏好（本人自助） ───────────────

/** 读取本人通知偏好 */
export function fetchMyNotifySetting() {
  return request<Api.Ticket.UserNotifySetting>({
    url: '/ticket/my-notify-setting',
    method: 'get'
  });
}

/** 保存本人通知偏好 */
export function updateMyNotifySetting(data: Api.Ticket.UserNotifySetting) {
  return request<boolean>({
    url: '/ticket/my-notify-setting',
    method: 'put',
    data
  });
}
