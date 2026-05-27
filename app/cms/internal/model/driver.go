package model

// DriverConfig 驱动配置（免费/付费切换，每个 config_key 代表一个能力维度）
type DriverConfig struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID      uint64 `gorm:"column:org_id;uniqueIndex:uk_org_key" json:"org_id"`
	ConfigKey  string `gorm:"column:config_key;size:64;uniqueIndex:uk_org_key" json:"config_key"`
	DriverName string `gorm:"column:driver_name;size:64"  json:"driver_name"`
	DriverType string `gorm:"column:driver_type;size:20"  json:"driver_type"` // free|paid
	ConfigJSON string `gorm:"column:config_json;type:text" json:"config_json"` // 加密后的JSON
	Enabled    int8   `gorm:"column:enabled;default:1"    json:"enabled"`
	CreatedAt  string `gorm:"column:created_at"           json:"created_at"`
	UpdatedAt  string `gorm:"column:updated_at"           json:"updated_at"`
}

func (DriverConfig) TableName() string { return "cms_driver_configs" }

// DriverKey 驱动配置键常量
const (
	DriverKeyWechatScraper   = "wechat.scraper"
	DriverKeyRednoteScraper  = "rednote.scraper"
	DriverKeyDouyinScraper   = "douyin.scraper"
	DriverKeyCsdnScraper     = "csdn.scraper"
	DriverKeyAIRewriter      = "ai.rewriter"
	DriverKeyComplianceLocal = "compliance.local"
	DriverKeyComplianceAPI   = "compliance.api"
	DriverKeyProxyPool       = "proxy.pool"
	DriverKeyWechatPublisher = "wechat.publisher"
	DriverKeyRednotePublisher = "rednote.publisher"
	DriverKeyDouyinPublisher = "douyin.publisher"
	DriverKeyCsdnPublisher   = "csdn.publisher"
)

// DriverName 驱动名称常量
const (
	// 采集驱动
	DriverSogouFree      = "sogou_free"
	DriverRSSHubFree     = "rsshub_free"
	DriverPlaywrightFree = "playwright_free"
	DriverOpenapiPaid    = "openapi_paid"
	DriverNewrankPaid    = "newrank_paid"
	DriverQianguaPaid    = "qiangua_paid"
	DriverCsdnRSSFree    = "rss_free"
	// AI 驱动
	DriverOllamaFree     = "ollama_free"
	DriverTongyiPaid     = "tongyi_paid"
	DriverOpenAIPaid     = "openai_paid"
	DriverWenxinPaid     = "wenxin_paid"
	DriverZhipuPaid      = "zhipu_paid"
	// 合规驱动
	DriverWordlistFree   = "wordlist_free"
	DriverAliyunPaid     = "aliyun_paid"
	DriverBaiduPaid      = "baidu_paid"
	DriverTencentPaid    = "tencent_paid"
	// 代理驱动
	DriverDirectFree     = "direct_free"
	DriverZhimaPaid      = "zhima_paid"
	// 发布驱动
	DriverWechatOfficial = "official_api"
	DriverCsdnUnofficial = "unofficial_api"
)

// DriverMeta 驱动元信息（前端展示用，不存库）
type DriverMeta struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"` // free|paid
	Desc        string `json:"desc"`
	CostDesc    string `json:"cost_desc"`
	Stability   int    `json:"stability"` // 1-5 星
	NeedConfig  bool   `json:"need_config"`
	ConfigSchema []ConfigField `json:"config_schema,omitempty"`
}

type ConfigField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`  // text|password|select|url
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
}

// AllDriverMeta 所有可用驱动的元信息
var AllDriverMeta = map[string][]DriverMeta{
	DriverKeyWechatScraper: {
		{Name: DriverSogouFree, DisplayName: "搜狗微信搜索", Type: "free",
			Desc: "解析搜狗微信频道，免费可用，有反爬限制", CostDesc: "免费", Stability: 3,
			NeedConfig: false},
		{Name: DriverRSSHubFree, DisplayName: "RSSHub 自建", Type: "free",
			Desc: "自建 RSSHub 实例，覆盖约60%公众号", CostDesc: "服务器成本", Stability: 4,
			NeedConfig: true,
			ConfigSchema: []ConfigField{{Key: "rsshub_url", Label: "RSSHub 地址", Type: "url", Required: true, Placeholder: "https://rsshub.example.com"}}},
		{Name: DriverNewrankPaid, DisplayName: "新榜 API（付费）", Type: "paid",
			Desc: "正规数据商，稳定覆盖全量公众号", CostDesc: "¥500-2000/月", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{{Key: "api_key", Label: "API Key", Type: "password", Required: true, Placeholder: "请输入新榜 API Key"}}},
		{Name: "yimei_paid", DisplayName: "易媒助手 API（付费）", Type: "paid",
			Desc: "专注新媒体数据，接入简单", CostDesc: "¥299-999/月", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{{Key: "api_key", Label: "API Key", Type: "password", Required: true}}},
	},
	DriverKeyCsdnScraper: {
		{Name: DriverCsdnRSSFree, DisplayName: "CSDN RSS 订阅", Type: "free",
			Desc: "标准 Atom Feed，零风险，完全免费", CostDesc: "免费", Stability: 5, NeedConfig: false},
	},
	DriverKeyRednoteScraper: {
		{Name: DriverPlaywrightFree, DisplayName: "Playwright 自动化（免费）", Type: "free",
			Desc: "模拟浏览器采集，需绑定账号Cookie，有封号风险", CostDesc: "服务器+代理IP", Stability: 2, NeedConfig: false},
		{Name: DriverQianguaPaid, DisplayName: "千瓜数据 API（付费）", Type: "paid",
			Desc: "行业级小红书数据平台，合规稳定", CostDesc: "¥500-3000/月", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{{Key: "api_key", Label: "API Key", Type: "password", Required: true}}},
	},
	DriverKeyAIRewriter: {
		{Name: DriverOllamaFree, DisplayName: "Ollama 本地模型（免费）", Type: "free",
			Desc: "本地推理，数据不外传，质量依赖GPU性能", CostDesc: "GPU服务器成本", Stability: 3,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "base_url", Label: "Ollama 地址", Type: "url", Required: true, Placeholder: "http://localhost:11434"},
				{Key: "model", Label: "模型名称", Type: "text", Required: true, Placeholder: "qwen2.5:7b"},
			}},
		{Name: DriverTongyiPaid, DisplayName: "通义千问（付费·国内合规）", Type: "paid",
			Desc: "阿里云大模型，国内合规优先推荐", CostDesc: "¥0.04-0.12/千token", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				{Key: "model", Label: "模型", Type: "text", Required: false, Placeholder: "qwen-max"},
			}},
		{Name: DriverWenxinPaid, DisplayName: "文心一言（付费·国内合规）", Type: "paid",
			Desc: "百度大模型，中文优化", CostDesc: "¥0.04-0.12/千token", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				{Key: "secret_key", Label: "Secret Key", Type: "password", Required: true},
			}},
		{Name: DriverOpenAIPaid, DisplayName: "OpenAI GPT-4o（付费）", Type: "paid",
			Desc: "质量最高，需境外网络", CostDesc: "¥0.3/千token", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				{Key: "base_url", Label: "代理地址（可选）", Type: "url", Required: false, Placeholder: "https://api.openai.com"},
			}},
	},
	DriverKeyComplianceLocal: {
		{Name: DriverWordlistFree, DisplayName: "本地敏感词库（免费）", Type: "free",
			Desc: "50w+词条本地匹配，零延迟零成本", CostDesc: "免费", Stability: 3, NeedConfig: false},
	},
	DriverKeyComplianceAPI: {
		{Name: DriverAliyunPaid, DisplayName: "阿里云内容安全（付费）", Type: "paid",
			Desc: "覆盖多维度，识别率最高", CostDesc: "¥0.003/次", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "access_key_id", Label: "AccessKeyId", Type: "text", Required: true},
				{Key: "access_key_secret", Label: "AccessKeySecret", Type: "password", Required: true},
			}},
		{Name: DriverBaiduPaid, DisplayName: "百度内容审核（付费）", Type: "paid",
			Desc: "中文语义优化，价格最低", CostDesc: "¥0.001/次", Stability: 5,
			NeedConfig: true,
			ConfigSchema: []ConfigField{
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				{Key: "secret_key", Label: "Secret Key", Type: "password", Required: true},
			}},
	},
}
