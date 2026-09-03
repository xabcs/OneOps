/**
 * 判定规则引擎 + Finding 构建入口
 *
 * 原则：
 * 1. 规则只作用于解析后的结构化数据，永不碰原始文本（解析失败降级 raw，不误报）
 * 2. 根因是"候选"（启发式提取首个业务帧），界面必须标注，由人裁决
 * 3. 数据不足走 unknown（JMC 的 Not Applicable 思想），宁可未判定不可误报
 *
 * 阈值参考：awesome-prometheus-alerts（JVM 节）+ JMC 规则语义
 * 模式参考：fastthread（IO 等待识别、线程分组）+ 社区（idle worker 栈顶模式）
 */

import type { DiagnosticFinding, FindingAction, FindingRow, FindingContext } from './types';
import {
  parseThreads,
  parseMemory,
  parseJvm,
  parseSc,
  parseHeapdumpFile,
  parseLoggerCommand,
  type ThreadRow
} from './parsers';

// ========== 共享启发式 ==========

/** 框架/JDK 类前缀：提取业务帧时跳过 */
const FRAMEWORK_PREFIXES = [
  'java.',
  'javax.',
  'jdk.',
  'sun.',
  'com.sun.',
  'org.springframework.',
  'org.apache.',
  'io.netty.',
  'com.alibaba.',
  'com.mysql.',
  'org.postgresql.',
  'redis.clients.',
  'org.mongodb.',
  'okhttp3.',
  'org.asynchttpclient.',
  'org.hibernate.',
  'org.mybatis.',
  'ch.qos.logback.',
  'org.apache.logging.'
];

/** IO 等待的堆栈特征（RUNNABLE ≠ CPU 消耗，native IO 阻塞也显示 RUNNABLE） */
const IO_WAIT_PATTERNS = [
  'socketRead0',
  'socketWrite0',
  'socketAccept',
  'epollWait',
  'poll0',
  'kevent0',
  'connect0',
  'waitForReferencePendingList'
];

/** idle worker 的栈顶模式（线程池空闲而非异常） */
const IDLE_WORKER_PATTERNS = [
  'ThreadPoolExecutor.getTask',
  'LinkedBlockingQueue.take',
  'LinkedTransferQueue.take',
  'ArrayBlockingQueue.take',
  'LockSupport.park'
];

/** 判定某线程堆栈是否为 IO 等待（看前几帧） */
function isIoWait(t: ThreadRow): boolean {
  const head = t.stack.slice(0, 5).join(' ');
  return IO_WAIT_PATTERNS.some(p => head.includes(p));
}

/** 判定某线程是否为空闲 worker */
function isIdleWorker(t: ThreadRow): boolean {
  if (t.stack.length === 0) return false;
  const head = t.stack.slice(0, 3).join(' ');
  const isPool = IDLE_WORKER_PATTERNS.some(p => head.includes(p));
  return isPool && (t.cpu === null || t.cpu < 5);
}

/** 从堆栈中提取第一个业务代码帧（根因候选） */
function firstBizFrame(stack: string[]): string | null {
  for (const frame of stack) {
    const m = frame.match(/(?:at\s+)?(\S+?)\(([\w$.]+)?(?:\.java)?:(\d+|-)\)/);
    if (!m) continue;
    if (FRAMEWORK_PREFIXES.some(p => m[1].startsWith(p))) continue;
    return m[1];
  }
  return null;
}

/** 解析业务帧为 类名 + 方法名（用于生成联动命令） */
function splitFrame(bizFrame: string): { cls: string; method: string } | null {
  // com.demo.OrderService.createOrder → 类 + 方法
  const m = bizFrame.match(/^(.+)\.([^.]+)$/);
  if (!m) return null;
  return { cls: m[1], method: m[2] };
}

/** 线程名去掉尾部编号用于分组（http-nio-8080-exec-12 → http-nio-8080-exec） */
function threadGroupKey(name: string): string {
  return name.replace(/[\d-]*\d+$/, '').replace(/-$/, '') || name;
}

/** 业务工作线程名特征（用于保守推断影响面） */
const WORKER_NAME_RE = /(http|exec|worker|biz|task|qtp|web|nio)/i;

// ========== 场景规则 ==========

/** CPU 飙升定位（thread -n N）：判定型 Finding */
function ruleCpuTop(ctx: FindingContext): DiagnosticFinding {
  const { threads } = parseThreads(ctx.output);
  if (threads.length === 0) return parseFailed();

  const sorted = [...threads].sort((a, b) => (b.cpu ?? -1) - (a.cpu ?? -1));
  const top = sorted[0];
  const topCpu = top.cpu;

  // IO 等待识别：高"CPU"但栈在 native IO → 实为 IO 等待，非计算热点
  const ioWait = topCpu !== null && topCpu >= 20 && isIoWait(top);
  const bizFrame = top ? firstBizFrame(top.stack) : null;
  const frameParts = bizFrame ? splitFrame(bizFrame) : null;

  // 判定级别（IO 等待场景降级：非计算瓶颈）
  let level: DiagnosticFinding['level'] = 'ok';
  let symptom = '';
  if (topCpu === null) {
    level = 'unknown';
    symptom = '未取到线程 CPU 数据，请查看原始输出';
  } else if (topCpu >= 50 && !ioWait) {
    level = 'critical';
    symptom = `最高线程 ${top.name} CPU ${topCpu}%（显著计算热点）`;
  } else if (topCpu >= 20 && !ioWait) {
    level = 'warning';
    symptom = `最高线程 ${top.name} CPU ${topCpu}%（偏高）`;
  } else if (ioWait) {
    level = 'warning';
    symptom = `线程 ${top.name} 占用 ${topCpu}% 但堆栈处于 IO 等待（如 DB/网络读），非计算热点`;
  } else {
    symptom = `最高线程 CPU ${topCpu}%，未见明显计算热点`;
  }

  // 批量热点：同组线程多个高占用 → 线程池任务堆积方向
  const byGroup = new Map<string, { count: number; maxCpu: number }>();
  for (const t of sorted) {
    if (t.cpu === null || t.cpu < 20) continue;
    const key = threadGroupKey(t.name);
    const cur = byGroup.get(key) ?? { count: 0, maxCpu: 0 };
    cur.count += 1;
    cur.maxCpu = Math.max(cur.maxCpu, t.cpu);
    byGroup.set(key, cur);
  }
  const bulk = [...byGroup.entries()].find(([, v]) => v.count >= 3);
  if (bulk && level === 'ok') {
    level = 'warning';
  }
  if (bulk) {
    symptom += `；线程组「${bulk[0]}*」共 ${bulk[1].count} 个线程 CPU≥20%（任务堆积方向）`;
  }

  // 根因候选（启发式）
  const rootCause = !ioWait && bizFrame ? `${bizFrame}（首个业务帧）` : undefined;
  const impact =
    topCpu !== null && topCpu >= 50 && WORKER_NAME_RE.test(top.name)
      ? '该线程为业务工作线程，高占用将直接拖慢其承载的请求处理'
      : undefined;

  // 证据表格（行级动作：送终端）
  const rows: FindingRow[] = sorted.slice(0, 10).map(t => {
    const biz = firstBizFrame(t.stack);
    const parts = biz ? splitFrame(biz) : null;
    return {
      cells: [t.name, t.id, t.state, t.cpu !== null ? `${t.cpu}%` : '-', biz ?? (t.stack[0] ?? '-')],
      highlight: t.cpu !== null && t.cpu >= 50,
      note: isIdleWorker(t) ? '空闲 worker' : isIoWait(t) ? 'IO 等待' : undefined,
      actions: parts ? [{ label: '看调用栈', command: `stack ${parts.cls} ${parts.method} -n 3` }] : undefined
    };
  });

  // 下一步动作（联动终端）
  const nextActions: FindingAction[] = [];
  if (frameParts) {
    nextActions.push({
      label: `trace ${frameParts.method} 耗时`,
      command: `trace ${frameParts.cls} ${frameParts.method} '#cost > 100' -n 5`
    });
    nextActions.push({ label: `stack ${frameParts.method}`, command: `stack ${frameParts.cls} ${frameParts.method} -n 3` });
  }
  if (ioWait) {
    nextActions.push({ label: '查看数据库连接等待', command: 'thread --state WAITING' });
  } else if (level === 'critical' || level === 'warning') {
    nextActions.push({ label: '火焰图采样30s', command: 'profiler start --duration 30' });
  }

  return { kind: 'finding', level, symptom, rootCause, impact, sections: [
    { title: '线程 Top', type: 'table', columns: ['线程', 'Id', '状态', 'CPU', '热点帧'], rows }
  ], nextActions };
}

/** 死锁/阻塞检测（thread --blocked-thread-locks）：判定型 */
function ruleDeadlock(ctx: FindingContext): DiagnosticFinding {
  const { threads } = parseThreads(ctx.output);
  const text = ctx.output;

  if (threads.length === 0) {
    const clean = /no\s+(?:most\s+)?block\w*|no\s+deadlock|没有.*阻塞/i.test(text);
    return {
      kind: 'finding',
      level: 'ok',
      symptom: clean ? '未发现阻塞线程 / 死锁' : '未解析到线程数据，请查看原始输出',
      sections: []
    };
  }

  const blocked = threads.filter(t => t.state === 'BLOCKED' || t.waitingLock);
  const holders = new Set(blocked.map(t => t.lockedBy).filter(Boolean) as string[]);
  // 环检测：持有者自身也在 blocked 名单中 → 确认死锁
  const deadlock = blocked.some(t => holders.has(t.name));
  const holderThread = blocked.find(t => t.lockedBy)?.lockedBy;
  const holderBiz = blocked.find(t => firstBizFrame(t.stack));

  const level = deadlock ? 'critical' : blocked.length > 0 ? 'warning' : 'ok';
  const symptom =
    blocked.length > 0
      ? `${blocked.length} 个线程处于 BLOCKED${deadlock ? '，且等待关系成环——确认死锁' : '（未成环，可能是长持锁而非死锁）'}`
      : '未发现 BLOCKED 线程';
  const rootCause = holderThread
    ? `锁持有者线程「${holderThread}」${holderBiz ? `，热点 ${firstBizFrame(holderBiz.stack)}` : ''}`
    : undefined;
  const impact = blocked.length > 0 ? `被阻塞线程的请求将挂起直至锁释放${deadlock ? '（死锁不会自愈，需重启或干预）' : ''}` : undefined;

  const rows: FindingRow[] = blocked.map(t => {
    const biz = firstBizFrame(t.stack);
    const parts = biz ? splitFrame(biz) : null;
    return {
      cells: [t.name, t.id, t.waitingLock ?? '-', t.lockedBy ?? '-', biz ?? '-'],
      highlight: true,
      actions: parts ? [{ label: '看持有者调用栈', command: `stack ${parts.cls} ${parts.method} -n 3` }] : undefined
    };
  });

  return {
    kind: 'finding',
    level,
    symptom,
    rootCause,
    impact,
    sections: rows.length
      ? [{ title: '阻塞线程', type: 'table', columns: ['线程', 'Id', '等待锁', '锁持有者', '热点帧'], rows }]
      : [],
    nextActions: rows.length ? [{ label: '线程全景', command: 'thread' }] : undefined
  };
}

/** 线程全景（thread）：信息型 describe + 轻判定 */
function ruleThreadAll(ctx: FindingContext): DiagnosticFinding {
  const { header, threads } = parseThreads(ctx.output);
  if (threads.length === 0) return parseFailed();

  const countBy = (s: string) => threads.filter(t => t.state === s).length;
  const blockedCount = countBy('BLOCKED');
  const level = blockedCount > 0 ? 'warning' : 'ok';

  const metrics = [
    { label: '总数', value: String(threads.length) },
    { label: 'RUNNABLE', value: String(countBy('RUNNABLE')) },
    { label: 'BLOCKED', value: String(blockedCount), level: blockedCount > 0 ? 'warning' as const : undefined },
    { label: 'WAITING', value: String(countBy('WAITING')) },
    { label: 'TIMED_WAITING', value: String(countBy('TIMED_WAITING')) }
  ];

  const rows: FindingRow[] = threads.map(t => {
    const biz = firstBizFrame(t.stack);
    const parts = biz ? splitFrame(biz) : null;
    return {
      cells: [t.name, t.state, t.cpu !== null ? `${t.cpu}%` : '-', biz ?? (t.stack[0] ?? '-')],
      highlight: t.state === 'BLOCKED',
      note: isIdleWorker(t) ? '空闲 worker' : isIoWait(t) ? 'IO 等待' : undefined,
      actions: parts ? [{ label: 'stack', command: `stack ${parts.cls} ${parts.method} -n 3` }] : undefined
    };
  });

  return {
    kind: 'describe',
    level,
    symptom: blockedCount > 0 ? `${blockedCount} 个线程 BLOCKED，建议用「死锁/阻塞检测」跟进` : `共 ${threads.length} 个线程`,
    sections: [
      { title: header ?? '线程状态分布', type: 'metrics', metrics },
      { title: '线程明细', type: 'table', columns: ['线程', '状态', 'CPU', '热点帧'], rows }
    ],
    nextActions: blockedCount > 0 ? [{ label: '阻塞检测', command: 'thread --blocked-thread-locks' }] : undefined
  };
}

/** 内存分区（memory）：信息 + 水位判定 */
function ruleMemory(ctx: FindingContext): DiagnosticFinding {
  const regions = parseMemory(ctx.output);
  if (regions.length === 0) return parseFailed();

  const find = (...names: string[]) => regions.find(r => names.some(n => r.name.toLowerCase().includes(n)));
  const old = find('old_gen', 'tenured_gen', 'ps_old', 'g1_old');
  const metaspace = find('metaspace');
  const heap = regions.find(r => r.name.toLowerCase() === 'heap');

  // 分区差异化判定（eden/survivor 高是常态，不判定避免误报）
  let level: DiagnosticFinding['level'] = 'ok';
  const notes: string[] = [];
  if (old?.percent !== null && old?.percent !== undefined) {
    if (old.percent >= 85) {
      level = 'critical';
      notes.push(`Old 区 ${old.percent}%（OOM 前兆 / Full GC 频发风险）`);
    } else if (old.percent >= 70) {
      level = 'warning';
      notes.push(`Old 区 ${old.percent}% 偏高`);
    }
  }
  if (metaspace?.percent !== null && metaspace?.percent !== undefined && metaspace.percent >= 90) {
    level = level === 'ok' ? 'warning' : level;
    notes.push(`Metaspace ${metaspace.percent}%（类加载泄漏线索，可用「类加载检索」跟进）`);
  }

  const metricOf = (label: string, r: typeof regions[number]) => ({
    label,
    value: `${r.used} / ${r.total}${r.percent !== null ? `（${r.percent}%）` : ''}`,
    percent: r.percent ?? undefined,
    level:
      r.percent === null
        ? undefined
        : r.percent >= 85
          ? ('critical' as const)
          : r.percent >= 70
            ? ('warning' as const)
            : ('ok' as const)
  });

  const metrics = [
    ...(heap ? [metricOf('堆整体', heap)] : []),
    ...(old ? [metricOf('Old 区', old)] : []),
    ...(metaspace ? [metricOf('Metaspace', metaspace)] : []),
    ...regions
      .filter(r => r !== heap && r !== old && r !== metaspace)
      .map(r => metricOf(r.name, r))
  ];

  return {
    kind: 'describe',
    level,
    symptom: notes.length ? notes.join('；') : `内存水位正常（Old ${old?.percent ?? '-'}%）`,
    sections: [{ title: '内存分区水位', type: 'metrics', metrics }],
    nextActions:
      level !== 'ok'
        ? [
            { label: 'JVM 全景（GC 次数）', command: 'jvm' },
            { label: '实时面板', command: 'dashboard' }
          ]
        : undefined
  };
}

/** JVM 全景（jvm）：信息型 describe 分组 */
function ruleJvm(ctx: FindingContext): DiagnosticFinding {
  const groups = parseJvm(ctx.output);
  if (groups.length === 0) return parseFailed();

  // 轻判定：THREAD 组中 Blocked 计数
  let blocked = 0;
  groups
    .find(g => g.title === 'THREAD')
    ?.kv.forEach(item => {
      if (/blocked/i.test(item.key)) blocked = Number(item.value) || 0;
    });

  return {
    kind: 'describe',
    level: blocked > 0 ? 'warning' : 'ok',
    symptom: blocked > 0 ? `BLOCKED 线程 ${blocked} 个，建议用「死锁/阻塞检测」跟进` : 'JVM 运行信息',
    sections: groups.map(g => ({ title: g.title, type: 'kv' as const, kv: g.kv })),
    nextActions: blocked > 0 ? [{ label: '阻塞检测', command: 'thread --blocked-thread-locks' }] : undefined
  };
}

/** 类加载检索（sc -d）：信息型 */
function ruleSc(ctx: FindingContext): DiagnosticFinding {
  const blocks = parseSc(ctx.output);
  if (blocks.length === 0) return parseFailed();

  return {
    kind: 'describe',
    level: 'info',
    symptom: blocks.length === 1 ? `已加载：${blocks[0].className ?? '未知类'}` : `匹配到 ${blocks.length} 个类`,
    sections: blocks.map(b => ({
      title: b.className ?? '类信息',
      type: 'kv' as const,
      kv: b.kv
    })),
    nextActions: blocks[0].className
      ? [{ label: '列出该类方法', command: `sm ${blocks[0].className}` }]
      : undefined
  };
}

/** 堆转储（heapdump）：导出型产物卡片 */
function ruleHeapdump(ctx: FindingContext): DiagnosticFinding {
  const file = parseHeapdumpFile(ctx.output);
  const failed = /error|fail|exception/i.test(ctx.output) && !file;
  return {
    kind: 'export',
    level: failed ? 'critical' : 'ok',
    symptom: failed ? '堆转储失败，请查看原始输出' : file ? `堆转储已生成：${file}` : '命令已执行，未见输出中的文件路径',
    sections: [
      {
        title: '后续分析建议',
        type: 'kv',
        kv: [
          { key: '离线分析', value: '下载 hprof 文件，用 Eclipse MAT 打开，查看 Leak Suspects 报告' },
          { key: '体积提示', value: 'live 转储通常接近存活对象总量，注意目标节点磁盘水位' },
          { key: '副作用', value: '执行时已触发一次 Full GC（STW），观察服务是否有抖动' }
        ]
      }
    ]
  };
}

/** 调整日志级别（logger）：变更型操作回执 */
function ruleLogger(ctx: FindingContext): DiagnosticFinding {
  const { name, level } = parseLoggerCommand(ctx.command);
  const success = /success|成功/i.test(ctx.output);
  const restoreLevel = level.toUpperCase() === 'DEBUG' ? 'INFO' : 'INFO';
  return {
    kind: 'mutate',
    level: success ? 'ok' : 'critical',
    symptom: success ? `已将 ${name} 日志级别调整为 ${level}` : '调整失败，请查看原始输出',
    sections: [
      {
        title: '操作回执',
        type: 'kv',
        kv: [
          { key: '变更对象', value: name },
          { key: '目标级别', value: level || '-' },
          { key: '生效范围', value: '运行时即时生效，重启后失效（回到启动配置）' }
        ]
      }
    ],
    nextActions: [{ label: `还原为 ${restoreLevel}`, command: `logger --name ${name} --level ${restoreLevel}` }]
  };
}

// ========== 入口 ==========

function parseFailed(): DiagnosticFinding {
  return { kind: 'describe', level: 'unknown', sections: [], parseFailed: true };
}

const RULES: Record<string, (ctx: FindingContext) => DiagnosticFinding> = {
  'cpu-top': ruleCpuTop,
  deadlock: ruleDeadlock,
  'thread-all': ruleThreadAll,
  memory: ruleMemory,
  jvm: ruleJvm,
  sc: ruleSc,
  heapdump: ruleHeapdump,
  logger: ruleLogger
};

/**
 * 构建诊断 Finding：按场景 id 分发到对应规则。
 * 未知场景（含后续新增未适配的）降级为 parseFailed，前端仅展示原始输出。
 */
export function buildFinding(ctx: FindingContext): DiagnosticFinding {
  const rule = RULES[ctx.scenarioId];
  if (!rule) return parseFailed();
  try {
    return rule(ctx);
  } catch {
    // 解析/判定异常一律降级，绝不阻断结果展示
    return parseFailed();
  }
}
