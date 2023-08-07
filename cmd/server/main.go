package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"FirewallIPSyncer/firewall"
	"FirewallIPSyncer/log"
)

// EnvVars 腾讯云轻量服务器环境变量
type EnvVars struct {
	// SecretID 和 SecretKey 的获取地址：https://console.cloud.tencent.com/cam/capi
	SecretID  string
	SecretKey string
	// Region 如 ap-guangzhou
	Region string
	// InstanceID 实例ID
	InstanceID string
	// Tag 给当前运行的服务器一个标签，用于区分不同的服务器的IP
	Tag string
}

//go:generate bash -c "GOOS=linux GOARCH=amd64 go build -o ./bin/server"
func main() {
	env, err := GetEnvVars()
	if err != nil {
		log.Error.Fatalln(err)
	}

	client, err := firewall.CreateClient(env.SecretID, env.SecretKey, env.Region, "lighthouse.tencentcloudapi.com")
	if err != nil {
		log.Error.Fatalln(err)
	}

	ip, err := GetClientIP()
	if err != nil {
		log.Error.Fatalln(err)
	}

	if err := firewall.UpdateFirewallRule(client, env.InstanceID, env.Tag, ip); err != nil {
		log.Error.Println(err)
	}
}

// GetEnvVars 获取环境变量
func GetEnvVars() (*EnvVars, error) {
	var envVars EnvVars
	vars := []struct {
		envVarName string
		envVarVal  *string
	}{
		{"SecretID", &envVars.SecretID},
		{"SecretKey", &envVars.SecretKey},
		{"Region", &envVars.Region},
		{"InstanceID", &envVars.InstanceID},
		{"Tag", &envVars.Tag},
	}
	for _, v := range vars {
		*v.envVarVal = os.Getenv(v.envVarName)
		if *v.envVarVal == "" {
			return nil, fmt.Errorf("%s not set in environment", v.envVarName)
		}
	}
	return &envVars, nil
}

// GetClientIP 获取 SSH 连接的公网 IP
func GetClientIP() (string, error) {
	log.Info.Println("GetClientIP: get client ip from SSH_CLIENT")
	if sshClient, exist := os.LookupEnv("SSH_CLIENT"); exist {
		sshClientSlice := strings.Split(sshClient, " ")
		ip := strings.TrimSpace(sshClientSlice[0])
		return ip, nil
	}
	return "", errors.New("env: SSH_CLIENT Not Exist")
}
