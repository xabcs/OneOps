package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oneops/backend2/pkg/logger"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AssignPermissionToGroup 为 JumpServer 用户组分配权限
// 实现 PermissionOperable 接口
func (j *JumpserverAdapter) AssignPermissionToGroup(authGroupID uint, mapping *AuthGroupPermissionMapping) error {
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
func (j *JumpserverAdapter) createAuthorizationRule(mapping *AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {
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
	baseURL := ""                          // 需要从服务中获取
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
func (j *JumpserverAdapter) assignAssetPermission(mapping *AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {
	// JumpServer 的资产权限通常通过授权规则来管理
	// 这里可以复用 createAuthorizationRule 的逻辑，但专门针对资产

	// 提取资产列表

	// 构建规则名称
	ruleName := fmt.Sprintf("权限映射_%d_资产权限", mapping.ID)

	// TODO: 实现实际的权限创建逻辑
	logger.Info("为用户组分配资产权限",
		zap.String("ruleName", ruleName),
		zap.String("externalID", mapping.ExternalID))

	return nil
}

// assignNodePermission 为用户组分配节点权限
func (j *JumpserverAdapter) assignNodePermission(mapping *AuthGroupPermissionMapping, permissionDetail map[string]interface{}) error {

	ruleName := fmt.Sprintf("权限映射_%d_节点权限", mapping.ID)

	// TODO: 实现实际的权限创建逻辑
	logger.Info("为用户组分配节点权限",
		zap.String("ruleName", ruleName),
		zap.String("externalID", mapping.ExternalID))

	return nil
}

// RevokePermissionFromGroup 撤销 JumpServer 用户组权限
// 实现 PermissionOperable 接口
func (j *JumpserverAdapter) RevokePermissionFromGroup(authGroupID uint, mapping *AuthGroupPermissionMapping) error {
	logger.Info("开始撤销 JumpServer 用户组权限",
		zap.Uint("authGroupId", authGroupID),
		zap.String("externalId", mapping.ExternalID))

	// JumpServer 撤销权限需要删除对应的授权规则
	// 从 externalId 中获取规则ID，然后调用删除API

	ruleID := mapping.ExternalID
	baseURL := ""                          // 需要从服务中获取
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
func (j *JumpserverAdapter) CheckGroupPermission(authGroupID uint, mapping *AuthGroupPermissionMapping) (bool, error) {
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

// AuthorizationRuleUser 授权规则用户
type AuthorizationRuleUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// AuthorizationRuleUserGroup 授权规则用户组
type AuthorizationRuleUserGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AuthorizationRuleAsset 授权规则资产
type AuthorizationRuleAsset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host string `json:"hostname"`
	IP   string `json:"ip"`
}

// AuthorizationRuleNode 授权规则节点
type AuthorizationRuleNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AuthorizationRuleDetail 授权规则详情
type AuthorizationRuleDetail struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name"`
	Users      []AuthorizationRuleUser      `json:"users"`
	UserGroups []AuthorizationRuleUserGroup `json:"user_groups"`
	Assets     []AuthorizationRuleAsset     `json:"assets"`
	Nodes      []AuthorizationRuleNode      `json:"nodes"`
	Actions    interface{}                  `json:"actions"` // 改为 interface{} 以支持不同格式
	IsActive   bool                         `json:"is_active"`
	IsExpired  bool                         `json:"is_expired"`
}

// GetAuthorizationRuleDetail 获取授权规则详情（包括用户列表）
func (j *JumpserverAdapter) GetAuthorizationRuleDetail(baseURL string, authConfig map[string]interface{}, ruleID string) (*AuthorizationRuleDetail, error) {
	// 构建 API URL
	detailURL := strings.TrimSuffix(baseURL, "/") + "/api/v1/perms/asset-permissions/" + ruleID + "/"

	logger.Info("获取 JumpServer 授权规则详情",
		zap.String("ruleID", ruleID),
		zap.String("url", detailURL))

	// 创建请求
	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证
	if err := j.setAuthentication(req, authConfig); err != nil {
		return nil, fmt.Errorf("设置认证失败: %w", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取授权规则详情失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var ruleDetail AuthorizationRuleDetail
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, &ruleDetail); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logger.Info("成功获取 JumpServer 授权规则详情",
		zap.String("ruleID", ruleID),
		zap.String("ruleName", ruleDetail.Name),
		zap.Int("userCount", len(ruleDetail.Users)),
		zap.Int("userGroupCount", len(ruleDetail.UserGroups)),
		zap.Int("assetCount", len(ruleDetail.Assets)),
		zap.Int("nodeCount", len(ruleDetail.Nodes)))

	return &ruleDetail, nil
}

// AddUsersToAuthorizationRule 添加用户到授权规则
func (j *JumpserverAdapter) AddUsersToAuthorizationRule(baseURL string, authConfig map[string]interface{}, ruleID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 先获取当前授权规则详情
	detailURL := strings.TrimSuffix(baseURL, "/") + "/api/v1/perms/asset-permissions/" + ruleID + "/"

	logger.Info("获取 JumpServer 授权规则详情以添加用户",
		zap.String("ruleID", ruleID))

	// 创建获取请求
	getReq, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return fmt.Errorf("创建获取请求失败: %w", err)
	}

	// 设置认证
	if err := j.setAuthentication(getReq, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}

	// 发送获取请求
	client := &http.Client{}
	getResp, err := client.Do(getReq)
	if err != nil {
		return fmt.Errorf("获取授权规则失败: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResp.Body)
		return fmt.Errorf("获取授权规则详情失败 (状态码 %d): %s", getResp.StatusCode, string(body))
	}

	// 解析当前规则详情
	var currentRule AuthorizationRuleDetail
	body, err := io.ReadAll(getResp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, &currentRule); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取现有用户ID列表
	existingUserIDs := make([]string, len(currentRule.Users))
	for i, user := range currentRule.Users {
		existingUserIDs[i] = user.ID
	}

	// 合并新旧用户ID（去重）
	allUserIDs := make(map[string]bool)
	for _, id := range existingUserIDs {
		allUserIDs[id] = true
	}
	for _, id := range userIDs {
		allUserIDs[id] = true
	}

	// 构建新的用户ID列表
	updatedUserIDs := make([]string, 0, len(allUserIDs))
	for id := range allUserIDs {
		updatedUserIDs = append(updatedUserIDs, id)
	}

	// 构建更新数据（使用与获取相同的结构）
	updateData := map[string]interface{}{
		"users": updatedUserIDs,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("序列化更新数据失败: %w", err)
	}

	logger.Info("更新 JumpServer 授权规则用户列表",
		zap.String("ruleID", ruleID),
		zap.Int("previousUserCount", len(existingUserIDs)),
		zap.Int("newUsersToAdd", len(userIDs)),
		zap.Int("totalUserCount", len(updatedUserIDs)))

	// 创建更新请求
	updateReq, err := http.NewRequest("PATCH", detailURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建更新请求失败: %w", err)
	}

	// 设置认证和请求头
	if err := j.setAuthentication(updateReq, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")

	// 发送更新请求
	updateResp, err := client.Do(updateReq)
	if err != nil {
		return fmt.Errorf("更新请求失败: %w", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK && updateResp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(updateResp.Body)
		return fmt.Errorf("添加用户到授权规则失败 (状态码 %d): %s", updateResp.StatusCode, string(body))
	}

	logger.Info("成功添加用户到 JumpServer 授权规则",
		zap.String("ruleID", ruleID),
		zap.Any("addedUserIDs", userIDs))

	return nil
}

// RemoveUsersFromAuthorizationRule 从授权规则移除用户
func (j *JumpserverAdapter) RemoveUsersFromAuthorizationRule(baseURL string, authConfig map[string]interface{}, ruleID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 先获取当前授权规则详情
	detailURL := strings.TrimSuffix(baseURL, "/") + "/api/v1/perms/asset-permissions/" + ruleID + "/"

	logger.Info("获取 JumpServer 授权规则详情以移除用户",
		zap.String("ruleID", ruleID))

	// 创建获取请求
	getReq, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return fmt.Errorf("创建获取请求失败: %w", err)
	}

	// 设置认证
	if err := j.setAuthentication(getReq, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}

	// 发送获取请求
	client := &http.Client{}
	getResp, err := client.Do(getReq)
	if err != nil {
		return fmt.Errorf("获取授权规则失败: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResp.Body)
		return fmt.Errorf("获取授权规则详情失败 (状态码 %d): %s", getResp.StatusCode, string(body))
	}

	// 解析当前规则详情
	var currentRule AuthorizationRuleDetail
	body, err := io.ReadAll(getResp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, &currentRule); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 构建用户ID到用户的映射
	userMap := make(map[string]*AuthorizationRuleUser)
	for i := range currentRule.Users {
		userMap[currentRule.Users[i].ID] = &currentRule.Users[i]
	}

	// 移除指定的用户
	for _, userID := range userIDs {
		delete(userMap, userID)
	}

	// 构建新的用户ID列表
	updatedUserIDs := make([]string, 0, len(userMap))
	for id := range userMap {
		updatedUserIDs = append(updatedUserIDs, id)
	}

	// 构建更新数据
	updateData := map[string]interface{}{
		"users": updatedUserIDs,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("序列化更新数据失败: %w", err)
	}

	logger.Info("从 JumpServer 授权规则移除用户",
		zap.String("ruleID", ruleID),
		zap.Int("previousUserCount", len(currentRule.Users)),
		zap.Int("usersToRemove", len(userIDs)),
		zap.Int("remainingUserCount", len(updatedUserIDs)))

	// 创建更新请求
	updateReq, err := http.NewRequest("PATCH", detailURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建更新请求失败: %w", err)
	}

	// 设置认证和请求头
	if err := j.setAuthentication(updateReq, authConfig); err != nil {
		return fmt.Errorf("设置认证失败: %w", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")

	// 发送更新请求
	updateResp, err := client.Do(updateReq)
	if err != nil {
		return fmt.Errorf("更新请求失败: %w", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK && updateResp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(updateResp.Body)
		return fmt.Errorf("从授权规则移除用户失败 (状态码 %d): %s", updateResp.StatusCode, string(body))
	}

	logger.Info("成功从 JumpServer 授权规则移除用户",
		zap.String("ruleID", ruleID),
		zap.Any("removedUserIDs", userIDs))

	return nil
}
