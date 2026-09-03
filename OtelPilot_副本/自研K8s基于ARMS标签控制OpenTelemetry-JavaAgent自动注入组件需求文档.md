# 自研K8s基于ARMS标签控制OpenTelemetry\-JavaAgent自动注入组件需求文档

## 1\. 需求概述

### 1\.1 项目背景

业务现参考阿里云ACK OnePilot的无侵入探针注入设计思路，为脱离公有云依赖、统一内部可观测体系，需自研一套同架构、同运行逻辑的探针注入组件。**借鉴ARMS OnePilot的实现方式，**自定义专属控制标签，将探针载体替换为开源OpenTelemetry\-JavaAgent，实现私有化、无侵入的Java应用链路追踪能力。

### 1\.2 核心需求（极简核心）

**借鉴ARMS OnePilot无侵入注入架构与运行逻辑，自定义专属控制标签，无需依赖、不兼容ARMS原生体系**，通过自研标签开关控制注入，仅复刻其无侵入Pod注入能力，将探针替换为OpenTelemetry\-JavaAgent，实现K8s集群内Java应用零侵入链路追踪探针自动注入。

### 1\.3 组件核心定位

组件采用标准K8s架构：**Mutating准入Webhook（同步注入）\+ Controller（自愈运维）**

- Webhook：借鉴ARMS OnePilot拦截逻辑，监听Pod创建事件，识别**自研自定义标签**，动态注入OpenTelemetry探针
- Controller：负责组件证书管理、配置同步、异常自愈，保障服务稳定运行

## 2\. 核心工作原理

整体实现无侵入能力，全程不修改业务源码、业务镜像、业务启动命令及原有Deployment配置：

1. 业务Deployment模板配置**自研自定义监控标签**，完全脱离ARMS体系
2. K8s APIServer创建新Pod时，触发组件Webhook拦截
3. Webhook校验Label规则，匹配成功则动态Patch Pod配置
4. 通过Init容器拉取OpenTelemetry\-JavaAgent包，借助共享卷挂载至业务容器
5. 注入环境变量自动加载OpenTelemetry探针，基于字节码增强实现无侵入链路追踪

## 3\. 核心功能需求

### 3\.1 自研自定义标签控制规则（借鉴ARMS设计）

借鉴阿里云ARMS OnePilot的标签开关设计模式、判定逻辑，**不使用、不兼容ARMS原生标签**，自研专属标签体系，独立控制探针注入，标签判定规则如下：

| 自研Label Key             | Label Value   | 组件执行动作                                                |
| ------------------------- | ------------- | ----------------------------------------------------------- |
| otel\.pilot\.auto\.enable | on（字符串）  | 开启注入：自动为Pod注入OpenTelemetry\-JavaAgent探针全套配置 |
| otel\.pilot\.auto\.enable | off/空/无标签 | 关闭注入：直接放行Pod，不做任何修改                         |

支持自研配套扩展标签：可自定义应用名称、JDK版本、探针版本等拓展标签，适配不同业务场景的探针配置。

### 3\.2 OpenTelemetry探针无侵入注入能力

标签匹配开启后，Webhook自动完成Pod动态Patch，无需人工干预：

- 新增emptyDir共享临时卷，用于Init容器与业务容器传输探针包
- 注入轻量Init容器，自动下载指定版本OpenTelemetry\-JavaAgent，并修复文件访问权限，适配非Root容器
- 自动注入`JAVA_TOOL_OPTIONS`环境变量，挂载OpenTelemetry探针启动参数及链路上报配置
- 新增Pod防重注入Annotation，避免Webhook循环修改、重复注入探针

约束：仅新建Pod生效，存量Pod需滚动更新触发注入，该行为逻辑**借鉴ARMS OnePilot**，保持标准无侵入注入特性。

### 3\.3 组件自愈与高可用能力

- Controller自动管理Webhook TLS证书，实现证书自动轮转、自愈
- 支持探针版本、上报地址、授权配置统一同步管理
- Webhook故障策略为`Ignore`，组件异常不阻塞Pod创建，保障业务高可用
- 适配Deployment、StatefulSet、DaemonSet全类型Java工作负载

## 4\. 链路追踪核心能力

依托OpenTelemetry\-JavaAgent字节码增强能力，实现无侵入链路追踪，无需业务改代码：

- 自动适配Spring、Dubbo、MyBatis、Redis、MQ等主流中间件
- 自动生成TraceId/SpanId，遵循W3C Trace协议跨服务透传链路上下文
- 自动补全线程池、异步任务链路，解决异步链路断裂问题
- 采集接口耗时、异常、堆栈信息，支撑分布式链路排查

## 5\. 约束与限制（借鉴ARMS原生行为逻辑）

- 自研监控标签必须配置在`spec.template.metadata.labels`，Deployment顶层标签不生效（借鉴ARMS标签生效逻辑）
- 开关标签值仅识别字符串`on/off`，布尔值、数字等无效
- 业务容器若覆盖`JAVA_TOOL_OPTIONS`环境变量，探针将失效
- 静态Pod、主机级Pod不经过APIServer，无法触发探针注入
- 不可与原生ACK OnePilot同时开启，避免双重Agent加载导致JVM异常

## 6\. 最终需求目标

自研轻量化K8s探针注入组件，**借鉴ARMS OnePilot无侵入注入架构与标签管控设计，完全自研独立体系、不依赖不兼容阿里云ARMS组件**，通过自定义标签开关控制OpenTelemetry\-JavaAgent注入，实现私有化、无侵入、标准化的Java应用分布式链路追踪能力，彻底摆脱阿里云公有云组件依赖。

> （注：部分内容可能由 AI 生成）
