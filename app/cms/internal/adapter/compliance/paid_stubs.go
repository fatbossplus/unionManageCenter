// Package compliance 付费云合规检查驱动桩
package compliance

import (
	"errors"
	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterChecker(&AliyunChecker{})
	adapter.RegisterChecker(&BaiduChecker{})
}

// AliyunChecker 阿里云内容安全（¥0.003/次）
type AliyunChecker struct {
	AccessKeyID     string
	AccessKeySecret string
}

func (a *AliyunChecker) Name() string { return "aliyun_paid" }
func (a *AliyunChecker) Check(text, platform string) (adapter.ComplianceResult, error) {
	if a.AccessKeyID == "" {
		return adapter.ComplianceResult{}, errors.New("阿里云内容安全：请先配置 access_key_id 和 access_key_secret")
	}
	// TODO: 调用阿里云绿网 API
	// POST https://green-cip.cn-shanghai.aliyuncs.com/green/v2/text/scan
	return adapter.ComplianceResult{}, errors.New("阿里云内容安全驱动开发中，请参考文档 https://help.aliyun.com/product/28415.html")
}

// BaiduChecker 百度内容审核（¥0.001/次）
type BaiduChecker struct {
	APIKey    string
	SecretKey string
}

func (b *BaiduChecker) Name() string { return "baidu_paid" }
func (b *BaiduChecker) Check(text, platform string) (adapter.ComplianceResult, error) {
	if b.APIKey == "" {
		return adapter.ComplianceResult{}, errors.New("百度内容审核：请先配置 api_key 和 secret_key")
	}
	// TODO: POST https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined
	return adapter.ComplianceResult{}, errors.New("百度内容审核驱动开发中")
}
