package cmdb

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// dialSSH 建立 SSH 连接（支持密码和私钥认证）
// 注意：此函数仅用于 Agent 部署/重启/卸载操作，不用于数据采集
func dialSSH(server *Server) (*ssh.Client, error) {
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
