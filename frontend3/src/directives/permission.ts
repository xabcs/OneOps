import type { DirectiveBinding } from 'vue';
import { ElMessage } from 'element-plus';
import { useAuthStore } from '@/store/modules/auth';

/**
 * 权限指令 v-permission（置灰模式）
 *
 * 设计原则：无权限一律置灰显示，不做隐藏——用户能看到功能入口，
 * 通过 tooltip / 点击提示得知缺失的权限码，便于向管理员申请开通。
 *
 * 使用示例（按钮请统一使用 PermissionButton 组件，指令用于其余交互元素）：
 * <ElDropdownItem v-permission="'k8s.resource.delete'">删除</ElDropdownItem>
 * <ElMenuItem v-permission="'cmdb.group.create'">新增分组</ElMenuItem>
 * <a v-permission="'cmdb.server.connect'" @click="handleConnect(row)">连接终端</a>
 * TSX 中指令不生效（会被当作普通 prop），需 withDirectives 手动挂载。
 *
 * 置灰实现：
 * - 原生 button/a：加 disabled 属性 + is-disabled 样式类（Element Plus 认此类名），
 *   并外包 span 承载 title 提示（disabled 元素不响应鼠标事件，tooltip/点击必须外置）
 * - 其他元素（如 ElDropdownItem 的 li）：is-disabled 样式类 + cursor: not-allowed，
 *   捕获阶段拦截 click 阻断交互，title 直接挂在元素上
 *
 * 点击反馈：点击置灰元素弹出 warning message 告知缺失的权限码
 */

/** 无权限提示（message），grouping 合并连续重复弹出的相同内容 */
function notifyMissingPermission(value: unknown) {
  ElMessage({
    type: 'warning',
    message: `缺少权限：${describePermission(value)}，请联系管理员在角色管理中开通`,
    grouping: true,
    showClose: true
  });
}

/** 记录置灰时插入的包裹元素，用于权限恢复后还原 DOM */
const wrappers = new WeakMap<HTMLElement, HTMLSpanElement>();

/** 记录各元素的无权限点击拦截器，用于权限恢复后解绑 */
const blockers = new WeakMap<HTMLElement, EventListener>();

function checkPermission(value: unknown): boolean {
  const authStore = useAuthStore();
  if (typeof value === 'string') return authStore.hasPermission(value);
  if (Array.isArray(value)) return authStore.hasAnyPermission(value);
  if (typeof value === 'object' && value !== null) {
    const { code, codes, mode = 'any' } = value as { code?: string; codes?: string[]; mode?: 'any' | 'all' };
    if (code) return authStore.hasPermission(code);
    if (codes && Array.isArray(codes)) {
      return mode === 'all' ? authStore.hasAllPermissions(codes) : authStore.hasAnyPermission(codes);
    }
  }
  return false;
}

/** 生成提示文案中的权限描述 */
function describePermission(value: unknown): string {
  if (typeof value === 'string') return value;
  if (Array.isArray(value)) return value.join(' 或 ');
  if (typeof value === 'object' && value !== null) {
    const { code, codes } = value as { code?: string; codes?: string[] };
    if (code) return code;
    if (codes && Array.isArray(codes)) return codes.join(' 或 ');
  }
  return '所需权限';
}

function applyDisabled(el: HTMLElement, binding: DirectiveBinding) {
  const tip = `缺少权限：${describePermission(binding.value)}\n请联系管理员在角色管理中开通`;
  el.classList.add('is-disabled');
  el.style.cursor = 'not-allowed';

  if (el.tagName === 'BUTTON' || el.tagName === 'A') {
    el.setAttribute('disabled', 'disabled');
    // 关键：原生 disabled button 不派发 click（也不冒泡），wrap 上的监听收不到；
    // 关掉指针事件让点击穿透到 wrap，由 wrap 弹出 message 提示
    el.style.pointerEvents = 'none';
    if (el.parentNode && !wrappers.has(el)) {
      const wrap = document.createElement('span');
      wrap.style.cursor = 'not-allowed';
      wrap.title = tip;
      wrap.addEventListener('click', () => notifyMissingPermission(binding.value));
      el.parentNode.insertBefore(wrap, el);
      wrap.appendChild(el);
      wrappers.set(el, wrap);
    }
  } else {
    if (!el.dataset.permissionDisabled) {
      el.dataset.permissionDisabled = '1';
      el.title = tip;
      // 捕获阶段拦截 click（先于组件自身 handler），并弹出提示
      const blocker = (e: Event) => {
        e.preventDefault();
        e.stopPropagation();
        notifyMissingPermission(binding.value);
      };
      blockers.set(el, blocker);
      el.addEventListener('click', blocker, true);
    }
  }
}

function removeDisabled(el: HTMLElement) {
  el.classList.remove('is-disabled');
  el.style.cursor = '';
  el.style.pointerEvents = '';

  const wrap = wrappers.get(el);
  if (wrap?.parentNode) {
    // 彻底解包：span 连同其 click 监听、title 一起脱离 DOM（无需 cloneNode，
    // 脱离文档的元素及其监听器会被 GC 回收，且不会残留过时的 title 提示）
    wrap.parentNode.insertBefore(el, wrap);
    wrap.parentNode.removeChild(wrap);
    wrappers.delete(el);
  }
  if (el.dataset.permissionDisabled) {
    delete el.dataset.permissionDisabled;
    el.removeAttribute('title');
    const blocker = blockers.get(el);
    if (blocker) {
      el.removeEventListener('click', blocker, true);
      blockers.delete(el);
    }
  }
  el.removeAttribute('disabled');
}

export default {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    if (!checkPermission(binding.value)) {
      applyDisabled(el, binding);
    }
  },
  updated(el: HTMLElement, binding: DirectiveBinding) {
    if (!checkPermission(binding.value)) {
      applyDisabled(el, binding);
    } else {
      removeDisabled(el);
    }
  }
};
