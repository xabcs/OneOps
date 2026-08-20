package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// nativeBindingManagedByLabel OneOps 创建的集群内 Binding 统一标记，便于辨识与 Helm 等区分
const nativeBindingManagedByLabel = "app.kubernetes.io/managed-by"

// GetNativeBindings 获取集群的原生 RBAC 绑定列表（B 模式）
func (s *K8sClusterService) GetNativeBindings(clusterID uint) ([]map[string]interface{}, error) {
	bindings, err := s.clusterRepo.FindNativeBindingsByCluster(clusterID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(bindings))
	for _, b := range bindings {
		item := map[string]interface{}{
			"id":                b.ID,
			"subjectType":       b.SubjectType,
			"roleKind":          b.RoleKind,
			"roleName":          b.RoleName,
			"namespace":         b.Namespace,
			"impersonationName": b.ImpersonationName,
			"k8sBindingName":    b.K8sBindingName,
			"createdAt":         b.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if b.SubjectType == "user" && b.User != nil {
			item["userId"] = b.UserID
			item["subjectName"] = b.User.Username
			item["subjectNickname"] = b.User.Nickname
		}
		if b.SubjectType == "group" && b.Group != nil {
			item["groupId"] = b.GroupID
			item["subjectName"] = b.Group.Name
			item["subjectCode"] = b.Group.Code
		}
		result = append(result, item)
	}
	return result, nil
}

// GetAllNativeBindings 全局原生绑定列表（授权管理页）：clusterID=0 表示全部集群
func (s *K8sClusterService) GetAllNativeBindings(clusterID uint) ([]map[string]interface{}, error) {
	bindings, err := s.clusterRepo.FindAllNativeBindings(clusterID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(bindings))
	for _, b := range bindings {
		item := map[string]interface{}{
			"id":                b.ID,
			"clusterId":         b.ClusterID,
			"subjectType":       b.SubjectType,
			"roleKind":          b.RoleKind,
			"roleName":          b.RoleName,
			"namespace":         b.Namespace,
			"impersonationName": b.ImpersonationName,
			"k8sBindingName":    b.K8sBindingName,
			"createdAt":         b.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if b.Cluster != nil {
			item["clusterName"] = b.Cluster.Name
		}
		if b.SubjectType == "user" && b.User != nil {
			item["userId"] = b.UserID
			item["subjectName"] = b.User.Username
			item["subjectNickname"] = b.User.Nickname
		}
		if b.SubjectType == "group" && b.Group != nil {
			item["groupId"] = b.GroupID
			item["subjectName"] = b.Group.Name
			item["subjectCode"] = b.Group.Code
		}
		result = append(result, item)
	}
	return result, nil
}

// GetClusterOptions 集群候选（授权管理页筛选与表单用，权限归属 k8s.permission.list）
func (s *K8sClusterService) GetClusterOptions() ([]map[string]interface{}, error) {
	return s.clusterRepo.FindClusterOptions()
}

// AssignNativeBinding 创建原生 RBAC 绑定（B 模式授权执行层）：
// 1. 校验主体与角色参数并落库；2. 在目标集群创建真实的 ClusterRoleBinding / RoleBinding。
// 之后该主体在此集群的资源操作将经 impersonation 以真实身份执行，由 kube-apiserver 原生 RBAC 判定。
func (s *K8sClusterService) AssignNativeBinding(operatorID uint, clusterID uint, req *modelk8s.AssignNativeRoleBindingRequest) error {
	if req.SubjectType == "user" {
		if req.UserID == 0 {
			return fmt.Errorf("用户直绑必须指定用户")
		}
	} else {
		if req.GroupID == 0 {
			return fmt.Errorf("组绑定必须指定用户组")
		}
	}

	// 解析命名空间范围：Namespace 单值兼容 + Namespaces 多选，合并去重
	nsSet := make(map[string]struct{})
	if req.Namespace != "" {
		nsSet[req.Namespace] = struct{}{}
	}
	for _, ns := range req.Namespaces {
		if ns != "" {
			nsSet[ns] = struct{}{}
		}
	}
	nsList := make([]string, 0, len(nsSet))
	for ns := range nsSet {
		nsList = append(nsList, ns)
	}
	if req.RoleKind == "Role" && len(nsList) == 0 {
		return fmt.Errorf("命名空间级角色（Role）必须指定命名空间")
	}
	// ClusterRole + ns 列表非空 = 每个命名空间建 RoleBinding 引用该 ClusterRole（ns 级授权）；
	// ClusterRole + ns 列表为空 = 集群级 ClusterRoleBinding
	if req.RoleKind != "Role" && len(nsList) == 0 {
		nsList = []string{""}
	}

	// 计算集群内模拟身份名
	var impersonationName string
	if req.SubjectType == "user" {
		user, err := s.clusterRepo.FindUserByID(req.UserID)
		if err != nil {
			return fmt.Errorf("用户不存在: %w", err)
		}
		impersonationName = "oneops-" + user.Username
	} else {
		group, err := s.clusterRepo.FindUserGroupByID(req.GroupID)
		if err != nil {
			return fmt.Errorf("用户组不存在: %w", err)
		}
		impersonationName = "oneops-group-" + group.Code
	}

	// 验证目标角色在集群内真实存在（共享凭据校验，属管理动作）
	clientset, err := s.GetClient(clusterID)
	if err != nil {
		return fmt.Errorf("获取集群连接失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// OneOps 预设角色（oneops-ops 运维人员等）按需同步到集群，幂等
	if err := s.ensurePresetClusterRoles(ctx, clientset); err != nil {
		return err
	}

	// 验证目标角色在集群内真实存在（共享凭据校验，属管理动作）
	if _, err := clientset.RbacV1().ClusterRoles().Get(ctx, req.RoleName, metav1.GetOptions{}); err == nil {
		// ClusterRole 存在
	} else if req.RoleKind == "Role" {
		if _, nsErr := clientset.RbacV1().Roles(nsList[0]).Get(ctx, req.RoleName, metav1.GetOptions{}); nsErr != nil {
			return fmt.Errorf("目标角色 %s 在命名空间 %s 内不存在: %w", req.RoleName, nsList[0], nsErr)
		}
	} else {
		return fmt.Errorf("目标 ClusterRole %s 在集群内不存在: %w", req.RoleName, err)
	}

	// 按命名空间范围逐条创建（多选 ns 时每 ns 一条记录 + 一个集群内 RoleBinding），失败回滚已建
	created := make([]*modelk8s.K8sNativeRoleBinding, 0, len(nsList))
	for _, ns := range nsList {
		record := &modelk8s.K8sNativeRoleBinding{
			SubjectType:       req.SubjectType,
			UserID:            req.UserID,
			GroupID:           req.GroupID,
			ClusterID:         clusterID,
			RoleKind:          req.RoleKind,
			RoleName:          req.RoleName,
			Namespace:         ns,
			ImpersonationName: impersonationName,
		}
		if err := s.clusterRepo.CreateNativeBinding(record); err != nil {
			s.rollbackNativeBindings(ctx, clientset, created)
			return fmt.Errorf("创建原生绑定记录失败（namespace=%q，若为重复授权请先撤销原绑定）: %w", ns, err)
		}
		if err := s.createNativeK8sBinding(ctx, clientset, record); err != nil {
			_ = s.clusterRepo.DeleteNativeBinding(record.ID)
			s.rollbackNativeBindings(ctx, clientset, created)
			return err
		}
		created = append(created, record)
	}

	logger.Info("创建原生 RBAC 绑定",
		zap.Int("binding_count", len(created)),
		zap.String("subject_type", req.SubjectType),
		zap.String("impersonation", impersonationName),
		zap.Uint("cluster_id", clusterID),
		zap.String("role_kind", req.RoleKind),
		zap.String("role_name", req.RoleName),
		zap.Strings("namespaces", nsList),
		zap.Uint("operator_id", operatorID))
	return nil
}

// rollbackNativeBindings 批量授权中途失败时回滚：删除已创建的集群内 Binding 与 DB 记录
func (s *K8sClusterService) rollbackNativeBindings(ctx context.Context, clientset *kubernetes.Clientset, records []*modelk8s.K8sNativeRoleBinding) {
	for _, r := range records {
		if r.K8sBindingName != "" {
			if r.Namespace == "" {
				if err := clientset.RbacV1().ClusterRoleBindings().Delete(ctx, r.K8sBindingName, metav1.DeleteOptions{}); err != nil {
					logger.Warn("回滚：删除集群内 ClusterRoleBinding 失败", zap.String("binding", r.K8sBindingName), zap.Error(err))
				}
			} else if err := clientset.RbacV1().RoleBindings(r.Namespace).Delete(ctx, r.K8sBindingName, metav1.DeleteOptions{}); err != nil {
				logger.Warn("回滚：删除集群内 RoleBinding 失败", zap.String("binding", r.K8sBindingName), zap.Error(err))
			}
		}
		if err := s.clusterRepo.DeleteNativeBinding(r.ID); err != nil {
			logger.Warn("回滚：删除原生绑定记录失败", zap.Uint("binding_id", r.ID), zap.Error(err))
		}
	}
}

// presetClusterRoleManifests OneOps 预设 ClusterRole 定义（云厂商六档语义的原生实现，全部自建封装）
//
// oneops-admin           管理员 = 全资源读写（cluster-admin 同规则副本，但不绑 system:masters，走普通 RBAC 判定更可控）；
// oneops-readonly        只读管理员 = 聚合角色：内置 view 的全部可见范围 + 基础设施只读，rules 由集群控制器实时聚合，
//
//	自动跟随集群版本升级与带 aggregate-to-view 标签的 CRD 扩展，无静态漂移；
//
// oneops-readonly-infra  聚合源角色：承载 view 刻意排除的基础设施只读规则，仅被 oneops-readonly 选择器聚合，不直接绑定；
// oneops-ops             运维人员 = 控制台核心资源全操作 + 节点/存储卷/命名空间/配额读取更新（不可删除）+ 其他资源只读；
// oneops-dev             开发人员 = 控制台可见资源读写（刻意不含 serviceaccounts/RBAC 资源，防间接提权）；
// oneops-view            受限用户 = 控制台可见资源只读（含 pods/log 日志读取）；
// 控制台可见资源清单由 dev/view 共用（console* 函数），控制台新增资源类型时在清单处统一维护
func presetClusterRoleManifests() []rbacv1.ClusterRole {
	return []rbacv1.ClusterRole{
		{
			// 管理员：全资源读写（与内置 cluster-admin 规则等价的静态副本）
			ObjectMeta: metav1.ObjectMeta{Name: "oneops-admin"},
			Rules: []rbacv1.PolicyRule{
				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "oneops-ops"},
			Rules: []rbacv1.PolicyRule{
				// 工作负载与网络/配置等控制台核心资源：全操作（显式列举，避免 "" 组 "*" 波及 nodes 等基础设施）
				{
					APIGroups: []string{""},
					Resources: []string{
						"pods", "pods/attach", "pods/exec", "pods/portforward", "pods/log",
						"services", "endpoints", "configmaps", "secrets", "events",
						"replicationcontrollers", "serviceaccounts",
					},
					Verbs: []string{"*"},
				},
				{
					APIGroups: []string{"apps", "extensions"},
					Resources: []string{"deployments", "statefulsets", "daemonsets", "replicasets", "controllerrevisions"},
					Verbs:     []string{"*"},
				},
				{
					APIGroups: []string{"batch"},
					Resources: []string{"jobs", "cronjobs"},
					Verbs:     []string{"*"},
				},
				{
					APIGroups: []string{"networking.k8s.io", "extensions"},
					Resources: []string{"ingresses", "networkpolicies"},
					Verbs:     []string{"*"},
				},
				{
					APIGroups: []string{"autoscaling"},
					Resources: []string{"horizontalpodautoscalers"},
					Verbs:     []string{"*"},
				},
				// 节点、存储卷、命名空间、配额：读取与更新（无 delete/create）
				{
					APIGroups: []string{""},
					Resources: []string{"nodes", "persistentvolumes", "namespaces", "resourcequotas", "limitranges"},
					Verbs:     []string{"get", "list", "watch", "update", "patch"},
				},
				// 其他全部资源：只读兜底
				{
					APIGroups: []string{"*"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			// 聚合源：view 刻意排除的基础设施只读，被 oneops-readonly 选中聚合
			ObjectMeta: metav1.ObjectMeta{
				Name:   "oneops-readonly-infra",
				Labels: map[string]string{"oneops.io/aggregate-to-readonly": "true"},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"nodes", "persistentvolumes"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{"storage.k8s.io"},
					Resources: []string{"storageclasses", "volumeattachments"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
		{
			// 只读管理员（聚合角色）：与内置 view 的构造方式一致——rules 不落静态清单，
			// 由 kube-controller-manager 的 RBAC 聚合控制器按下列选择器并集填充
			ObjectMeta: metav1.ObjectMeta{Name: "oneops-readonly"},
			AggregationRule: &rbacv1.AggregationRule{
				ClusterRoleSelectors: []metav1.LabelSelector{
					// 内置 view 的完整可见范围（secrets 依旧不含，与 view 安全语义一致）
					{MatchLabels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-view": "true"}},
					// 补齐缺口：节点/存储卷/存储类只读
					{MatchLabels: map[string]string{"oneops.io/aggregate-to-readonly": "true"}},
				},
			},
		},
		{
			// 开发人员：控制台可见资源读写；刻意不含 serviceaccounts/roles/rolebindings
			// （SA token 可被用于间接提权），权限边界即控制台管理域
			ObjectMeta: metav1.ObjectMeta{Name: "oneops-dev"},
			Rules:      append(consoleWriteRules(), consoleReadRules()...),
		},
		{
			// 受限用户：控制台可见资源只读（含 pods/log，控制台日志页依赖）
			ObjectMeta: metav1.ObjectMeta{Name: "oneops-view"},
			Rules:      consoleReadRules(),
		},
	}
}

// consoleReadRules 控制台可见资源的只读规则（受限用户 = 此清单；开发人员在其上叠加写权限）
func consoleReadRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{
				"pods", "pods/log", "pods/attach", "pods/portforward",
				"services", "endpoints", "configmaps", "secrets", "events",
				"replicationcontrollers", "persistentvolumeclaims",
			},
			Verbs: []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"apps", "extensions"},
			Resources: []string{"deployments", "statefulsets", "daemonsets", "replicasets", "controllerrevisions"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"jobs", "cronjobs"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"networking.k8s.io", "extensions"},
			Resources: []string{"ingresses", "networkpolicies"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"autoscaling"},
			Resources: []string{"horizontalpodautoscalers"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}
}

// consoleWriteRules 开发人员相对受限用户叠加的写规则：对只读清单中的工作负载/网络/配置资源全操作
func consoleWriteRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{
				"pods", "pods/attach", "pods/exec", "pods/portforward",
				"services", "endpoints", "configmaps", "secrets", "events",
				"replicationcontrollers", "persistentvolumeclaims",
			},
			Verbs: []string{"*"},
		},
		{
			APIGroups: []string{"apps", "extensions"},
			Resources: []string{"deployments", "statefulsets", "daemonsets", "replicasets", "controllerrevisions"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"jobs", "cronjobs"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"networking.k8s.io", "extensions"},
			Resources: []string{"ingresses", "networkpolicies"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"autoscaling"},
			Resources: []string{"horizontalpodautoscalers"},
			Verbs:     []string{"*"},
		},
	}
}

// nativeBindingRulesHashAnnotation 预设角色规则指纹：语义演进时指纹变化即触发 Update，
// 使"改档位语义 = 改这里的定义"，所有已绑定用户随下次授权/同步即时生效
const nativeBindingRulesHashAnnotation = "oneops.io/rules-hash"

// presetRulesHash 计算预设角色"平台定义部分"的指纹：静态角色取 Rules，
// 聚合角色取 AggregationRule（其 rules 由集群控制器填充，不参与比对）
func presetRulesHash(cr *rbacv1.ClusterRole) string {
	material := struct {
		Rules           []rbacv1.PolicyRule
		AggregationRule *rbacv1.AggregationRule
	}{Rules: cr.Rules, AggregationRule: cr.AggregationRule}
	if cr.AggregationRule != nil {
		material.Rules = nil
	}
	raw, err := json.Marshal(material)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// ensurePresetClusterRoles 确保预设 ClusterRole 在目标集群内存在且与平台定义一致：
// 不存在则创建；OneOps 管理的角色在规则指纹变化时更新（保留角色上可能存在的其他标签/注解）；
// 非 OneOps 管理的同名角色（用户手工创建）保持自治，不覆盖
func (s *K8sClusterService) ensurePresetClusterRoles(ctx context.Context, clientset *kubernetes.Clientset) error {
	for _, want := range presetClusterRoleManifests() {
		hash := presetRulesHash(&want)
		if want.Labels == nil {
			want.Labels = make(map[string]string)
		}
		want.Labels[nativeBindingManagedByLabel] = "oneops"
		want.Annotations = map[string]string{
			"oneops.io/preset-role":          "true",
			nativeBindingRulesHashAnnotation: hash,
		}

		existing, err := clientset.RbacV1().ClusterRoles().Get(ctx, want.Name, metav1.GetOptions{})
		if err == nil {
			if existing.Labels[nativeBindingManagedByLabel] != "oneops" {
				continue
			}
			if existing.Annotations[nativeBindingRulesHashAnnotation] == hash {
				continue
			}
			updated := existing.DeepCopy()
			updated.Rules = want.Rules
			updated.AggregationRule = want.AggregationRule
			if updated.Annotations == nil {
				updated.Annotations = make(map[string]string)
			}
			for k, v := range want.Annotations {
				updated.Annotations[k] = v
			}
			if _, err := clientset.RbacV1().ClusterRoles().Update(ctx, updated, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("更新预设 ClusterRole %s 失败: %w", want.Name, err)
			}
			logger.Info("已更新 OneOps 预设 ClusterRole（规则指纹变化）", zap.String("cluster_role", want.Name))
		} else if apierrors.IsNotFound(err) {
			if _, err := clientset.RbacV1().ClusterRoles().Create(ctx, &want, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
				return fmt.Errorf("创建预设 ClusterRole %s 失败: %w", want.Name, err)
			}
			logger.Info("已创建 OneOps 预设 ClusterRole", zap.String("cluster_role", want.Name))
		} else {
			return fmt.Errorf("查询预设 ClusterRole %s 失败: %w", want.Name, err)
		}
	}
	return nil
}

// createNativeK8sBinding 在集群内创建真实的 Binding 资源并回写资源名
func (s *K8sClusterService) createNativeK8sBinding(ctx context.Context, clientset *kubernetes.Clientset, record *modelk8s.K8sNativeRoleBinding) error {
	bindingName := fmt.Sprintf("oneops-native-%d", record.ID)

	kind := rbacv1.UserKind
	if record.SubjectType == "group" {
		kind = rbacv1.GroupKind
	}
	subject := rbacv1.Subject{
		Kind:      kind,
		APIGroup:  rbacv1.GroupName,
		Name:      record.ImpersonationName,
		Namespace: record.Namespace,
	}
	roleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     record.RoleKind,
		Name:     record.RoleName,
	}
	labels := map[string]string{nativeBindingManagedByLabel: "oneops"}

	// 按授权范围而非角色类型判定：Namespace 为空 = 集群级 ClusterRoleBinding；
	// Namespace 非空 = 命名空间级 RoleBinding（roleRef 可为 Role 或 ClusterRole，
	// 后者即"RB 引用 ClusterRole"的 ns 级授权，edit/view 限定命名空间的标准做法）
	if record.Namespace == "" {
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   bindingName,
				Labels: labels,
			},
			Subjects: []rbacv1.Subject{subject},
			RoleRef:  roleRef,
		}
		if _, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("集群内创建 ClusterRoleBinding 失败: %w", err)
		}
	} else {
		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: record.Namespace,
				Labels:    labels,
			},
			Subjects: []rbacv1.Subject{subject},
			RoleRef:  roleRef,
		}
		if _, err := clientset.RbacV1().RoleBindings(record.Namespace).Create(ctx, rb, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("集群内创建 RoleBinding 失败: %w", err)
		}
	}

	if err := s.clusterRepo.UpdateNativeBindingK8sName(record, bindingName); err != nil {
		logger.Warn("回写原生绑定集群资源名失败", zap.Uint("binding_id", record.ID), zap.Error(err))
	}
	return nil
}

// RevokeNativeBinding 撤销原生 RBAC 绑定：删除集群内 Binding（尽力而为）+ 删除 DB 记录
func (s *K8sClusterService) RevokeNativeBinding(operatorID uint, clusterID uint, bindingID uint) error {
	record, err := s.clusterRepo.FindNativeBindingByID(bindingID)
	if err != nil {
		return fmt.Errorf("原生绑定不存在: %w", err)
	}
	if record.ClusterID != clusterID {
		return fmt.Errorf("绑定与集群不匹配")
	}

	// 集群内删除（best-effort：集群不可达时仍删 DB，避免死锁；残留 Binding 可手工清理）
	clientset, err := s.GetClient(clusterID)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		// 与创建逻辑对称：按 Namespace 判定删除 CRB 或 RB
		if record.Namespace == "" {
			if err := clientset.RbacV1().ClusterRoleBindings().Delete(ctx, record.K8sBindingName, metav1.DeleteOptions{}); err != nil {
				logger.Warn("删除集群内 ClusterRoleBinding 失败",
					zap.String("binding", record.K8sBindingName), zap.Error(err))
			}
		} else {
			if err := clientset.RbacV1().RoleBindings(record.Namespace).Delete(ctx, record.K8sBindingName, metav1.DeleteOptions{}); err != nil {
				logger.Warn("删除集群内 RoleBinding 失败",
					zap.String("binding", record.K8sBindingName), zap.Error(err))
			}
		}
	}

	if err := s.clusterRepo.DeleteNativeBinding(bindingID); err != nil {
		return fmt.Errorf("删除原生绑定记录失败: %w", err)
	}

	logger.Info("撤销原生 RBAC 绑定",
		zap.Uint("binding_id", bindingID),
		zap.Uint("cluster_id", clusterID),
		zap.String("impersonation", record.ImpersonationName),
		zap.Uint("operator_id", operatorID))
	return nil
}

// GetSubjectOptions 授权主体候选（用户/用户组），供原生授权表单使用。
// 权限归属 k8s.permission.list，与"能否查看绑定"对齐，不借用 system.user.list
func (s *K8sClusterService) GetSubjectOptions() (map[string]interface{}, error) {
	users, groups, err := s.clusterRepo.FindSubjectOptions()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"users":  users,
		"groups": groups,
	}, nil
}

// GetRoleOptions 集群内角色候选，仅返回名称列表（kind=ClusterRole|Role，默认 ClusterRole）。
// 共享凭据查询属管理动作，与 AssignNativeBinding 的角色存在性校验同一链路
func (s *K8sClusterService) GetRoleOptions(clusterID uint, kind string) ([]string, error) {
	clientset, err := s.GetClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群连接失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	names := make([]string, 0)
	if kind == "Role" {
		list, err := clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 Role 列表失败: %w", err)
		}
		for i := range list.Items {
			names = append(names, list.Items[i].Name)
		}
	} else {
		list, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取 ClusterRole 列表失败: %w", err)
		}
		for i := range list.Items {
			names = append(names, list.Items[i].Name)
		}
	}
	sort.Strings(names)
	return names, nil
}
