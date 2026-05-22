package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"oneops/backend/logger"
	"oneops/backend/models"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// SyncServerHardwareConfig 通过 SSH 异步采集目标主机硬件配置，写入数据库
func SyncServerHardwareConfig(serverID uint) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("SSH硬件采集 panic", zap.Any("recover", r), zap.Uint("serverID", serverID))
		}
	}()

	// 1. 从 DB 读取 Server（预加载 SSHCredential）
	var server models.Server
	if err := db.Preload("SSHCredential").First(&server, serverID).Error; err != nil {
		logger.Warn("SSH硬件采集：读取主机失败", zap.Uint("serverID", serverID), zap.Error(err))
		return
	}

	// 2. 无凭证则尝试通过 CredentialID/SSHCredentialID 兼容加载
	if server.SSHCredential == nil {
		credID := server.SSHCredentialID
		if credID == 0 {
			credID = server.CredentialID
		}
		if credID == 0 {
			logger.Warn("SSH硬件采集：主机无 SSH 凭证，跳过", zap.Uint("serverID", serverID))
			return
		}
		var cred models.SSHCredential
		if err := db.First(&cred, credID).Error; err != nil {
			logger.Warn("SSH硬件采集：读取凭证失败", zap.Uint("serverID", serverID), zap.Error(err))
			return
		}
		server.SSHCredential = &cred
	}

	// 3. 建立 SSH 连接
	sshClient, err := dialSSH(&server)
	if err != nil {
		logger.Warn("SSH硬件采集：连接失败", zap.Uint("serverID", serverID), zap.String("ip", server.IP), zap.Error(err))
		return
	}
	defer sshClient.Close()

	// 4. 采集硬件配置
	updates := map[string]interface{}{
		"metrics_updated_at": time.Now(),
	}

	if cpu, err := runSSHCommand(sshClient, "nproc 2>/dev/null || grep -c ^processor /proc/cpuinfo 2>/dev/null || echo 0"); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(cpu)); err == nil && v > 0 {
			updates["cpu"] = v
		}
	}

	if mem, err := runSSHCommand(sshClient, "free -m 2>/dev/null | awk '/Mem:/{print int($2/1024)}' || echo 0"); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(mem)); err == nil && v > 0 {
			updates["memory"] = v
		}
	}

	if disk, err := runSSHCommand(sshClient, "df -BG / 2>/dev/null | awk 'NR==2{gsub(/G/,\"\",$2);print $2}' || echo 0"); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(disk)); err == nil && v > 0 {
			updates["disk"] = v
		}
	}

	// 5. 写入数据库
	if err := db.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error; err != nil {
		logger.Warn("SSH硬件采集：写入数据库失败", zap.Uint("serverID", serverID), zap.Error(err))
		return
	}

	logger.Info("SSH硬件采集：完成", zap.Uint("serverID", serverID), zap.Any("updates", updates))
}

// dialSSH 建立 SSH 连接（支持密码和私钥认证）
func dialSSH(server *models.Server) (*ssh.Client, error) {
	credential := server.SSHCredential
	if credential == nil {
		return nil, fmt.Errorf("无SSH凭证")
	}

	var authMethods []ssh.AuthMethod
	if credential.Password != "" {
		authMethods = append(authMethods, ssh.Password(credential.Password))
	}
	if credential.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(credential.PrivateKey))
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("无可用认证方法")
	}

	sshPort := server.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	loginUser := credential.Username
	if loginUser == "" {
		loginUser = server.SSHUser
	}
	if loginUser == "" {
		loginUser = "root"
	}

	cfg := &ssh.ClientConfig{
		User:            loginUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         10 * time.Second,
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.IP, sshPort), cfg)
}

// runSSHCommand 在远程主机上执行命令并返回输出
func runSSHCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
