package service

import (
	"strconv"

	"github.com/QuantumNous/new-api/setting"
)

// AppendNativePayMethods 将微信支付/支付宝支付追加到支付方式列表中。
// 返回更新后的列表以及两种支付是否启用的状态。
func AppendNativePayMethods(payMethods []map[string]string) ([]map[string]string, bool, bool) {
	enableWechatPay := setting.WechatPayEnabled &&
		setting.WechatPayAppId != "" &&
		setting.WechatPayMchId != "" &&
		setting.WechatPayApiKey != ""
	if enableWechatPay {
		hasWechatPay := false
		for _, method := range payMethods {
			if method["type"] == "wechatpay_native" {
				hasWechatPay = true
				break
			}
		}
		if !hasWechatPay {
			payMethods = append(payMethods, map[string]string{
				"name":      "微信支付",
				"type":      "wechatpay_native",
				"color":     "rgba(var(--semi-green-5), 1)",
				"min_topup": strconv.Itoa(setting.WechatPayMinTopUp),
			})
		}
	}

	enableAlipay := setting.AlipayEnabled &&
		setting.AlipayAppId != "" &&
		setting.AlipayPrivateKey != "" &&
		setting.AlipayPublicKey != ""
	if enableAlipay {
		hasAlipay := false
		for _, method := range payMethods {
			if method["type"] == "alipay_page" {
				hasAlipay = true
				break
			}
		}
		if !hasAlipay {
			payMethods = append(payMethods, map[string]string{
				"name":      "支付宝",
				"type":      "alipay_page",
				"color":     "rgba(var(--semi-blue-5), 1)",
				"min_topup": strconv.Itoa(setting.AlipayMinTopUp),
			})
		}
	}

	return payMethods, enableWechatPay, enableAlipay
}
