package k8s

import (
	"encoding/json"
	"fmt"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ListIngress 获取 Ingress 列表（支持分页）
func (s *K8sResourceService) ListIngress(clusterID uint, namespace string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	list, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("获取 Ingress 列表失败: %w", err)
	}

	total := int64(len(list.Items))

	start := (page - 1) * pageSize
	end := start + pageSize

	if start < 0 {
		start = 0
	}
	totalInt := int(total)
	if end > totalInt {
		end = totalInt
	}

	items := list.Items
	if start >= len(items) {
		return []map[string]interface{}{}, total, nil
	}
	pagedItems := items[start:end]

	result := make([]map[string]interface{}, 0, len(pagedItems))
	for _, item := range pagedItems {
		result = append(result, s.formatIngress(&item))
	}

	return result, total, nil
}

// GetIngress 获取 Ingress 详情
func (s *K8sResourceService) GetIngress(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 详情失败: %w", err)
	}

	return s.formatIngressDetail(ingress), nil
}

// CreateIngress 创建 Ingress
func (s *K8sResourceService) CreateIngress(clusterID uint, namespace string, manifest map[string]interface{}) error {
	clientset, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ingress := &networkingv1.Ingress{}
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("序列化 Ingress manifest 失败: %w", err)
	}
	if err := json.Unmarshal(jsonData, ingress); err != nil {
		return fmt.Errorf("解析 Ingress manifest 失败: %w", err)
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		dynamicClient, dynErr := dynamic.NewForConfig(config)
		if dynErr != nil {
			return fmt.Errorf("创建 Ingress 失败（typed client）: %w", err)
		}

		gvr := schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1",
			Resource: "ingresses",
		}

		uns := &unstructured.Unstructured{}
		uns.SetUnstructuredContent(manifest)

		ctx2, cancel2 := s.createContextWithTimeout()
		defer cancel2()
		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx2, uns, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 Ingress 失败（dynamic client）: %w", err)
		}
	}

	logger.Info("创建 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", ingress.Name))

	return nil
}

// UpdateIngress 更新 Ingress
func (s *K8sResourceService) UpdateIngress(clusterID uint, namespace string, manifest map[string]interface{}) error {
	_, config, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	uns := &unstructured.Unstructured{}
	uns.SetUnstructuredContent(manifest)

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, uns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Ingress 失败: %w", err)
	}

	logger.Info("更新 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", uns.GetName()))

	return nil
}

// DeleteIngress 删除 Ingress
func (s *K8sResourceService) DeleteIngress(clusterID uint, namespace, name string) error {
	clientset, _, err := s.clientPool.GetClient(clusterID)
	if err != nil {
		return err
	}

	ctx, cancel := s.createContextWithTimeout()
	defer cancel()
	err = clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 Ingress 失败: %w", err)
	}

	logger.Warn("删除 Ingress 成功",
		zap.Uint("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("name", name))

	return nil
}

// formatIngress 格式化 Ingress 列表数据
func (s *K8sResourceService) formatIngress(ingress *networkingv1.Ingress) map[string]interface{} {
	result := map[string]interface{}{
		"name":             ingress.Name,
		"namespace":        ingress.Namespace,
		"age":              ingress.CreationTimestamp.Format("2006-01-02 15:04:05"),
		"hosts":            []string{},
		"addresses":        []string{},
		"ports":            []string{},
		"annotations":      ingress.Annotations,
		"ingressClassName": ingress.Spec.IngressClassName,
	}

	if ingress.Spec.Rules != nil {
		hosts := make([]string, 0)
		for _, rule := range ingress.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}
		result["hosts"] = hosts
	}

	if ingress.Spec.TLS != nil {
		tlsHosts := make([]string, 0)
		for _, tls := range ingress.Spec.TLS {
			tlsHosts = append(tlsHosts, tls.Hosts...)
		}
		result["tlsHosts"] = tlsHosts
	}

	if ingress.Status.LoadBalancer.Ingress != nil {
		addresses := make([]string, 0)
		for _, lb := range ingress.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				addresses = append(addresses, lb.IP)
			}
			if lb.Hostname != "" {
				addresses = append(addresses, lb.Hostname)
			}
		}
		result["addresses"] = addresses
	}

	return result
}

// formatIngressDetail 格式化 Ingress 详情数据
func (s *K8sResourceService) formatIngressDetail(ingress *networkingv1.Ingress) map[string]interface{} {
	result := s.formatIngress(ingress)

	manifestBytes, _ := json.Marshal(ingress)
	result["manifest"] = string(manifestBytes)

	if ingress.Spec.Rules != nil {
		rules := make([]map[string]interface{}, 0)
		for _, rule := range ingress.Spec.Rules {
			ruleMap := map[string]interface{}{
				"host": rule.Host,
			}
			if rule.HTTP != nil {
				paths := make([]map[string]interface{}, 0)
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, map[string]interface{}{
						"path":        path.Path,
						"pathType":    path.PathType,
						"serviceName": path.Backend.Service.Name,
						"servicePort": path.Backend.Service.Port,
					})
				}
				ruleMap["http"] = map[string]interface{}{
					"paths": paths,
				}
			}
			rules = append(rules, ruleMap)
		}
		result["rules"] = rules
	}

	if ingress.Spec.TLS != nil {
		tls := make([]map[string]interface{}, 0)
		for _, t := range ingress.Spec.TLS {
			tls = append(tls, map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			})
		}
		result["tls"] = tls
	}

	if ingress.Spec.IngressClassName != nil {
		result["ingressClassName"] = *ingress.Spec.IngressClassName
	}

	return result
}
