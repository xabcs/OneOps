import { $t } from '@/locales';

/**
 * Transform record to option
 *
 * @example
 *   ```ts
 *   const record = {
 *     key1: 'label1',
 *     key2: 'label2'
 *   };
 *   const options = transformRecordToOption(record);
 *   // [
 *   //   { value: 'key1', label: 'label1' },
 *   //   { value: 'key2', label: 'label2' }
 *   // ]
 *   ```;
 *
 * @param record
 */
export function transformRecordToOption<T extends Record<string, string>>(record: T) {
  return Object.entries(record).map(([value, label]) => ({
    value,
    label
  })) as CommonType.Option<keyof T, T[keyof T]>[];
}

/**
 * Translate options
 *
 * @param options
 */
export function translateOptions(options: CommonType.Option<string, App.I18n.I18nKey>[]) {
  return options.map(option => ({
    ...option,
    label: $t(option.label)
  }));
}

/**
 * Toggle html class
 *
 * @param className
 */
export function toggleHtmlClass(className: string) {
  function add() {
    document.documentElement.classList.add(className);
  }

  function remove() {
    document.documentElement.classList.remove(className);
  }

  return {
    add,
    remove
  };
}

/** ElTag 组件可用的 type 取值 */
export type TagType = 'primary' | 'success' | 'info' | 'warning' | 'danger';

/**
 * 创建「状态 key → ElTag 标签」映射函数
 *
 * 用于收口各页面重复的 getStatusTag / getStatusType / getLevelType 类样板函数：
 * 传入业务各自的映射表，返回一个 (key) => { text, type } 的查询函数；
 * 未命中映射时回退为 { text: key || '-', type: 'info' }。
 *
 * @example
 *   ```ts
 *   const getStatusTag = createTagMap({
 *     active: { text: '活跃', type: 'success' },
 *     closed: { text: '已关闭', type: 'info' }
 *   });
 *   getStatusTag('active'); // { text: '活跃', type: 'success' }
 *   ```
 */
export function createTagMap<T extends string>(map: Partial<Record<T, { text: string; type: TagType }>>) {
  return (key: string | null | undefined): { text: string; type: TagType } =>
    (key && map[key as T]) || { text: key || '-', type: 'info' };
}
