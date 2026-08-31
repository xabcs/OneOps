declare namespace Api {
  /**
   * namespace Ticket
   *
   * backend api module: "ticket"（工单系统）
   */
  namespace Ticket {
    /** 工单状态 */
    type TicketStatus = 'pending' | 'approved' | 'rejected' | 'canceled';

    /** 工单优先级 */
    type TicketPriority = 'low' | 'normal' | 'high' | 'urgent';

    /** 审批人类型 */
    type ApproverType = 'user' | 'role' | 'initiator';

    /** 多实例审批方式 */
    type MultiType = 'any' | 'all';

    /** 表单字段类型 */
    type FormFieldType = 'input' | 'textarea' | 'number' | 'select' | 'date';

    /** 动态表单字段定义 */
    type FormField = {
      key: string;
      label: string;
      type: FormFieldType;
      required: boolean;
      options?: string[];
      placeholder?: string;
    };

    /** 节点激活条件 */
    type ConditionItem = {
      field: string;
      op: 'eq' | 'ne' | 'in' | 'gt' | 'lt';
      value: string;
    };

    // ═══════════════ 流程定义 ═══════════════

    /** 流程节点 */
    type WorkflowNode = {
      id: number;
      workflowId: number;
      name: string;
      approverType: ApproverType;
      approverIds: string;
      multiType: MultiType;
      condition: string;
      /** 审批超时阈值（小时），0=不启用 */
      timeoutHours: number;
      sortOrder: number;
    };

    /** 流程定义（typeId：归属的工单场景） */
    type Workflow = {
      id: number;
      typeId: number;
      name: string;
      code: string;
      description: string;
      status: number;
      version: number;
      /** 审批节点数量（列表接口批量填充） */
      nodeCount?: number;
      createdAt: string;
      updatedAt: string;
      nodes?: WorkflowNode[];
    };

    /** 工单通知事件策略（事件矩阵一行：事件 × 渠道绑定） */
    type NotifyPolicy = {
      event: string;
      name: string;
      desc: string;
      /** 是否已定制（false=默认：该事件走全部启用渠道） */
      configured: boolean;
      enabled: number;
      channels: number[];
      /** 标题模板（空=内置默认文案） */
      titleTpl: string;
      /** 正文模板（空=内置默认文案） */
      bodyTpl: string;
      /** 是否已自定义模板 */
      hasTpl: boolean;
    };

    /** 夜间静默期配置（防轰炸，单行配置） */
    type NotifyQuietConfig = {
      quietEnabled: number;
      quietStartHour: number;
      quietEndHour: number;
    };

    /** 通知发送记录（外部渠道投递结果） */
    type NotifyLog = {
      id: number;
      event: string;
      ticketId: number;
      ticketNo: string;
      channelType: string;
      channelName: string;
      recipient: string;
      status: number;
      error: string;
      createdAt: string;
    };

    /** 工单站内消息（本人收件箱） */
    type TicketMessage = {
      id: number;
      userId: number;
      ticketId: number;
      ticketNo: string;
      event: string;
      title: string;
      content: string;
      isRead: number;
      createdAt: string;
    };

    /** 用户通知偏好（本人自助） */
    type UserNotifySetting = {
      dingtalkId: string;
      wechatId: string;
      /** 完全屏蔽的事件（站内与外部渠道均不发） */
      mutedEvents: string[];
      /** 停用的外部渠道（站内消息照常落） */
      offChannels: string[];
    };

    /** 流程保存请求 */
    type WorkflowSaveRequest = {
      typeId: number;
      name: string;
      code: string;
      description: string;
      status: number;
      nodes: {
        name: string;
        approverType: ApproverType;
        approverIds: number[];
        multiType: MultiType;
        condition: ConditionItem[];
        timeoutHours: number;
      }[];
    };

    /** 场景下的流程选项（发起工单用） */
    type WorkflowOption = {
      id: number;
      name: string;
      code: string;
      description: string;
      nodeCount: number;
    };

    // ═══════════════ 工单类型 ═══════════════

    /** 工单类型（工单场景） */
    type TicketType = {
      id: number;
      name: string;
      code: string;
      icon: string;
      description: string;
      formSchema: string;
      status: number;
      /** 该场景下启用的审批流程数量（0=发起工单时无流程可选） */
      workflowCount: number;
      createdAt: string;
      updatedAt: string;
    };

    /** 类型选项（发起工单用） */
    type TicketTypeOption = {
      id: number;
      name: string;
      code: string;
      icon: string;
      description: string;
      formSchema: string;
    };

    // ═══════════════ 工单 ═══════════════

    /** 工单列表项 */
    type TicketItem = {
      id: number;
      ticketNo: string;
      title: string;
      typeId: number;
      typeName: string;
      workflowId: number;
      workflowName: string;
      currentNodeKey: string;
      currentNodeName: string;
      status: TicketStatus;
      priority: TicketPriority;
      creatorId: number;
      creatorName: string;
      finishedAt: string | null;
      lastUrgeAt: string | null;
      createdAt: string;
      updatedAt: string;
      currentApprovers?: string;
      canApprove?: boolean;
    };

    /** 工单列表查询 */
    type TicketSearchParams = {
      page: number;
      pageSize: number;
      scope: 'created' | 'todo' | 'done' | 'all';
      status?: string;
      typeId?: number;
      keyword?: string;
    };

    /** 发起工单请求（workflowId：所选场景下的流程） */
    type TicketCreateRequest = {
      typeId: number;
      workflowId: number;
      title: string;
      priority: TicketPriority;
      formData: Record<string, unknown>;
    };

    /** 驳回后重新提交请求（可修改标题/优先级/表单） */
    type TicketResubmitRequest = {
      title: string;
      priority?: TicketPriority;
      formData: Record<string, unknown>;
    };

    /** 工单节点审批记录 */
    type TicketNodeRecord = {
      id: number;
      ticketId: number;
      nodeKey: string;
      nodeName: string;
      status: 'waiting' | 'pending' | 'approved' | 'rejected' | 'skipped' | 'canceled';
      approverIds: string;
      approverNames: string;
      approvedIds: string;
      multiType: MultiType;
      comment: string;
      startedAt: string | null;
      finishedAt: string | null;
      createdAt: string;
    };

    /** 流转日志 */
    type TicketFlowLog = {
      id: number;
      ticketId: number;
      action:
        | 'submit'
        | 'approve'
        | 'reject'
        | 'cancel'
        | 'resubmit'
        | 'reassign'
        | 'comment'
        | 'skip'
        | 'urge'
        | 'auto';
      nodeName: string;
      operatorId: number;
      operatorName: string;
      comment: string;
      createdAt: string;
    };

    /** 工单详情 */
    type TicketDetail = {
      ticket: TicketItem;
      formSchema: FormField[];
      formData: Record<string, unknown>;
      nodes: TicketNodeRecord[];
      logs: TicketFlowLog[];
      canApprove: boolean;
      canCancel: boolean;
      canResubmit: boolean;
      /** 审批中且发起人可催办（冷却已结束） */
      canUrge: boolean;
    };
  }
}
