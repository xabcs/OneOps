declare namespace Audit {
  /** 审计统计信息 */
  type Stats = {
    login: {
      total: number;
      success: number;
      failed: number;
      today: number;
      thisWeek: number;
      thisMonth: number;
    };
    operation: {
      total: number;
      success: number;
      failed: number;
    };
    system: {
      total: number;
      info: number;
      warning: number;
      error: number;
      critical: number;
    };
  };

  /** 登录日志 */
  type LoginLog = {
    id: number;
    userId: number;
    username: string;
    nickname: string;
    ip: string;
    userAgent: string;
    location: string;
    status: string;
    failReason: string;
    loginTime: string;
    logoutTime: string | null;
    duration: number;
    time: string;
    createdAt: string;
  };

  /** 操作日志 */
  type OperationLog = {
    id: number;
    userId: number;
    username: string;
    nickname: string;
    module: string;
    action: string;
    description: string;
    method: string;
    path: string;
    params: string;
    response: string;
    statusCode: number;
    ip: string;
    userAgent: string;
    duration: number;
    status: string;
    errorMsg: string;
    operateTime: string;
    time: string;
    createdAt: string;
  };

  /** 系统事件日志 */
  type SystemEventLog = {
    id: number;
    level: string;
    source: string;
    category: string;
    message: string;
    details: string;
    ip: string;
    eventTime: string;
    time: string;
    createdAt: string;
  };

  /** 日志查询参数 */
  type LogQuery = {
    username?: string;
    status?: string;
    module?: string;
    action?: string;
    level?: string;
    source?: string;
    category?: string;
    location?: string;
    startTime?: string;
    endTime?: string;
    page?: number;
    pageSize?: number;
  };
}
