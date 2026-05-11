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
  }
}
