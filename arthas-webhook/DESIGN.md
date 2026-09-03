# msre-pilot 架构设计

> 代码目录：`arthas-webhook/`（历史命名，服务名 msre-pilot）
> 部署清单：`msrepilot-deploy.yaml` ｜ 部署手册：`deploy/1.txt` 第 13 条

## 1. 服务定位

**OneOps 管理面在 K8s 集群内的自治代理**：平台只表达意图（ConfigMap），
msre-pilot 负责把意图落成集群实际状态（MWC / 证书 / 注入），并以准入回调
服务身份参与 Pod 生命周期。

| 维度 | 定位 |
|---|---|
| 角色 | 意图执行器（管理面的"手"）+ 准入回调（数据面的"钩子"） |
| 形态 | 2 个 controller + 1 个 admission server，单副本 Deployment |
| 术语 | 是 K8s controller（watch 原生资源驱动收敛）；不是 operator（无 CRD） |
| 结构参照 | 架构模式学 ingress-nginx（controller+数据面合体）；注入子系统同构 istiod sidecar injector |
| 不做 | 业务流量转发；平台 UI/API（backend3 职责）；依赖平台网络可达（断联自治） |

命名的意图：MSRE（SRE 场景）+ Pilot（领航）——未来 MSRE 域其他集群内
自治职责（探针注入、故障演练等）的挂载点，Arthas 注入是第一个租户。

## 2. 演进史（为什么长成这样）

| 阶段 | 形态 | 遇到的问题 |
|---|---|---|
| v0 | webhook 内嵌 backend3，URL 型 MWC | ACK master 出网不通，apiserver 回调平台超时 |
| v1 | 独立工程部署集群内，Service 型 MWC | 改名遗留孤儿 MWC（Fail 策略）拦截全集群 Pod 创建 |
| v2 | MWC 生命周期下放 pilot（本设计） | 平台持有 cluster-scoped 写权限过重；启停是 edge-triggered |
| v3 | 意图 watch 快路径 | v2 收敛依赖 15s ticker，启用接口同步等待最长 20s |

核心教训：**MWC 名硬编码 + 按名删除 + apiserver 不校验后端存活 = 孤儿准入规则
是生产事故级隐患**。v2 的所有权 label + 控制器化是对它的系统性回答。

## 3. 架构总览

```
┌─────────────── OneOps 管理面（平台侧）───────────────┐
│  治理页 ──> backend3 (Enable/Disable/GetStatus)      │
│              │ 只写意图（ns 级 ConfigMap 权限）        │
└──────────────┼───────────────────────────────────────┘
               ▼
   ConfigMap msre-pilot-config { WEBHOOK_ENABLED: true/false,
                                 ARTHAS_TUNNEL_WS, ARTHAS_INIT_IMAGE }
               │ watch（快路径，亚秒）
               ▼
┌─────────────── msre-pilot（集群内 test ns，自治）─────┐
│ ① MWC controller (reconcile.go)                      │
│    watch 意图 + 15s ticker 兜底 → 建/删/修 MWC        │
│    孤儿清理（所有权 label + 历史名收编）               │
│ ② 证书 controller (rotator.go)                       │
│    ECDSA 自签 CA/leaf → Secret；12h 巡检，<30 天轮转  │
│ ③ admission server (inject.go + main.go, TLS :9443)  │
│    /webhook/msre-pilot ← apiserver 准入回调           │
│    /healthz                                          │
└──────┬───────────────────────────────────────────────┘
       │ ①确保存在+caBundle正确            ③回调时注入
       ▼                                   ▼
  MutatingWebhookConfiguration         Patch：initContainer
  msre-pilot-inject                    oneops-arthas-agent
  (Service 型→ClusterIP:9443)          + label 触发条件
                                           │
                                           ▼
                                    Arthas agent → tunnel-server
```

### 三个角色

| 角色 | 文件 | 触发 | 动作 |
|---|---|---|---|
| MWC controller | reconcile.go | 意图 watch + 15s ticker | 建/删/修 MWC、孤儿清理 |
| 证书 controller | rotator.go | 启动 + 12h 巡检 | 自签/轮转 Secret、跟随更新 caBundle |
| admission server | inject.go / main.go | apiserver 回调 | Patch initContainer 注入 agent |

## 4. 核心设计决策

### 4.1 意图与实现分离（本设计的根基）

治理页启停只写 `ConfigMap: WEBHOOK_ENABLED`，MWC 实际状态由集群内控制器收敛。

- **收益 1**：平台不再需要 cluster-scoped MWC 写权限（RBAC 最小化）
- **收益 2**：level-triggered——漂移自愈，重启/断联后自动追平，无状态
- **收益 3**：审计清晰——意图（ConfigMap 变更）与实现（控制器日志）分离

### 4.2 双路径收敛（v3）

```
watch 意图变化 ──秒级──> reconcile（快路径，事件驱动）
15s ticker ───────────> reconcile（慢路径，漂移自愈 + watch 断线补偿）
```

- wake channel 容量 1：事件风暴合并，reconcile 串行无竞态
- backend3 启用接口等待窗口因此从 20s 收窄到 5s（通常 1s 内返回）

### 4.3 所有权 label（按属性，不按名字）

```yaml
oneops.io/managed-by: oneops
oneops.io/component: arthas-webhook
```

- 识别"我创建的"MWC 与名字解耦——服务再改名，旧资源仍可识别回收
- 孤儿清理两来源：label 匹配的非当前名（改名遗留）+ 历史名
  `oneops-arthas-inject`（无 label，一次性收编）

### 4.4 证书全自签（rotator.go）

- ECDSA P256；CA/leaf 各 10 年；SAN 覆盖 Service DNS
- 12h 巡检，剩余 <30 天轮转；resourceVersion 乐观锁防并发
- 轮转后即时更新 MWC caBundle，os.Exit 由 K8s 重启加载新证书
- **不留 openssl 手工证书兼容路径**；非本体系证书（无 ca.key）强制重建接管

### 4.5 Service 型 MWC clientConfig

```yaml
clientConfig.service: { namespace: test, name: msre-pilot, port: 9443, path: /webhook/msre-pilot }
```

- apiserver → ClusterIP 集群内必达，规避 ACK master 出网受限
- URL 型（`https://...`）为已废弃的历史形态，代码仅存于 backend3 兜底分支

### 4.6 准入安全边界

- `failurePolicy: Ignore`——pilot 不可用时放行，不阻塞业务 Pod 创建
- `objectSelector: oneops-arthas-injection=enabled`——仅显式打标的 Pod 进入注入
- `namespaceSelector` 排除 kube-system / kube-public / kube-node-lease
- annotation `oneops-arthas-injection-disabled=true` 可单 Pod 逃生
- timeoutSeconds 5，sideEffects None，admissionReviewVersions v1

### 4.7 刻意不做的

| 不做 | 理由 |
|---|---|
| CRD / operator 化 | 单 key 意图 + 单 MWC 规模下 overkill；ConfigMap 当 spec 够用且零搭建成本（信号：意图膨胀、多实例管理、status 可观察性需求出现时再迁） |
| LeaderElection | 证书/MWC 写为 10 年频级幂等操作，乐观锁 + 抖动 + 单副本足够 |
| 多副本 | 当前 1 副本（准入 Ignore 策略下短暂不可用可接受）；需高可用改 replicas 即可 |

## 5. 关键数据流

### 5.1 启用注入（治理页 → MWC 生效）

```
治理页点"启用"
 → backend3 setWebhookIntent: upsert ConfigMap WEBHOOK_ENABLED=true（≤1s）
 → pilot watch 感知 → reconcile → 创建 MWC msre-pilot-inject（带所有权 label）
 → backend3 轮询 Get MWC 确认（≤5s 兜底，超时仅 Warn 最终一致）
 → 同轮 reconcile 顺带清理孤儿（旧 oneops-arthas-inject）
```

### 5.2 Pod 注入（数据面）

```
业务 Deploy 打 label oneops-arthas-injection=enabled → 创建 Pod
 → apiserver 准入链调用 MWC → ClusterIP:9443/webhook/msre-pilot
 → pilot 校验（系统 ns / 逃生 annotation / label）→ 200 + JSONPatch
 → Pod 带 initContainer oneops-arthas-agent 启动
 → agent 反连 tunnel-server（ws://arthas-tunnel.test:7777/ws）
```

### 5.3 证书轮转

```
12h 巡检发现 leaf 剩余 <30 天
 → 重签 leaf（CA 不动）→ 写回 Secret（乐观锁）
 → 更新 MWC caBundle → os.Exit(0)
 → K8s 重启新 Pod → kubelet 同步挂载卷 → 新证书生效
```

## 6. RBAC 边界

| 权限 | 范围 | 用途 |
|---|---|---|
| Role `msre-pilot-cert` | ns 级 secrets get/update（resourceNames 限定 msre-pilot-tls） | 证书读写 |
| 同上 | ns 级 configmaps get/list/watch | 意图读取 + informer。注：RBAC resourceNames 不适用于 list/watch（请求不带资源名），故放开 ns 级只读，数据流由 fieldSelector metadata.name 限定 |
| ClusterRole `msre-pilot-cabundle` | 集群级 mutatingwebhookconfigurations get/list/create/update/delete | MWC 生命周期（cluster-scoped 资源无法降为 ns 级）。范围自律：所有权 label + 历史名，代码层保证只动自己的 |

平台侧（backend3）仅持有目标 ns 的 ConfigMap 读写——**无任何 cluster-scoped 权限**。
agent 不调 K8s API，无需任何 RBAC。

## 7. 故障模式与自愈矩阵

| 故障/漂移 | 系统行为 | 恢复路径 |
|---|---|---|
| MWC caBundle 被改 | ticker 15s 内修正 | reconcile 漂移自愈 |
| MWC Service 指向/选择器被改 | 同上 | 同上 |
| 孤儿 MWC（改名/手工创建） | watch/ticker 触发即删 | cleanupOrphanMWCs |
| pilot Pod 挂 | 新 Pod 创建不注入（Ignore 放行），存量已注入不受影响 | K8s 拉起，启动即一轮 reconcile 追平 |
| watch 断线 | client-go 自动重连 | 重连间隙由 ticker 兜底 |
| 平台失联 | 意图不变，pilot 持续维持现状 | 断联自治（不依赖平台） |
| 证书将过期 | 12h 巡检内轮转 | 4.4 流程 |
| ConfigMap 被删 | 意图缺失 = false | reconcile 删 MWC（禁用语义） |

## 8. 观测与运维

```sh
# 控制器快路径是否生效（启用后 1s 内出现）
kubectl logs -n test deploy/msre-pilot | grep "意图事件触发快路径收敛"

# MWC 状态（应只存在 msre-pilot-inject，无 oneops-arthas-inject 残留）
kubectl get mutatingwebhookconfigurations | grep -E 'msre-pilot|oneops-arthas'

# 意图
kubectl get cm -n test msre-pilot-config -o jsonpath='{.data.WEBHOOK_ENABLED}'

# 证书有效期
kubectl get secret -n test msre-pilot-tls -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
```

关键日志：`MWC 已创建（意图 enabled）` / `MWC 漂移已修正` / `已清理孤儿 MWC` /
`等待 msre-pilot 控制器收敛超时`（backend3，仅 Warn）。

## 9. 演进路线

| 阶段 | 触发信号 | 动作 |
|---|---|---|
| 现状 | — | ConfigMap 单 key 意图 + 单 MWC |
| CRD 化 | 意图膨胀（按 ns/团队/应用分组的注入策略）| 迁移到 controller-runtime，`ArthasInjection` 一应用一对象；现有 reconcile 逻辑可近平移 |
| 多租户 | MSRE 其他职责入驻（探针/演练） | pilot 作为壳，按 component label 区分资源族 |
| OLM | 需要产品化分发 | 打包 operator bundle |

## 10. 文件清单

| 文件 | 职责 |
|---|---|
| main.go | 入口：env 配置、证书保障、gin TLS server、挂载 reconciler |
| inject.go | 注入核心：触发判定 + JSONPatch 构建 + admission handler（与 backend3 BuildInjectionPatch 双份维护） |
| rotator.go | 证书 controller：自签/轮转/caBundle 跟随 |
| reconcile.go | MWC controller：意图 watch、收敛、漂移自愈、孤儿清理 |
| cert.go | 证书就绪校验（verifyCert/coversHost/waitCertReady） |
| Dockerfile | 多阶段构建（golang:1.26-alpine → alpine:3.20 非 root），产物 msre-pilot |
| msrepilot-deploy.yaml | 全套清单：SA/Role/ClusterRole/ConfigMap/Secret/Deployment/Service |
