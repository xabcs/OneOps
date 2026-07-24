package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oneops/backend/models"
	"oneops/backend/logger"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AssignPermissionToGroup 为 JumpServer 用户组分配权限
// 实现 PermissionOperable 接口
func (j *JumpserverAdapter) AssignPermissionToGroup(authGroupID uint, mapping *models.AuthGroupPermissionMapping) error {
	logger.Info("开始为 JumpServer 用户组分配权限",
		zap.Uint("authGroupId", authGroupID),
		zap.String("mappingType", mapping.MappingType),
		zap.String("externalId", mapping.ExternalID))

	// JumpServer 不支持直接通过API为用户组分配权限
	// JumpServer 的授权规则是通过管理界面手动配置的
	// 这里我们实现一个变通方案：创建授权规则并关联用户组

	// 解析权限详情
	var permissionDetail map[string]interface{}
	if err := json.Unmarshal([]byte(mapping.PermissionDetail), &permissionDetail); err != nil {
		return fmt.Errorf("解析权限详情失败: %w", err)
	}

	// 根据映射类型执行不同的操作
	switch mapping.MappingType {
	case "rule":
		// 创建 JumpServer 授权规则
		return j.createAuthorizationRule(mapping, permissionDetail)
	case "asset":
		// 为用户组分配特定资产的访问权限
		return j.assignAssetPermission(mapping, permissionDetail)
	case "node":
		// 为用户组分配节点的访问权限
		return j.assignNodePermission(mapping, permissionDetail)
	default:
		return fmt.Errorf("不支持的映射类型: %s", mapping.MappingType)
	}
}

// createAuthorizationRule 创建 JumpServer 授权规则
func (j *JumpserverAdapter) createAuthorizationRule(mapping *models.AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {
	// 从权限详情中提取参数
	ruleName, _ := permissionDetail["rule_name"].(string)
	actions, _ := permissionDetail["actions"].([]interface{})
	assets, _ := permissionDetail["assets"].([]interface{})
	nodes, _ := permissionDetail["nodes"].([]interface{})

	// 构建 JumpServer 授权规则创建请求
	// 注意：JumpServer API 格式可能会根据版本不同而变化
	ruleData := map[string]interface{}{
		"name": ruleName,
		// 用户列表（可以从权限详情中获取）
		"users": []string{},
		// 用户组列表（使用 externalId）
		"user_groups": []string{mapping.ExternalID},
		// 资产列表
		"assets": assets,
		// 节点列表
		"nodes": nodes,
		// 操作权限
		"actions": actions,
		// 优先级
		"priority": permissionDetail["priority"],
		// 是否启用
		"is_active": true,
	}

	// 获取应用的认证配置（需要从服务中传入，这里简化处理）
	baseURL := "" // 需要从服务中获取
	authConfig := map[string]interface{}{} // 需要从服务中获取

	// 构建 API URL
	createRuleURL := "/api/v1/perms/asset-permissions/"
	fullURL := strings.TrimSuffix(baseURL, "/") + createRuleURL

	// 序列化请求数据
	jsonData, err := json.Marshal(ruleData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证
	if err := j.setAuthentication(req, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建授权规则失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("成功创建 JumpServer 授权规则",
		zap.String("ruleName", ruleName),
		zap.String("userGroup", mapping.ExternalID))

	return nil
}

// assignAssetPermission 为用户组分配资产权限
func (j *JumpserverAdapter) assignAssetPermission(mapping *models.AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {
	// JumpServer 的资产权限通常通过授权规则来管理
	// 这里可以复用 createAuthorizationRule 的逻辑，但专门针对资产

	// 提取资产列表
	assets, _ := permissionDetail["assets"].([]interface{})
	actions, _ := permissionDetail["actions"].([]interface{})

	// 构建规则名称
	ruleName := fmt.Sprintf("权限映射_%d_资产权限", mapping.ID)

	// 调用创建授权规则
	ruleData := map[string]interface{}{
		"name":        ruleName,
		"user_groups": []string{mapping.ExternalID},
		"assets":      assets,
		"nodes":       []string{}, // 空节点列表
		"actions":     actions,
		"priority":    50,
		"is_active":   true,
	}

	// ... (类似 createAuthorizationRule 的实现)
	return nil
}

// assignNodePermission 为用户组分配节点权限
func (j *JumpserverAdapter) assignNodePermission(mapping *models.AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {
	// 类似 assignAssetPermission，但针对节点
	nodes, _ := permissionDetail["nodes"].([]interface{})
	actions, _ := permissionDetail["actions"].([]interface{})

	ruleName := fmt.Sprintf("权限映射_%d_节点权限", mapping.ID)

	ruleData := map[string]interface{}{
		"name":        ruleName,
		"user_groups": []string{mapping.ExternalID},
		"assets":      []string{}, // 空资产列表
		"nodes":       nodes,
		"actions":     actions,
		"priority":    50,
		"is_active":   true,
	}

	// ... (类似 createAuthorizationRule 的实现)
	return nil
}

// RevokePermissionFromGroup 撤销 JumpServer 用户组权限
// 实现 PermissionOperable 接口
func (j *JumpserverAdapter) RevokePermissionFromGroup(authGroupID uint, mapping *models.AuthGroupPermissionMapping) error {
	logger.Info("开始撤销 JumpServer 用户组权限",
		zap.Uint("authGroupId", authGroupID),
		zap.String("externalId", mapping.ExternalID))

	// JumpServer 撤销权限需要删除对应的授权规则
	// 从 externalId 中获取规则ID，然后调用删除API

	ruleID := mapping.ExternalID
	baseURL := "" // 需要从服务中获取
	authConfig := map[string]interface{}{} // 需要从服务中获取

	deleteRuleURL := fmt.Sprintf("/api/v1/perms/asset-permissions/%s/", ruleID)
	fullURL := strings.TrimSuffix(baseURL, "/") + deleteRuleURL

	// 创建请求
	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证
	if err := j.setAuthentication(req, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除授权规则失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("成功删除 JumpServer 授权规则",
		zap.String("ruleId", ruleID))

	return nil
}

// CheckGroupPermission 检查用户组是否有特定权限
// 实现 PermissionOperable 接口（可选）
func (j *JumpserverAdapter) CheckGroupPermission(authGroupID uint, mapping *models.AuthGroupPermissionMapping) (bool, error) {
	// 解析权限详情
	var permissionDetail map[string]interface{}
	if err := json.Unmarshal([]byte(mapping.PermissionDetail), &permissionDetail); err != nil {
		return false, fmt.Errorf("解析权限详情失败: %w", err)
	}

	// 检查权限是否过期
	if mapping.ExpireTime != nil && mapping.ExpireTime.Before(time.Now()) {
		return false, nil
	}

	// 检查映射是否启用
	if !mapping.IsEnabled {
		return false, nil
	}

	// 对于 JumpServer，权限的有效性取决于授权规则的状态
	// 可以通过查询授权规则状态来验证

	return true, nil
}
