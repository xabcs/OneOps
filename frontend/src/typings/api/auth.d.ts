declare namespace Api {
  /**
   * namespace Auth
   *
   * backend api module: "auth"
   */
  namespace Auth {
    interface LoginToken {
      token: string;
      user: UserInfo;
    }

    interface UserInfo {
      id: number;
      username: string;
      nickname: string;
      avatar?: string;
      email?: string;
      roleIds?: number[];
      status?: string | number;
      homePath?: string;
      createdAt?: string;
      updatedAt?: string;
      roleNames: string[];
      menuTree: any[];
      permissions: string[];
    }
  }
}
