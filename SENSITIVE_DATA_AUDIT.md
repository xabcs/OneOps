# 敏感数据加密存储安全审查报告

## 审查日期
2026年6月9日

## 审查范围
- SSH凭证存储（密码、私钥、 passphrase）
- 用户密码存储
- 配置文件中的敏感信息
- Token和会话信息

## 发现的安全问题

### 🔴 P0 严重漏洞：SSH凭证明文存储

#### 问题1: SSH密码明文存储
**文件**: `services/cmdb.go:1056-1066`

```go
// 🔴 严重问题：密码明文存储！
func (s *CMDBService) CreateSSHCredential(credential *models.SSHCredential) error {
    // 加密密码和私钥
    if credential.Password != "" {
        // TODO: 实际应该使用加密算法加密  <-- 未实现！
        credential.Password = credential.Password  // 明文存储！
    }
    if credential.PrivateKey != "" {
        // TODO: 实际应该使用加密算法加密  <-- 未实现！
        credential.PrivateKey = credential.PrivateKey  // 明文存储！
    }
    return db.Create(credential).Error
}
```

**风险等级**: 🔴 严重（P0）
**影响**: 攻击者获取数据库访问权限后，可直接读取所有SSH密码和私钥
**数据泄露影响**: 完全控制所有服务器

#### 问题2: 更新时的加密TODO
**文件**: `services/cmdb.go:1070-1080`

```go
func (s *CMDBService) UpdateSSHCredential(id uint, updates map[string]interface{}) error {
    // 如果有密码或私钥更新，需要加密
    if _, ok := updates["password"]; ok {
        // TODO: 加密处理  <-- 未实现
    }
    if _, ok := updates["private_key"]; ok {
        // TODO: 加密处理  <-- 未实现
    }
    return db.Model(&models.SSHCredential{}).Where("id = ?", id).Updates(updates).Error
}
```

#### 问题3: 私钥和Passphrase明文存储
同样的问题，私钥和passphrase都是明文存储。

### ✅ 安全的部分

#### 1. 用户密码哈希存储
**文件**: `utils/password.go`

```go
// ✅ 正确：使用bcrypt哈希
func HashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hashedBytes), err
}
```

**状态**: 用户密码已经正确使用bcrypt哈希存储 ✅

## 加密方案

### 推荐加密方案：AES-256-GCM

对于SSH凭证等敏感数据，推荐使用以下加密方案：

1. **对称加密**: AES-256-GCM
   - 密钥长度：32字节
   - Nonce：12字节
   - 认证标签：16字节

2. **密钥管理**:
   - 主密钥存储在环境变量或密钥管理系统中
   - 每次部署时生成新的数据加密密钥
   - 使用密钥派生函数（HKDF）从主密钥派生数据加密密钥

3. **实施步骤**:
   - 创建 `utils/encryption.go` 加密工具
   - 在保存SSH凭证前加密
   - 在使用SSH凭证前解密
   - 配置文件中存储加密密钥

## 修复方案

### 立即修复（P0）

#### 1. 创建加密工具模块

```go
// utils/encryption.go
package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "fmt"
)

var encryptionKey []byte

// InitEncryption 初始化加密密钥
func InitEncryption() error {
    keyStr := os.Getenv("ENCRYPTION_KEY")
    if keyStr == "" {
        return errors.New("ENCRYPTION_KEY环境变量未设置")
    }
    
    key, err := base64.URLEncoding.DecodeString(keyStr)
    if err != nil {
        return fmt.Errorf("解析加密密钥失败: %w", err)
    }
    
    if len(key) != 32 {
        return fmt.Errorf("加密密钥长度错误，期望32字节，实际%d字节", len(key))
    }
    
    encryptionKey = key
    return nil
}

// EncryptData 加密数据
func EncryptData(plaintext []byte) (string, error) {
    block, err := aes.NewCipher(encryptionKey, aes.NewGCM(nonce[:12], nil, nil))
    if err != nil {
        return "", fmt.Errorf("创建加密块失败: %w", err)
    }
    
    ciphertext := block.Seal(nil, plaintext, nil)
    nonce := nonce[:]
    
    // 返回格式: nonce(12) + ciphertext + tag(16)
    result := append(nonce[:], ciphertext...)
    result = append(result, block.GetTag()...)
    
    return base64.URLEncoding.EncodeToString(result), nil
}

// DecryptData 解密数据
func DecryptData(ciphertext string) ([]byte, error) {
    data, err := base64.URLEncoding.DecodeString(ciphertext)
    if err != nil {
        return nil, fmt.Errorf("Base64解码失败: %w", err)
    }
    
    if len(data) < 28 { // 12(nonce) + 16(tag) = 28
        return nil, errors.New("密文长度不足")
    }
    
    nonce := data[:12]
    tag := data[len(data)-16:]
    ciphertext := data[12 : len(data)-16]
    
    block, err := aes.NewCipher(encryptionKey, aes.NewGCM(nonce, nil, nil))
    if err != nil {
        return nil, fmt.Errorf("创建解密块失败: %w", err)
    }
    
    plaintext, err := block.Open(nil, ciphertext, nonce, tag)
    if err != nil {
        return nil, fmt.Errorf("解密失败: %w", err)
    }
    
    return plaintext, nil
}
```

#### 2. 修改CreateSSHCredential

```go
func (s *CMDBService) CreateSSHCredential(credential *models.SSHCredential) error {
    // 加密敏感字段
    if credential.Password != "" {
        encrypted, err := utils.EncryptData([]byte(credential.Password))
        if err != nil {
            return fmt.Errorf("加密密码失败: %w", err)
        }
        credential.Password = encrypted
    }
    
    if credential.PrivateKey != "" {
        encrypted, err := utils.EncryptData([]byte(credential.PrivateKey))
        if err != nil {
            return fmt.Errorf("加密私钥失败: %w", err)
        }
        credential.PrivateKey = encrypted
    }
    
    if credential.Passphrase != "" {
        encrypted, err := utils.EncryptData([]byte(credential.Passphrase))
        if err != nil {
            return fmt.Errorf("加密passphrase失败: %w", err)
        }
        credential.Passphrase = encrypted
    }
    
    return db.Create(credential).Error
}
```

#### 3. 添加解密方法

```go
func (s *CMDBService) GetSSHCredentialByID(id uint) (*models.SSHCredential, error) {
    var credential models.SSHCredential
    if err := db.First(&credential, id).Error; err != nil {
        return nil, err
    }
    
    // 解密敏感字段
    if credential.Password != "" {
        decrypted, err := utils.DecryptData(credential.Password)
        if err != nil {
            return nil, fmt.Errorf("解密密码失败: %w", err)
        }
        credential.Password = string(decrypted)
    }
    
    if credential.PrivateKey != "" {
        decrypted, err := utils.DecryptData(credential.PrivateKey)
        if err != nil {
            return nil, fmt.Errorf("解密私钥失败: %w", err)
        }
        credential.PrivateKey = string(decrypted)
    }
    
    // ... 其他解密
    
    return &credential, nil
}
```

### 优先级排序

| 优先级 | 任务 | 状态 |
|--------|------|------|
| P0 | 实现SSH凭证加密 | 🔴 紧急 |
| P0 | 实现解密功能 | 🔴 紧急 |
| P0 | 修复UpdateSSHCredential | 🔴 紧急 |
| P0 | 配置加密密钥管理 | 🔴 紧急 |
| P1 | 数据迁移脚本 | 中 |
| P1 | 密钥轮换机制 | 中 |

## 风险评估

### 当前风险
- **风险等级**: 🔴 严重
- **CVSS评分**: ~8.5 (High)
- **CWE类别**: CWE-312 (密钥管理不当), CWE-311 (缺失加密)
- **影响范围**: 所有SSH凭证存储

### 潜在攻击场景
1. **数据库泄露**：攻击者获取数据库后，得到所有SSH明文密码
2. **备份泄露**：包含明文密码的备份文件被窃取
3. **日志泄露**：敏感信息可能被记录到日志
4. **内部威胁**：有数据库访问权限的内部人员可以读取所有密码

### 业务影响
- **服务器完全沦陷**：攻击者可以连接所有服务器
- **横向移动**：使用获取的凭证访问其他系统
- **数据窃取**：直接访问所有服务器上的数据
- **系统破坏**：删除或修改所有数据

## 修复时间估算
- **开发时间**: 2-3小时
- **测试时间**: 1小时
- **数据迁移**: 1-2小时
- **总计**: 4-6小时

## 立即行动项

1. **立即执行**：
   - [ ] 创建 `utils/encryption.go` 加密工具
   - [ ] 修改 `CreateSSHCredential` 实现加密
   - [ ] 修改 `UpdateSSHCredential` 实现加密
   - [ ] 添加解密方法到 `GetSSHCredentialByID`
   - [ ] 设置 `ENCRYPTION_KEY` 环境变量

2. **数据迁移**：
   - [ ] 编写数据迁移脚本
   - [ ] 加密现有SSH凭证数据
   - [ ] 验证解密功能正常

3. **测试验证**：
   - [ ] 单元测试（加密/解密）
   - [ ] 集成测试（SSH凭证创建和使用）
   - [ ] 回归测试（现有功能不受影响）

---

**审查人**: Claude Code AI  
**审查状态**: 🔴 发现严重漏洞  
**建议**: 立即修复
