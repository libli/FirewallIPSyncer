package firewall

import (
	"fmt"

	"FirewallIPSyncer/log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

// CreateClient 创建腾讯云轻量服务调用 Client
func CreateClient(secretId string, secretKey string, region string, endpoint string) (*lighthouse.Client, error) {
	log.Info.Println("CreateClient: starting CreateClient...")
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endpoint
	client, err := lighthouse.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("CreateClient: error creating client with secretId %s: %w", secretId, err)
	}

	return client, nil
}

// UpdateFirewallRule 更新防火墙规则
func UpdateFirewallRule(client *lighthouse.Client, instanceID, tag, ip string) error {
	log.Info.Println("UpdateFirewallRule: starting UpdateRules for instance", instanceID)

	rules, err := getRules(client, instanceID)
	if err != nil {
		return fmt.Errorf("UpdateFirewallRule: error getting rules for instance %s: %w", instanceID, err)
	}

	ruleInfo, found := findRule(rules, ip, tag)
	if found {
		log.Info.Println("UpdateFirewallRule: no need to update")
		return nil
	}

	if ruleInfo != nil {
		// 删除与tag匹配但IP不匹配的规则
		log.Info.Println("UpdateFirewallRule: deleting rule")
		if err := deleteFirewallRule(client, instanceID, convertToFirewallRule(ruleInfo)); err != nil {
			return fmt.Errorf("UpdateFirewallRule: error deleteFirewallRule for instance %s: %w", instanceID, err)
		}
		log.Info.Println("UpdateFirewallRule: successfully deleted firewall rule")
	} else {
		log.Info.Println("UpdateFirewallRule: no rule found, creating new rule")
	}

	// 创建新规则
	rule := &lighthouse.FirewallRule{
		Protocol:                common.StringPtr("TCP"),
		Port:                    common.StringPtr("ALL"),
		CidrBlock:               &ip,
		Action:                  common.StringPtr("ACCEPT"),
		FirewallRuleDescription: &tag,
	}
	if err := createFirewallRule(client, instanceID, rule); err != nil {
		return fmt.Errorf("UpdateFirewallRule: error createFirewallRule for instance %s: %w", instanceID, err)
	}
	log.Info.Println("UpdateFirewallRule: successfully created firewall rule")
	return nil
}

// getRules 获取防火墙规则
func getRules(client *lighthouse.Client, instanceID string) ([]*lighthouse.FirewallRuleInfo, error) {
	log.Info.Println("getRules: starting GetRules for instance", instanceID)
	request := lighthouse.NewDescribeFirewallRulesRequest()
	request.InstanceId = &instanceID
	response, err := client.DescribeFirewallRules(request)
	if err != nil {
		return nil, fmt.Errorf("getRules: error getting rules for instance %s: %w", instanceID, err)
	}
	if response.Response == nil {
		return nil, fmt.Errorf("getRules: response is nil for instance %s", instanceID)
	}

	log.Info.Printf("getRules: RequestId: %s, FirewallVersion: %d, TotalCount: %d",
		*response.Response.RequestId, *response.Response.FirewallVersion, *response.Response.TotalCount)

	return response.Response.FirewallRuleSet, nil
}

// findRule 查找防火墙规则
func findRule(rules []*lighthouse.FirewallRuleInfo, ip, tag string) (*lighthouse.FirewallRuleInfo, bool) {
	// 定义一个变量来存储找到的只匹配标签的规则
	var tagMatchedRule *lighthouse.FirewallRuleInfo

	// 遍历规则，寻找匹配的IP或标签
	for _, rule := range rules {
		if rule.CidrBlock != nil && *rule.CidrBlock == ip {
			// 找到匹配的IP，返回规则和完全匹配的标志
			return rule, true
		}
		if rule.FirewallRuleDescription != nil && *rule.FirewallRuleDescription == tag {
			// 找到匹配的标签但IP不匹配，存储规则
			tagMatchedRule = rule
		}
	}

	// 如果找到了匹配标签的规则，返回存储的规则和非完全匹配的标志
	if tagMatchedRule != nil {
		return tagMatchedRule, false
	}

	// 未找到匹配的规则，返回nil和非完全匹配的标志
	return nil, false
}

// createFirewallRule 创建防火墙规则
func createFirewallRule(client *lighthouse.Client, instanceID string, rule *lighthouse.FirewallRule) error {
	log.Info.Println("createFirewallRule: starting CreateFirewallRule for instance", instanceID)
	request := lighthouse.NewCreateFirewallRulesRequest()
	request.InstanceId = &instanceID
	request.FirewallRules = []*lighthouse.FirewallRule{rule}

	response, err := client.CreateFirewallRules(request)
	if err != nil {
		return fmt.Errorf("createFirewallRule: error creating firewall rule for instance %s: %w", instanceID, err)
	}
	if response.Response == nil {
		return fmt.Errorf("createFirewallRule: response is nil for instance %s", instanceID)
	}

	log.Info.Printf("createFirewallRule: RequestId: %s", *response.Response.RequestId)

	return nil
}

// deleteFirewallRule 删除防火墙规则
func deleteFirewallRule(client *lighthouse.Client, instanceID string, rule *lighthouse.FirewallRule) error {
	log.Info.Println("deleteFirewallRule: starting DeleteFirewallRule for instance", instanceID)
	request := lighthouse.NewDeleteFirewallRulesRequest()
	request.InstanceId = &instanceID
	request.FirewallRules = []*lighthouse.FirewallRule{rule}

	response, err := client.DeleteFirewallRules(request)
	if err != nil {
		return fmt.Errorf("deleteFirewallRule: error deleting firewall rules for instance %s: %w", instanceID, err)
	}

	if response.Response == nil {
		return fmt.Errorf("deleteFirewallRule: response is nil for instance %s", instanceID)
	}

	log.Info.Printf("deleteFirewallRule: RequestId: %s", *response.Response.RequestId)

	return nil
}

// convertToFirewallRule 将防火墙规则进行转换
func convertToFirewallRule(ruleInfo *lighthouse.FirewallRuleInfo) *lighthouse.FirewallRule {
	if ruleInfo == nil {
		return nil
	}

	return &lighthouse.FirewallRule{
		Protocol:                ruleInfo.Protocol,
		Port:                    ruleInfo.Port,
		CidrBlock:               ruleInfo.CidrBlock,
		Action:                  ruleInfo.Action,
		FirewallRuleDescription: ruleInfo.FirewallRuleDescription,
	}
}
