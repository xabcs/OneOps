package cmdb

import (
	"fmt"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"go.uber.org/zap"
)

// ========== SSH凭证管理 ==========

// credentialMask 敏感字段掩码：任何 API 出口（列表/详情）都不允许明文或密文出网
const credentialMask = "******"

// maskCredentialSensitive 将凭证敏感字段替换为掩码（在返回 API 前调用）
func maskCredentialSensitive(c *modelcmdb.SSHCredential) {
	if c.Password != "" {
		c.Password = credentialMask
	}
	if c.PrivateKey != "" {
		c.PrivateKey = credentialMask
	}
	if c.Passphrase != "" {
		c.Passphrase = credentialMask
	}
}

// decryptCredential 将凭证解密到局部副本，仅供连接路径（dialSSH/Test）内部使用，绝不返回给 API 层
func (s *CMDBService) decryptCredential(c *modelcmdb.SSHCredential) (*modelcmdb.SSHCredential, error) {
	cp := *c
	if cp.Password != "" {
		decrypted, err := utils.DecryptString(cp.Password)
		if err != nil {
			logger.Warn("SSH凭证密码解密失败", zap.Uint("credential_id", c.ID), zap.Error(err))
			return nil, fmt.Errorf("密码解密失败: %w", err)
		}
		cp.Password = decrypted
	}
	if cp.PrivateKey != "" {
		decrypted, err := utils.DecryptString(cp.PrivateKey)
		if err != nil {
			logger.Warn("SSH凭证私钥解密失败", zap.Uint("credential_id", c.ID), zap.Error(err))
			return nil, fmt.Errorf("私钥解密失败: %w", err)
		}
		cp.PrivateKey = decrypted
	}
	if cp.Passphrase != "" {
		decrypted, err := utils.DecryptString(cp.Passphrase)
		if err != nil {
			logger.Warn("SSH凭证passphrase解密失败", zap.Uint("credential_id", c.ID), zap.Error(err))
			return nil, fmt.Errorf("passphrase解密失败: %w", err)
		}
		cp.Passphrase = decrypted
	}
	return &cp, nil
}

// GetSSHCredentials 获取SSH凭证列表（敏感字段以掩码返回，不出网）
func (s *CMDBService) GetSSHCredentials(credentialType string) ([]modelcmdb.SSHCredential, error) {
	credentials, err := s.credRepo.FindSSHCredentials(credentialType)
	if err != nil {
		return nil, err
	}
	for i := range credentials {
		maskCredentialSensitive(&credentials[i])
	}
	return credentials, nil
}

// GetSSHCredentialByID 根据ID获取SSH凭证（敏感字段以掩码返回，不出网）
func (s *CMDBService) GetSSHCredentialByID(id uint) (*modelcmdb.SSHCredential, error) {
	credential, err := s.credRepo.FindSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}
	maskCredentialSensitive(credential)
	return credential, nil
}

// CreateSSHCredential 创建SSH凭证
func (s *CMDBService) CreateSSHCredential(credential *modelcmdb.SSHCredential) error {
	if credential.Password != "" {
		encrypted, err := utils.EncryptString(credential.Password)
		if err != nil {
			return fmt.Errorf("加密密码失败: %w", err)
		}
		credential.Password = encrypted
	}

	if credential.PrivateKey != "" {
		encrypted, err := utils.EncryptString(credential.PrivateKey)
		if err != nil {
			return fmt.Errorf("加密私钥失败: %w", err)
		}
		credential.PrivateKey = encrypted
	}

	if credential.Passphrase != "" {
		encrypted, err := utils.EncryptString(credential.Passphrase)
		if err != nil {
			return fmt.Errorf("加密passphrase失败: %w", err)
		}
		credential.Passphrase = encrypted
	}

	return s.credRepo.CreateSSHCredential(credential)
}

// UpdateSSHCredential 更新SSH凭证
func (s *CMDBService) UpdateSSHCredential(id uint, updates map[string]interface{}) error {
	// 敏感字段防误伤：值等于掩码（列表/详情回显值被原样回传）或空串（编辑表单"不修改请留空"）一律视为"不修改"，
	// 必须先从更新集中剔除——否则掩码会被二次加密写库覆盖真实密码，空串会把密文清空
	dropUnchangedSensitiveFields(updates)

	if password, ok := updates["password"]; ok {
		if passwordStr, ok := password.(string); ok {
			encrypted, err := utils.EncryptString(passwordStr)
			if err != nil {
				return fmt.Errorf("加密密码失败: %w", err)
			}
			updates["password"] = encrypted
		}
	}

	if privateKey, ok := updates["private_key"]; ok {
		if privateKeyStr, ok := privateKey.(string); ok {
			encrypted, err := utils.EncryptString(privateKeyStr)
			if err != nil {
				return fmt.Errorf("加密私钥失败: %w", err)
			}
			updates["private_key"] = encrypted
		}
	}

	if passphrase, ok := updates["passphrase"]; ok {
		if passphraseStr, ok := passphrase.(string); ok {
			encrypted, err := utils.EncryptString(passphraseStr)
			if err != nil {
				return fmt.Errorf("加密passphrase失败: %w", err)
			}
			updates["passphrase"] = encrypted
		}
	}

	return s.credRepo.UpdateSSHCredential(id, updates)
}

// dropUnchangedSensitiveFields 剔除更新请求中不应触发变更的敏感字段（值非字符串、空串或等于掩码）
func dropUnchangedSensitiveFields(updates map[string]interface{}) {
	for _, key := range []string{"password", "private_key", "passphrase"} {
		v, ok := updates[key]
		if !ok {
			continue
		}
		str, isStr := v.(string)
		if !isStr || str == "" || str == credentialMask {
			delete(updates, key)
		}
	}
}

// DeleteSSHCredential 删除SSH凭证
func (s *CMDBService) DeleteSSHCredential(id uint) error {
	return s.credRepo.DeleteSSHCredential(id)
}

// TestSSHCredential 测试SSH凭证连接
func (s *CMDBService) TestSSHCredential(id uint, testIP string, testPort int) (map[string]interface{}, error) {
	stored, err := s.credRepo.FindSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}
	// 解密仅发生在内部副本上，用于后续真实连接，不进入返回值
	if _, err := s.decryptCredential(stored); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["success"] = true
	result["message"] = "连接测试功能开发中"
	result["credential"] = stored.Username
	result["test_ip"] = testIP
	result["test_port"] = testPort

	return result, nil
}

// ClearAgentRecord 清空服务器的 Agent 相关字段
func (s *CMDBService) ClearAgentRecord(id uint) error {
	return s.repo.ClearAgentRecord(id)
}

// ========== 辅助函数 ==========

// filterServerColumns 过滤掉非数据库列字段
func filterServerColumns(updates map[string]interface{}) map[string]interface{} {
	skipKeys := map[string]bool{
		"cloudInfo":     true,
		"credential":    true,
		"credentials":   true,
		"credentialIds": true,
		"cabinet":       true,
		"tags":          true,
		"groups":        true,
		"groupIds":      true,
		"sshCredential": true,
		"business":      true,
		"cloudInfoData": true,
		"id":            true,
		"createdAt":     true,
	}
	result := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if !skipKeys[k] {
			result[k] = v
		}
	}
	return result
}

// extractUintSlice 从 interface{} 中提取 []uint
func extractUintSlice(val interface{}) []uint {
	switch v := val.(type) {
	case []uint:
		return v
	case []interface{}:
		result := make([]uint, 0, len(v))
		for _, item := range v {
			if f, ok := item.(float64); ok {
				result = append(result, uint(f))
			}
		}
		return result
	}
	return nil
}

// rawExists 检查 key 是否存在于 map 中
func rawExists(updates map[string]interface{}, key string) bool {
	_, exists := updates[key]
	return exists
}

// parseUint 辅助函数：将interface{}解析为uint
func parseUint(value interface{}) uint {
	switch v := value.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	case float32:
		return uint(v)
	case string:
		if parsedVal, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(parsedVal)
		}
	}
	return 0
}

// parseGroupIDUint 解析 groupID（用于 GetServersLight 中的类型转换）
func parseGroupIDUint(groupID interface{}) uint {
	return parseUint(groupID)
}
