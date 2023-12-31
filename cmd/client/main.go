package main

import (
	"fmt"
	"io"
	"net/http"
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

//go:generate bash -c "GOOS=linux GOARCH=arm64 go build -o ./bin/client"
func main() {
	env, err := GetEnvVars()
	if err != nil {
		log.Error.Fatalln(err)
	}

	client, err := firewall.CreateClient(env.SecretID, env.SecretKey, env.Region, "lighthouse.tencentcloudapi.com")
	if err != nil {
		log.Error.Fatalln(err)
	}

	ip, err := GetPublicIP()
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

// GetPublicIP 获取公网IP
func GetPublicIP() (string, error) {
	log.Info.Println("GetPublicIP: sending request to http://myexternalip.com/raw")

	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", fmt.Errorf("GetPublicIP: error get ip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetPublicIP: unexpected status code: %d", resp.StatusCode)
	}

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetPublicIP: error read response: %w", err)
	}

	log.Info.Printf("GetPublicIP: got IP: %s", ip)
	return strings.TrimSpace(string(ip)), nil
}
