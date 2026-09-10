import dayjs from 'dayjs';

/** 时间入参类型：字符串 / 时间戳 / Date / 空值 */
type TimeInput = string | number | Date | null | undefined;

/**
 * 格式化时间为指定格式字符串
 *
 * @param time 时间值（字符串、毫秒时间戳或 Date 对象）
 * @param format dayjs 格式化模板，默认 YYYY-MM-DD HH:mm:ss
 * @returns 格式化结果；空值或非法时间返回 '-'
 */
export function formatDateTime(time?: TimeInput, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (time === null || time === undefined || time === '') return '-';
  const date = dayjs(time);
  return date.isValid() ? date.format(format) : '-';
}

/**
 * 格式化为相对时间（各页面规则并集）：
 * 刚刚 / N分钟前 / N小时前 / N天前（7 天内）；超过 7 天显示 MM-DD HH:mm
 *
 * @param time 时间值（字符串、毫秒时间戳或 Date 对象）
 */
export function formatRelativeTime(time?: TimeInput): string {
  if (time === null || time === undefined || time === '') return '-';
  const date = dayjs(time);
  if (!date.isValid()) return '-';
  const diffSeconds = dayjs().diff(date, 'second');
  if (diffSeconds < 60) return '刚刚';
  if (diffSeconds < 3600) return `${Math.floor(diffSeconds / 60)}分钟前`;
  if (diffSeconds < 86400) return `${Math.floor(diffSeconds / 3600)}小时前`;
  if (diffSeconds < 7 * 86400) return `${Math.floor(diffSeconds / 86400)}天前`;
  return date.format('MM-DD HH:mm');
}

/**
 * 格式化持续时长（秒）为中文描述（紧凑版）：
 * X秒 / X分钟 / X小时X分（整小时时省略分钟）
 *
 * @param seconds 时长秒数；空值返回 '-'
 */
export function formatDuration(seconds?: number | null): string {
  if (seconds === null || seconds === undefined) return '-';
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return minutes > 0 ? `${hours}小时${minutes}分` : `${hours}小时`;
}

/**
 * 格式化持续时长（秒）为中文描述（口语版）：
 * X秒 / X分钟 / X小时X分钟
 *
 * @param seconds 时长秒数；空值或 0 返回 '-'
 */
export function formatDurationHuman(seconds?: number | null): string {
  if (!seconds) return '-';
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${minutes}分钟`;
}
