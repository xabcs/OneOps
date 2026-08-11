/**
 * 菜单表格列定义
 */
import { Bottom, Plus, Top } from '@element-plus/icons-vue';
import { Icon } from '@iconify/vue';
import { $t } from '@/locales';
import { isFirst, isLast } from './menu-tree-helper';

export function createMenuColumns(handlers: {
  handleMove: (row: Api.SystemManage.Menu, direction: 'up' | 'down') => void;
  handleStatusChange: (row: Api.SystemManage.Menu, val: number) => void;
  handleAddChild: (row: Api.SystemManage.Menu) => void;
  handleEdit: (id: number) => void;
  handleDelete: (id: number) => void;
  originalTreeData: () => Api.SystemManage.Menu[];
}) {
  return () => [
    { prop: 'id', label: 'ID', width: 80, align: 'center', fixed: 'left' },
    {
      prop: 'hierarchyIndex',
      label: '序号',
      width: 100,
      align: 'center',
      formatter: (row: any) => {
        return <span class="hierarchy-index">{row.hierarchyIndex || '-'}</span>;
      }
    },
    { prop: 'name', label: '菜单名称', width: 200, align: 'center', className: 'menu-name-column' },
    {
      prop: 'menuType',
      label: '菜单类型',
      width: 100,
      align: 'center',
      formatter: (row: any) => {
        const menuType = row.menuType || 'menu';
        const typeMap: Record<string, { text: string; type: UI.ThemeColor }> = {
          directory: { text: '目录', type: 'primary' },
          menu: { text: '菜单', type: 'success' }
        };
        const config = typeMap[menuType] || { text: '菜单', type: 'info' };
        return (
          <ElTag size="small" type={config.type}>
            {config.text}
          </ElTag>
        );
      }
    },
    {
      prop: 'icon',
      label: '图标',
      width: 80,
      align: 'center',
      formatter: (row: any) => {
        if (row.icon) {
          return (
            <div class="flex-center">
              <Icon icon={row.icon} style="font-size: 18px" />
            </div>
          );
        }
        return <span class="text-tertiary">-</span>;
      }
    },
    { prop: 'path', label: '路由路径', minWidth: 180, align: 'center' },
    {
      prop: 'permission',
      label: '权限标识',
      minWidth: 150,
      align: 'center',
      formatter: (row: any) => {
        if (row.permission) {
          return (
            <ElTag size="small" type="info">
              {row.permission}
            </ElTag>
          );
        }
        return <span class="text-tertiary">-</span>;
      }
    },
    {
      prop: 'sort',
      label: '排序',
      width: 120,
      align: 'center',
      formatter: (row: Api.SystemManage.Menu) => (
        <div class="flex items-center justify-center gap-4px">
          <ElButton
            link
            type="primary"
            icon={Top}
            onClick={() => handlers.handleMove(row, 'up')}
            disabled={isFirst(row.id, handlers.originalTreeData())}
            title="上移"
          />
          <span class="sort-value">{row.sort}</span>
          <ElButton
            link
            type="primary"
            icon={Bottom}
            onClick={() => handlers.handleMove(row, 'down')}
            disabled={isLast(row.id, handlers.originalTreeData())}
            title="下移"
          />
        </div>
      )
    },
    {
      prop: 'status',
      label: '状态',
      width: 80,
      align: 'center',
      formatter: (row: any) => {
        if (row.status !== undefined) {
          return (
            <ElSwitch
              v-model={row.status}
              activeValue={1}
              inactiveValue={0}
              onChange={(val: number) => handlers.handleStatusChange(row, val)}
            />
          );
        }
        return null;
      }
    },
    {
      prop: 'operate',
      label: '操作',
      width: 280,
      align: 'center',
      fixed: 'right',
      formatter: (row: Api.SystemManage.Menu) => (
        <div class="flex-center gap-8px">
          {row.menuType === 'directory' && (
            <ElButton
              type="primary"
              plain
              size="small"
              icon={Plus}
              onClick={() => handlers.handleAddChild(row)}
              title="添加子菜单"
            >
              添加子菜单
            </ElButton>
          )}
          <ElButton type="primary" plain size="small" onClick={() => handlers.handleEdit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handlers.handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small">
                  {$t('common.delete')}
                </ElButton>
              )
            }}
          </ElPopconfirm>
        </div>
      )
    }
  ];
}
