/**
 * OneOps API 安全审计工作流
 *
 * 功能：审计所有后端 API 路由的安全配置
 * 输入：无（或通过 args 传递特定路径）
 * 输出：安全审计报告
 */

// Workflow 版本
const WORKFLOW_VERSION = "1.0.0";

// 主要执行函数
async function run({ spawn, args }) {
  console.log("🔒 开始 OneOps API 安全审计...");

  // ========== 阶段 1: 路由扫描 ==========
  console.log("\n📋 阶段 1: 扫描路由定义文件");

  const scanAgent = await spawn({
    task: "扫描后端路由文件",
    prompt: `
      扫描以下文件中的所有 API 路由定义：
      1. backend/routes/routes.go
      2. backend/routes/routes_cmdb_section.go

      对于每个路由，提取：
      - HTTP 方法（GET/POST/PUT/DELETE）
      - 路由路径（如 /api/system/users）
      - 使用的中间件（特别注意 Auth()）
      - 处理函数名称

      返回结构化的路由列表，格式如下：
      [
        {
          "method": "POST",
          "path": "/api/system/users",
          "middleware": ["middlewares.Auth()"],
          "handler": "userController.CreateUser",
          "file": "routes.go",
          "line": 74
        },
        ...
      ]
    `
  });

  const routes = await scanAgent.result();
  console.log(`✅ 扫描完成，发现 ${routes.length} 个路由`);

  // ========== 阶段 2: 安全分析 ==========
  console.log("\n🔍 阶段 2: 分析安全配置");

  const analysisAgent = await spawn({
    task: "分析路由安全配置",
    prompt: `
      基于以下路由列表，分析安全配置：

      ${JSON.stringify(routes, null, 2)}

      分析规则：
      1. 认证检查：路由是否使用 middlewares.Auth() 中间件
         - 公开路由（如 /login）可以无认证
         - 其他路由应该有认证

      2. 敏感操作检查：
         - POST/PUT/DELETE 操作属于敏感操作
         - 必须有认证保护

      3. WebSocket 特殊处理：
         - WebSocket 连接可能使用自定义认证
         - 注明 "handler 自行验证 token"

      返回分析结果：
      {
        "total": 总路由数,
        "protected": 受保护的路由数,
        "unprotected": 未保护的路由数,
        "issues": [
          {
            "severity": "CRITICAL|HIGH|MEDIUM|LOW",
            "route": "POST /api/system/users",
            "issue": "缺少认证中间件",
            "recommendation": "添加 middlewares.Auth()"
          }
        ]
      }
    `
  });

  const analysis = await analysisAgent.result();
  console.log(`✅ 分析完成，发现 ${analysis.issues.length} 个问题`);

  // ========== 阶段 3: 交叉验证 ==========
  console.log("\n🔄 阶段 3: 交叉验证（检查控制器实现）");

  // 选择高优先级问题进行深度验证
  const criticalIssues = analysis.issues.filter(i => i.severity === "CRITICAL" || i.severity === "HIGH");

  let verificationResults = [];
  if (criticalIssues.length > 0) {
    const verifyAgent = await spawn({
      task: "验证高优先级问题",
      prompt: `
        对于以下高优先级安全问题，检查控制器代码中是否有额外的安全检查：

        ${JSON.stringify(criticalIssues, null, 2)}

        检查对应的控制器文件（如 backend/controllers/user.go）：
        1. 控制器方法内部是否有用户身份检查
        2. 是否有权限检查（如 CheckPermission）
        3. 是否记录审计日志

        如果控制器内部有适当的安全检查，可以降低问题优先级。

        返回更新后的问题列表，标注已验证的问题。
      `
    });

    verificationResults = await verifyAgent.result();
  }

  // ========== 阶段 4: 生成报告 ==========
  console.log("\n📊 阶段 4: 生成审计报告");

  const reportAgent = await spawn({
    task: "生成安全审计报告",
    prompt: `
      基于以下分析结果，生成专业的安全审计报告：

      === 路由统计 ===
      ${JSON.stringify(analysis, null, 2)}

      === 验证结果 ===
      ${JSON.stringify(verificationResults, null, 2)}

      报告格式：
      1. 执行摘要（关键发现）
      2. 详细发现（按严重程度分组）
      3. 修复建议（优先级排序）
      4. 合规性评估

      使用 Markdown 格式，包含表格和代码示例。
    `
  });

  const finalReport = await reportAgent.result();

  // ========== 输出结果 ==========
  console.log("\n" + "=".repeat(60));
  console.log("🎯 API 安全审计报告");
  console.log("=".repeat(60));
  console.log(finalReport);
  console.log("=".repeat(60));

  return {
    success: true,
    summary: {
      totalRoutes: analysis.total,
      protected: analysis.protected,
      unprotected: analysis.unprotected,
      criticalIssues: analysis.issues.filter(i => i.severity === "CRITICAL").length
    },
    report: finalReport
  };
}

// 导出供 Workflow 运行时使用
module.exports = { run };
