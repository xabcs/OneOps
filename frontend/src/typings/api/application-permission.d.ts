declare namespace Api {
  /**
   * namespace ApplicationPermission
   *
   * backend api module: "system"
   */
  namespace ApplicationPermission {
    type CommonSearchParams = {
      current: number;
      size: number;
    };

    /** API endpoints configuration */
    type ApplicationEndpoints = {
      /** Create user endpoint, e.g., /api/users (POST) */
      createUser: string;
      /** List users endpoint, e.g., /api/users (GET) */
      listUsers: string;
      /** Get single user endpoint, e.g., /api/users/{username} (GET) */
      getUser: string;
      /** Get roles endpoint, e.g., /api/roles (GET) */
      getRoles: string;
      /** Assign role to user endpoint, e.g., /api/users/{username}/roles (POST) */
      assignRole: string;
      /** Revoke role from user endpoint, e.g., /api/users/{username}/roles/{roleCode} (DELETE) */
      revokeRole: string;
      /** Get user roles endpoint, e.g., /api/users/{username}/roles (GET) */
      getUserRoles: string;
    };

    /** Authentication configuration */
    type ApplicationAuthConfig = {
      /** Auth type: basic, token, oauth2 */
      type: 'basic' | 'token' | 'oauth2';
      /** Username for basic auth */
      username?: string;
      /** Password for basic auth */
      password?: string;
      /** API token */
      token?: string;
    };

    /** application */
    type Application = Common.CommonRecord<{
      /** application name */
      name: string;
      /** application code */
      code: string;
      /** application type */
      type: string;
      /** base URL */
      baseUrl: string;
      /** endpoints config (JSON string) */
      endpoints: string;
      /** auth config (JSON string) */
      authConfig: string;
      /** description */
      description: string;
      /** sync interval */
      syncInterval: number;
      /** last sync time */
      lastSyncTime: string;
      /** status */
      status: number;
    }>;

    /** application form data */
    type ApplicationFormData = {
      id?: number;
      name: string;
      code: string;
      type: string;
      baseUrl: string;
      endpoints: ApplicationEndpoints;
      authConfig: ApplicationAuthConfig;
      description: string;
      syncInterval: number;
    };

    /** auth user (授权中心用户) */
    type AuthUser = Common.CommonRecord<{
      /** username */
      username: string;
      /** nickname */
      nickname: string;
      /** email */
      email: string;
      /** phone */
      phone: string;
      /** description */
      description: string;
      /** status */
      status: number;
    }>;

    /** auth group (授权中心用户组) */
    type AuthGroup = Common.CommonRecord<{
      /** group name */
      name: string;
      /** group code */
      code: string;
      /** description */
      description: string;
      /** status */
      status: number;
    }>;

    /** application role */
    type ApplicationRole = Common.CommonRecord<{
      /** role code */
      roleCode: string;
      /** role name */
      roleName: string;
      /** role type */
      roleType: string;
      /** description */
      description: string;
      /** sync time */
      syncTime: string;
    }>;

    /** application user */
    type ApplicationUser = Common.CommonRecord<{
      /** username */
      username: string;
      /** display name */
      displayName: string;
      /** email */
      email: string;
      /** status */
      status: string;
      /** sync time */
      syncTime: string;
    }>;

    /** application group */
    type ApplicationGroup = Common.CommonRecord<{
      /** group code */
      groupCode: string;
      /** group name */
      groupName: string;
      /** description */
      description: string;
      /** sync time */
      syncTime: string;
    }>;

    /** application authorization rule (应用授权规则) */
    type AuthorizationRule = Common.CommonRecord<{
      /** rule ID (from external system) */
      ruleId: string;
      /** rule name */
      ruleName: string;
      /** rule type: user, group */
      ruleType: string;
      /** subject type: user, group */
      subjectType: string;
      /** subject ID (user ID or group ID) */
      subjectId: string;
      /** subject name (for display) */
      subjectName: string;
      /** object type: asset, node, system */
      objectType: string;
      /** object ID (asset ID or * for all) */
      objectId: string;
      /** object name (for display) */
      objectName: string;
      /** actions (JSON array of action names) */
      actions: string;
      /** priority */
      priority: number;
      /** is enabled */
      isEnabled: boolean;
      /** is expired */
      isExpired: boolean;
      /** expire time */
      expireTime: string;
      /** sync time */
      syncTime: string;
    }>;

    /** group binding (用户组权限绑定) */
    type GroupBinding = Common.CommonRecord<{
      /** local group id */
      groupId: number;
      /** application id */
      appId: number;
      /** application role id */
      applicationRoleId: number;
    }>;

    /** auth user group (授权中心用户组成员) */
    type AuthUserGroup = Common.CommonRecord<{
      /** user id */
      userId: number;
      /** group id */
      groupId: number;
      /** granted by */
      grantedBy: string;
      /** granted at */
      grantedAt: string;
    }>;

    /** operation log */
    type ApplicationOperationLog = Common.CommonRecord<{
      /** operation type */
      operation: string;
      /** target */
      target: string;
      /** status */
      status: string;
      /** operator */
      operator: string;
      /** created at */
      createdAt: string;
    }>;

    /** application search params */
    type ApplicationSearchParams = CommonType.RecordNullable<
      Pick<Api.ApplicationPermission.Application, 'name'> & CommonSearchParams
    >;

    /** application list */
    type ApplicationList = {
      records: Application[];
      total: number;
    };

    /** operation log list */
    type OperationLogList = {
      records: ApplicationOperationLog[];
      total: number;
    };

    /** user identity mapping (用户身份映射) */
    type UserIdentityMapping = Common.CommonRecord<{
      /** auth user id */
      authUserId: number;
      /** auth user */
      authUser?: AuthUser;
      /** application id */
      appId: number;
      /** application */
      appIDField?: Application;
      /** external username */
      externalUsername: string;
      /** external user id */
      externalUserId?: string;
      /** mapping type: auto, manual */
      mappingType: 'auto' | 'manual';
      /** mapping status: active, inactive, deleted */
      mappingStatus: 'active' | 'inactive' | 'deleted';
      /** last sync time */
      lastSyncTime?: string;
    }>;

    /** identity mapping list */
    type IdentityMappingList = {
      records: UserIdentityMapping[];
      total: number;
    };

    /** user effective permission (用户有效权限) */
    type UserEffectivePermission = {
      /** username */
      username: string;
      /** nickname */
      nickname?: string;
      /** application id */
      appId: number;
      /** application name */
      appName: string;
      /** role code */
      roleCode: string;
      /** role name */
      roleName: string;
      /** role type */
      roleType: string;
      /** status: active, inactive, expired, pending */
      status: 'active' | 'inactive' | 'expired' | 'pending';
      /** external username */
      externalUsername?: string;
      /** group id (if assigned via group) */
      groupId?: number;
      /** group name */
      groupName?: string;
      /** group code */
      groupCode?: string;
      /** assigned at */
      assignedAt?: string;
    };

    /** user permission list */
    type UserPermissionList = {
      records: UserEffectivePermission[];
      total: number;
    };

    /** group binding execution (权限绑定执行记录) */
    type GroupBindingExecution = Common.CommonRecord<{
      /** group binding id */
      groupBindingId: number;
      /** auth user id */
      authUserId: number;
      /** auth user */
      authUser?: AuthUser;
      /** external username */
      externalUsername?: string;
      /** action type: created, granted, removed, failed */
      actionType: 'created' | 'granted' | 'removed' | 'failed';
      /** status: success, failed, pending */
      status: 'success' | 'failed' | 'pending';
      /** message */
      message?: string;
      /** operator */
      operator: string;
    }>;
  }
}
