/**
 * 菜单树操作辅助函数
 */

/** 带层级序号的菜单节点（层级序号为前端运行时附加字段） */
export type MenuWithHierarchy = Api.SystemManage.Menu & { hierarchyIndex?: string };

/** 将树形结构转换为扁平列表 */
export function flattenMenuTree(tree: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  const result: Api.SystemManage.Menu[] = [];

  function traverse(list: Api.SystemManage.Menu[]) {
    list.forEach(item => {
      result.push({
        id: item.id,
        parentId: item.parentId,
        name: item.name,
        path: item.path,
        icon: item.icon,
        permission: item.permission,
        menuType: item.menuType,
        sort: item.sort,
        status: item.status,
        hierarchyIndex: (item as MenuWithHierarchy).hierarchyIndex
      } as Api.SystemManage.Menu);

      if (item.children && Array.isArray(item.children) && item.children.length > 0) {
        traverse(item.children);
      }
    });
  }

  traverse(tree);
  return result;
}

/** 清理菜单树并添加层级序号 */
export function cleanMenuTree(tree: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  const sortTree = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] => {
    return list
      .sort((a, b) => (a.sort || 0) - (b.sort || 0))
      .map(item => {
        const cleaned: any = {
          id: item.id,
          parentId: item.parentId,
          name: item.name,
          path: item.path,
          icon: item.icon,
          permission: item.permission,
          menuType: item.menuType,
          sort: item.sort,
          status: item.status
        };

        if (item.children && Array.isArray(item.children) && item.children.length > 0) {
          cleaned.children = sortTree(item.children);
        }

        return cleaned;
      });
  };

  const assignHierarchyIndex = (list: Api.SystemManage.Menu[], prefix: string = ''): Api.SystemManage.Menu[] => {
    let siblingIndex = 1;

    return list.map(item => {
      const currentItem = { ...item } as MenuWithHierarchy;
      currentItem.hierarchyIndex = prefix ? `${prefix}.${siblingIndex}` : `${siblingIndex}`;

      if (item.children && Array.isArray(item.children) && item.children.length > 0) {
        currentItem.children = assignHierarchyIndex(item.children, currentItem.hierarchyIndex);
      }

      siblingIndex++;
      return currentItem;
    });
  };

  const sortedTree = sortTree(tree);
  return assignHierarchyIndex(sortedTree);
}

/** 过滤菜单树 */
export function filterMenuTree(list: Api.SystemManage.Menu[], filterText: string): Api.SystemManage.Menu[] {
  if (!filterText) return list;

  const result: Api.SystemManage.Menu[] = [];

  list.forEach(item => {
    const nameMatch = item.name.toLowerCase().includes(filterText.toLowerCase());
    const pathMatch = item.path && item.path.toLowerCase().includes(filterText.toLowerCase());
    const permissionMatch = item.permission && item.permission.toLowerCase().includes(filterText.toLowerCase());

    const children = item.children && item.children.length > 0 ? filterMenuTree(item.children, filterText) : [];
    const childrenMatch = children.length > 0;

    if (nameMatch || pathMatch || permissionMatch || childrenMatch) {
      const newItem = { ...item } as Api.SystemManage.Menu;
      if (children.length > 0) {
        newItem.children = children;
      }
      result.push(newItem);
    }
  });

  return result;
}

/** 查找兄弟节点 */
export function findSiblings(targetId: number, originalTreeData: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] | null {
  const findInList = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] | null => {
    for (let i = 0; i < list.length; i++) {
      if (list[i].id === targetId) return list;
      if (list[i].children && list[i].children.length > 0) {
        const found = findInList(list[i].children);
        if (found) return found;
      }
    }
    return null;
  };
  return findInList(originalTreeData);
}

/** 查找子菜单 */
export function findChildren(parentId: number, originalTreeData: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] {
  const findInList = (list: Api.SystemManage.Menu[]): Api.SystemManage.Menu[] => {
    for (const item of list) {
      if (item.id === parentId) {
        return item.children || [];
      }
      if (item.children && item.children.length > 0) {
        const found = findInList(item.children);
        if (found.length > 0) return found;
      }
    }
    return [];
  };
  return findInList(originalTreeData);
}

/** 判断是否是第一个兄弟节点 */
export function isFirst(id: number, originalTreeData: Api.SystemManage.Menu[]): boolean {
  const siblings = findSiblings(id, originalTreeData);
  if (!siblings || siblings.length === 0) return true;
  return siblings[0].id === id;
}

/** 判断是否是最后一个兄弟节点 */
export function isLast(id: number, originalTreeData: Api.SystemManage.Menu[]): boolean {
  const siblings = findSiblings(id, originalTreeData);
  if (!siblings || siblings.length === 0) return true;
  return siblings[siblings.length - 1].id === id;
}
