/**
 * Arthas 命令输出解析器（纯函数，无副作用）
 *
 * 只做"文本 → 结构化数据"，不做判定；解析失败返回空结果由上层降级为 raw 展示。
 * 注意：Arthas 输出可能含 ANSI 颜色码，统一先 stripAnsi。
 */

/** 清洗 ANSI 转义码与控制字符 */
export function stripAnsi(s: string): string {
  return (
    s
      .replace(/\r\n/g, '\n')
      // oxlint-disable-next-line no-control-regex
      .replace(/\x1B\[[0-9;]*[a-zA-Z]/g, '')
      // oxlint-disable-next-line no-control-regex
      .replace(/[\x00-\x08\x0B\x0C\x0E-\x1F]/g, '')
  );
}

// ========== thread 系列 ==========

export interface ThreadRow {
  name: string;
  id: string;
  state: string;
  /** CPU 占比（%），无数据为 null */
  cpu: number | null;
  stack: string[];
  /** waiting to lock 的锁标识 */
  waitingLock?: string;
  /** 该线程等待的锁被谁持有（"locked by"） */
  lockedBy?: string;
  /** 该线程持有的锁 */
  locked: string[];
}

export interface ThreadParse {
  /** 顶部统计行原文（如 Threads Total: 45, ...），可能不存在 */
  header?: string;
  threads: ThreadRow[];
}

/** 线程头：`"main" Id=1 RUNNABLE (cpu 82.30%)` */
const THREAD_HEAD_RE = /^"([^"]+)"\s+Id=([0-9a-fx]+)\s+([A-Z_]+)(?:[^\n]*?\(cpu\s+([\d.]+)%\))?/i;

export function parseThreads(output: string): ThreadParse {
  const res: ThreadParse = { threads: [] };
  const lines = stripAnsi(output).split('\n');
  let cur: ThreadRow | null = null;

  for (const rawLine of lines) {
    const line = rawLine.trimEnd();

    // 统计头行
    if (/^Threads\s+Total:/i.test(line)) {
      res.header = line.trim();
      continue;
    }

    const head = line.match(THREAD_HEAD_RE);
    if (head && !/^\s/.test(rawLine)) {
      if (cur) res.threads.push(cur);
      cur = {
        name: head[1],
        id: head[2],
        state: head[3],
        cpu: head[4] ? Number(head[4]) : null,
        stack: [],
        locked: []
      };
      continue;
    }

    if (!cur) continue;

    // 锁信息行（"- waiting to lock <0x...> (a java.xx.Obj)" / "- locked <0x..>" / "locked by ..."）
    if (/waiting to lock/i.test(line)) {
      const lock = line.match(/<[^>]+>|@[\da-fA-F]+/);
      cur.waitingLock = lock ? lock[0] : line.trim();
      continue;
    }
    if (/locked by/i.test(line)) {
      const owner = line.match(/"([^"]+)"/);
      cur.lockedBy = owner ? owner[1] : line.trim();
      continue;
    }
    if (/^\s*-\s*locked\b/i.test(rawLine)) {
      const lock = line.match(/<[^>]+>|@[\da-fA-F]+/);
      if (lock) cur.locked.push(lock[0]);
      continue;
    }

    // 堆栈帧（缩进 + at 开头，或任意缩进行归入堆栈）
    if (/^\s{2,}/.test(rawLine) && line.trim()) {
      cur.stack.push(line.trim());
    }
  }
  if (cur) res.threads.push(cur);
  return res;
}

// ========== memory ==========

export interface MemoryRegion {
  name: string;
  used: string;
  total: string;
  max: string;
  /** 使用率（0~100），无数据为 null */
  percent: number | null;
}

/** 表格行：`heap  32M  256M  4096M  0.79%`（列间多空格，max 可为 -/-1） */
export function parseMemory(output: string): MemoryRegion[] {
  const regions: MemoryRegion[] = [];
  for (const line of stripAnsi(output).split('\n')) {
    const m = line.trim().match(/^([\w.-]+)\s+([\d.]+[KMG]?|-)\s+([\d.]+[KMG]?|-)\s+(\S+)\s*$/);
    if (!m) continue;
    // 跳过表头
    if (m[1].toLowerCase() === 'memory') continue;
    const pct = m[4].match(/^([\d.]+)%$/);
    regions.push({
      name: m[1],
      used: m[2],
      total: m[3],
      max: m[4],
      percent: pct ? Number(pct[1]) : null
    });
  }
  return regions;
}

// ========== jvm ==========

export interface JvmGroup {
  title: string;
  kv: { key: string; value: string }[];
}

/** 分段结构：`RUNTIME` 标题行 + `----` 分隔线 + `key   value` 行 */
export function parseJvm(output: string): JvmGroup[] {
  const groups: JvmGroup[] = [];
  let cur: JvmGroup | null = null;
  for (const line of stripAnsi(output).split('\n')) {
    const t = line.trim();
    if (!t) continue;
    if (/^-{4,}$/.test(t)) continue;
    // 段标题：全大写字母/数字/下划线/连字符（如 RUNTIME / GARBAGE-COLLECTORS / CLASS-LOADING）
    if (/^[A-Z0-9][A-Z0-9_-]+$/.test(t)) {
      cur = { title: t, kv: [] };
      groups.push(cur);
      continue;
    }
    if (!cur) continue;
    // key 与 value 间多空格对齐；value 可能含空格
    const parts = t.split(/\s{2,}/);
    if (parts.length >= 2) {
      cur.kv.push({ key: parts[0], value: parts.slice(1).join('  ') });
    } else if (parts.length === 1) {
      cur.kv.push({ key: parts[0], value: '' });
    }
  }
  return groups;
}

// ========== sc -d ==========

export interface ScClassInfo {
  kv: { key: string; value: string }[];
  /** class-info 行提取的类全名 */
  className?: string;
}

/** 多类时多块；`class-info  com.xx.Xxx` 行开新块 */
export function parseSc(output: string): ScClassInfo[] {
  const blocks: ScClassInfo[] = [];
  let cur: ScClassInfo | null = null;
  for (const line of stripAnsi(output).split('\n')) {
    const t = line.trimEnd();
    if (!t.trim()) continue;
    const parts = t.split(/\s{2,}/).map(p => p.trim());
    if (parts[0] === 'class-info') {
      cur = { kv: [], className: parts.slice(1).join(' ') || undefined };
      blocks.push(cur);
      continue;
    }
    if (!cur) continue;
    if (parts.length >= 2) {
      cur.kv.push({ key: parts[0], value: parts.slice(1).join('  ') });
    }
  }
  return blocks;
}

// ========== heapdump / logger ==========

/** 从输出中提取堆转储文件路径 */
export function parseHeapdumpFile(output: string): string | null {
  const m = stripAnsi(output).match(/(\S+\.(?:hprof|phd)\S*)/);
  return m ? m[1].replace(/[.,;)]+$/, '') : null;
}

/** 从命令中提取 --name/--level 参数值 */
export function parseLoggerCommand(command: string): { name: string; level: string } {
  const name = command.match(/--name\s+(\S+)/)?.[1] ?? 'ROOT';
  const level = command.match(/--level\s+(\S+)/)?.[1] ?? '';
  return { name, level };
}
