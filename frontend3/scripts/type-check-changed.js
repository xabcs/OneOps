#!/usr/bin/env node
/**
 * 增量类型检查
 *
 * 优化点（对比全量 `vue-tsc --noEmit`）：
 * 1. git 层短路：没有变更的 .ts/.tsx/.vue 文件时直接跳过，秒级返回
 * 2. tsbuildinfo 增量：开启 --incremental，二次检查只分析受影响文件
 * 3. 缓存自愈：失败时自动清空增量缓存全量重试一次，避免脏缓存误报
 *
 * 用法：
 *   node scripts/type-check-changed.js           # 检查工作区相对 HEAD 的变更
 *   node scripts/type-check-changed.js --all     # 强制全量检查
 */
import { execSync, spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, rmSync } from 'node:fs';
import { resolve } from 'node:path';

const FORCE_ALL = process.argv.includes('--all');

const colors = process.stdout.isTTY
  ? {
      reset: '\x1B[0m',
      bright: '\x1B[1m',
      primary: '\x1B[38;5;111m',
      success: '\x1B[38;5;46m',
      error: '\x1B[38;5;203m',
      warning: '\x1B[38;5;220m'
    }
  : {
      reset: '',
      bright: '',
      primary: '',
      success: '',
      error: '',
      warning: ''
    };

const log = (message, color = '') => console.log(color + message + colors.reset);
const sep = (process.stdout.isTTY ? '━' : '-').repeat(55);

const CACHE_DIR = resolve('node_modules/.cache/vue-tsc');
const TS_BUILD_INFO_FILE = resolve(CACHE_DIR, 'tsconfig.tsbuildinfo');

/** 收集相对 HEAD 变更（含暂存、未暂存、未跟踪）的类型相关文件；非 git 环境返回 null */
function getChangedFiles() {
  try {
    // git diff 输出相对仓库根（带 frontend3/ 前缀），ls-files 输出相对当前目录，统一去除前缀
    const prefix = execSync('git rev-parse --show-prefix', { encoding: 'utf-8' }).trim();
    const normalize = line => (prefix && line.startsWith(prefix) ? line.slice(prefix.length) : line);

    const tracked = execSync('git diff --name-only HEAD --diff-filter=ACMR', { encoding: 'utf-8' });
    const untracked = execSync('git ls-files --others --exclude-standard', { encoding: 'utf-8' });
    return `${tracked}\n${untracked}`
      .split('\n')
      .map(line => normalize(line.trim()))
      .filter(line => /\.(ts|tsx|vue)$/.test(line))
      .filter(line => !line.includes('node_modules'))
      .filter(line => /^(src|packages|build)\//.test(line));
  } catch {
    return null;
  }
}

function runTypeCheck() {
  mkdirSync(CACHE_DIR, { recursive: true });

  const result = spawnSync(
    'pnpm',
    ['exec', 'vue-tsc', '--noEmit', '--skipLibCheck', '--incremental', '--tsBuildInfoFile', TS_BUILD_INFO_FILE],
    { stdio: 'inherit', shell: process.platform === 'win32' }
  );

  return result.status === 0;
}

function summary(startTime, success) {
  const duration = ((Date.now() - startTime) / 1000).toFixed(2);
  log('');
  log(sep, colors.bright + colors.primary);
  if (success) {
    log(`✅ 类型检查通过！(耗时: ${duration}s)`, colors.bright + colors.success);
  } else {
    log(`❌ 类型检查失败！(耗时: ${duration}s)`, colors.bright + colors.error);
    log('请修复上述类型错误后重试。', colors.warning);
  }
  log(sep, colors.bright + colors.primary);
  log('');
}

function main() {
  const startTime = Date.now();
  const cacheExisted = existsSync(TS_BUILD_INFO_FILE);

  log('');
  log(sep, colors.bright + colors.primary);
  log('🔍 TypeScript 增量类型检查', colors.bright + colors.primary);
  log(sep, colors.bright + colors.primary);
  log('');

  if (!FORCE_ALL) {
    const changed = getChangedFiles();

    if (changed !== null && changed.length === 0) {
      log('✅ 没有检测到变更的类型相关文件，跳过类型检查', colors.bright + colors.success);
      process.exit(0);
    }

    if (changed && changed.length > 0) {
      log(`📦 检测到 ${changed.length} 个变更文件：`, colors.bright + colors.warning);
      changed.slice(0, 10).forEach(file => log(`   ${file}`));
      if (changed.length > 10) log(`   ... 及其余 ${changed.length - 10} 个文件`);
      log('');
    }
  }

  let success = runTypeCheck();

  // 失败且此前存在增量缓存：可能是脏缓存（如切换分支），清空后全量重试一次
  if (!success && cacheExisted) {
    rmSync(CACHE_DIR, { recursive: true, force: true });
    log('♻️  增量缓存可能已过期，已清空缓存并全量重试...', colors.warning);
    log('');
    success = runTypeCheck();
  }

  summary(startTime, success);
  process.exit(success ? 0 : 1);
}

main();
