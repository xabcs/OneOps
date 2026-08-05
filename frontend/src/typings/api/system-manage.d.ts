declare namespace Api {
  /**
   * namespace SystemManage
   *
   * backend api module: "system"
   */
  namespace SystemManage {
    type CommonSearchParams = Pick<Common.PaginatingCommonParams, 'current' | 'size'>;

    /** role */
    type Role = Common.CommonRecord<{
      /** role name */
      name: string;
      /** role code */
      code: string;
      /** role description */
      description: string;
      /** menu ids (JSON string) */
      menuIds: string;
      /** status */
      status: number;
    }>;

    /** role search params */
    type RoleSearchParams = CommonType.RecordNullable<
      Pick<Api.SystemManage.Role, 'name' | 'code' | 'status'> & CommonSearchParams
    >;

    /** role list */
    type RoleList = Common.PaginatingQueryRecord<Role>;

    /** all role */
    type AllRole = Pick<Role, 'id' | 'name' | 'code' | 'menuIds'>;

    /**
     * user gender
     *
     * - "1": "male"
     * - "2": "female"
     */
    type UserGender = '1' | '2';

    /** user */
    type User = Common.CommonRecord<{
      /** user name */
      username: string;
      /** user nick name */
      nickname: string;
      /** user avatar */
      avatar: string;
      /** user email */
      email: string;
      /** user role id collection */
      roleIds: number[];
      /** user status */
      status: string;
      /** user home path */
      homePath: string;
    }>;

    /** user search params */
    type UserSearchParams = CommonType.RecordNullable<
      Pick<Api.SystemManage.User, 'username' | 'nickname' | 'email' | 'status'> & CommonSearchParams
    >;

    /** user list */
    type UserList = Common.PaginatingQueryRecord<User>;

    /**
     * menu type
     *
     * - "1": directory
     * - "2": menu
     */
    type MenuType = '1' | '2';

    type MenuButton = {
      /**
       * button code
       *
       * it can be used to control the button permission
       */
      code: string;
      /** button description */
      desc: string;
    };

    /**
     * icon type
     *
     * - "1": iconify icon
     * - "2": local icon
     */
    type IconType = '1' | '2';

    type MenuPropsOfRoute = Pick<
      import('vue-router').RouteMeta,
      | 'i18nKey'
      | 'keepAlive'
      | 'constant'
      | 'order'
      | 'href'
      | 'hideInMenu'
      | 'activeMenu'
      | 'multiTab'
      | 'fixedIndexInTab'
      | 'query'
    >;

    type Menu = Common.CommonRecord<{
      /** parent menu id */
      parentId: number;
      /** menu name */
      name: string;
      /** route path */
      path: string;
      /** iconify icon name or local icon name */
      icon: string;
      /** permission */
      permission: string;
      /** menu type: menu or directory */
      menuType: string;
      /** sort order */
      sort: number;
      /** status */
      status: number;
      /** children menu */
      children?: Menu[] | null;
    }>;

    /** menu list */
    type MenuList = Common.PaginatingQueryRecord<Menu>;

    type MenuTree = {
      id: number;
      name: string;
      parentId: number;
      icon?: string;
      path?: string;
      permission?: string;
      menuType?: string;
      sort?: number;
      status?: number;
      children: MenuTree[];
    };

    /** home directory option */
    type HomeDirectoryOption = {
      id: number;
      name: string;
      path: string;
    };

    /** 属性定义 */
    type AttributeDefinition = {
      id: number;
      name: string;
      key: string;
      category: AttributeCategory;
      type: AttributeType;
      options: string;
      required: boolean;
      defaultValue: string;
      sortOrder: number;
      status: number;
      description: string;
      createdAt: string;
      updatedAt: string;
    };

    /** 属性定义表单 */
    type AttributeDefinitionForm = {
      id?: number;
      name: string;
      key: string;
      category: AttributeCategory;
      type: AttributeType;
      options?: string;
      required?: boolean;
      defaultValue?: string;
      sortOrder?: number;
      description?: string;
    };

    /** 属性分类 */
    type AttributeCategory =
      | 'system' // 系统分类
      | 'location' // 地理位置
      | 'environment' // 环境信息
      | 'hardware' // 硬件配置
      | 'custom'; // 自定义

    /** 属性类型 */
    type AttributeType =
      | 'text' // 单行文本
      | 'select' // 下拉单选
      | 'multiselect' // 下拉多选
      | 'number' // 数字
      | 'date' // 日期
      | 'boolean'; // 布尔值

    /** 属性选项 */
    type AttributeOption = {
      value: string;
      label: string;
    };

    /** 主机属性值 */
    type ServerAttribute = {
      id: number;
      serverId: number;
      attributeId: number;
      attributeKey: string;
      attributeValue: string;
      valueType: string;
      category: string;
      createdAt: string;
      updatedAt: string;
      definition?: AttributeDefinition;
    };

    /** 权限 */
    type Permission = Common.CommonRecord<{
      /** 权限名称 */
      name: string;
      /** 权限编码 */
      code: string;
      /** 权限描述 */
      description: string;
      /** 所属模块 */
      module: string;
      /** 资源名称 */
      resource: string;
      /** 操作名称 */
      action: string;
      /** 权限级别 */
      level: number;
      /** 父权限ID */
      parentId?: number | null;
      /** 排序 */
      sortOrder: number;
      /** 状态 */
      status: number;
      /** 子权限 */
      children?: Permission[] | null;
    }>;

    /** 权限搜索参数 */
    type PermissionSearchParams = CommonType.RecordNullable<
      Pick<Api.SystemManage.Permission, 'name' | 'code' | 'module' | 'level' | 'status'> & CommonSearchParams
    >;

    /** 权限列表 */
    type PermissionList = Common.PaginatingQueryRecord<Permission>;

    /** 权限树 */
    type PermissionTree = {
      id: number;
      name: string;
      code: string;
      parentId?: number | null;
      level: number;
      children: PermissionTree[];
    };

    /** 角色权限分配请求 */
    type AssignRolePermissionsRequest = {
      roleId: number;
      permissionIds: number[];
    };

    /** 权限检查请求 */
    type CheckPermissionRequest = {
      userId: number;
      permission: string;
    };

    /** 权限检查响应 */
    type CheckPermissionResponse = {
      allowed: boolean;
      reason?: string;
    };
  }
}
