package setting

var (
	WechatPayEnabled   bool
	WechatPayAppId     string
	WechatPayMchId     string
	WechatPayApiKey    string
	WechatPayUnitPrice float64 = 7.3
	WechatPayMinTopUp  int     = 1

	AlipayEnabled      bool
	AlipayAppId        string
	AlipayPrivateKey   string
	AlipayPublicKey    string
	AlipayGatewayUrl   string
	AlipayUnitPrice    float64 = 7.3
	AlipayMinTopUp     int     = 1
)
