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

// GetSSHCredentials 获取SSH凭证列表
func (s *CMDBService) GetSSHCredentials(credentialType string) ([]modelcmdb.SSHCredential, error) {
	return s.credRepo.FindSSHCredentials(credentialType)
}

// GetSSHCredentialByID 根据ID获取SSH凭证
func (s *CMDBService) GetSSHCredentialByID(id uint) (*modelcmdb.SSHCredential, error) {
	credential, err := s.credRepo.FindSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}

	if credential.Password != "" {
		decrypted, err := utils.DecryptString(credential.Password)
		if err != nil {
			logger.Warn("SSH凭证密码解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.Password = decrypted
		}
	}

	if credential.PrivateKey != "" {
		decrypted, err := utils.DecryptString(credential.PrivateKey)
		if err != nil {
			logger.Warn("SSH凭证私钥解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.PrivateKey = decrypted
		}
	}

	if credential.Passphrase != "" {
		decrypted, err := utils.DecryptString(credential.Passphrase)
		if err != nil {
			logger.Warn("SSH凭证passphrase解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.Passphrase = decrypted
		}
	}

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
	if password, ok := updates["password"]; ok && password != "" {
		if passwordStr, ok := password.(string); ok {
			encrypted, err := utils.EncryptString(passwordStr)
			if err != nil {
				return fmt.Errorf("加密密码失败: %w", err)
			}
			updates["password"] = encrypted
		}
	}

	if privateKey, ok := updates["private_key"]; ok && privateKey != "" {
		if privateKeyStr, ok := privateKey.(string); ok {
			encrypted, err := utils.EncryptString(privateKeyStr)
			if err != nil {
				return fmt.Errorf("加密私钥失败: %w", err)
			}
			updates["private_key"] = encrypted
		}
	}

	if passphrase, ok := updates["passphrase"]; ok && passphrase != "" {
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

// DeleteSSHCredential 删除SSH凭证
func (s *CMDBService) DeleteSSHCredential(id uint) error {
	return s.credRepo.DeleteSSHCredential(id)
}

// TestSSHCredential 测试SSH凭证连接
func (s *CMDBService) TestSSHCredential(id uint, testIP string, testPort int) (map[string]interface{}, error) {
	credential, err := s.GetSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["success"] = true
	result["message"] = "连接测试功能开发中"
	result["credential"] = credential.Username
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
