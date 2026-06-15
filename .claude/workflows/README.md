# OneOps Workflow 使用说明

## 已安装的工作流

### security-audit（API 安全审计）

**功能**：审计后端所有 API 路由的安全配置

**运行方式**：
```bash
# 在 Claude Code 中
/security-audit

# 或通过提示词触发
ultracode: 审计所有 API 路由的安全配置
```

**执行阶段**：
1. **路由扫描**：解析 routes.go 中的所有路由定义
2. **安全分析**：检查认证中间件和敏感操作保护
3. **交叉验证**：验证控制器层面的安全检查
4. **生成报告**：输出 Markdown 格式的审计报告

**输出**：
- 安全统计（总数/受保护/未保护）
- 问题列表（按严重程度分组）
- 修复建议（优先级排序）

---

## Workflow 脚本结构

每个 Workflow 脚本需要导出 `run` 函数：

```javascript
async function run({ spawn, args }) {
  // spawn: 创建子 Agent
  // args: 用户传入的参数

  // 阶段 1
  const agent1 = await spawn({ task: "...", prompt: "..." });
  const result1 = await agent1.result();

  // 阶段 2（可以使用阶段 1 的结果）
  const agent2 = await spawn({
    task: "...",
    prompt: `基于 ${result1} 进行分析...`
  });
  const result2 = await agent2.result();

  // 返回最终结果
  return { summary: "...", data: result2 };
}

module.exports = { run };
```

---

## 核心概念

### 1. 并行执行

Workflow 可以并行运行多个 Agent：

```javascript
const [agent1, agent2, agent3] = await Promise.all([
  spawn({ task: "分析路由 A", prompt: "..." }),
  spawn({ task: "分析路由 B", prompt: "..." }),
  spawn({ task: "分析路由 C", prompt: "..." })
]);

const [result1, result2, result3] = await Promise.all([
  agent1.result(),
  agent2.result(),
  agent3.result()
]);
```

### 2. 串行依赖

某些阶段需要等待前置阶段完成：

```javascript
// 先扫描路由
const routes = await (await spawn({ task: "扫描路由" })).result();

// 基于扫描结果进行分析
const analysis = await (await spawn({
  task: "分析安全",
  prompt: `分析这些路由: ${routes}`
})).result();
```

### 3. 对抗性验证

可以让两个 Agent 互相验证：

```javascript
// Agent A 分析
const analysisA = await (await spawn({
  task: "安全分析",
  prompt: "找出所有安全问题"
})).result();

// Agent B 验证
const verification = await (await spawn({
  task: "交叉验证",
  prompt: `验证以下发现是否准确：${analysisA}`
})).result();
```

---

## Workflow vs Agent Team

| 特性 | Agent Team | Workflow |
|------|------------|----------|
| 协调方式 | Agent 逐轮决定 | 脚本预先定义 |
| 中间结果 | 共享任务列表 | 脚本变量 |
| 规模 | 少数长期运行的 Agent | 数十个短期 Agent |
| 可重复性 | 依赖 Agent 判断 | 编排脚本本身 |

---

## 创建新的 Workflow

1. 在 `.claude/workflows/` 或 `~/.claude/workflows/` 创建 `.js` 文件
2. 实现 `run` 函数
3. 运行 `/workflows` 查看列表
4. 按 `s` 保存为命令

---

## 成本控制

- **估算成本**：先在小范围测试（单个目录而非整个代码库）
- **监控进度**：`/workflows` 查看实时 token 使用情况
- **随时停止**：按 `x` 终止运行，已完成的工作不会丢失
- **Agent 上限**：最多 16 个并发，总计 1000 个 Agent

---

## 故障排查

### Workflow 不触发
- 检查是否安装了正确版本的 Claude Code（≥2.1.160）
- 检查 Dynamic workflows 是否在 `/config` 中启用
- 尝试使用明确的 `ultracode:` 关键词

### 运行中断
- 使用 `/workflows` 查看状态
- 按 `p` 暂停/恢复
- 同一会话内可以恢复，退出会话后需重新运行

### 权限问题
- Workflow 运行时继承工具白名单
- 将需要的命令（如 grep、find）添加到允许列表
- 使用 `Auto` 权限模式避免频繁确认
