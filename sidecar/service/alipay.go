package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/smartwalle/alipay/v3"
)

func getAlipayClient() (*alipay.Client, error) {
	if !setting.AlipayEnabled || setting.AlipayAppId == "" || setting.AlipayPrivateKey == "" || setting.AlipayPublicKey == "" {
		return nil, errors.New("支付宝未配置")
	}

	isProduction := setting.AlipayGatewayUrl == "" || setting.AlipayGatewayUrl == "https://openapi.alipay.com/gateway.do"

	client, err := alipay.New(setting.AlipayAppId, setting.AlipayPrivateKey, isProduction)
	if err != nil {
		return nil, fmt.Errorf("初始化支付宝客户端失败: %w", err)
	}

	if err := client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
	}

	return client, nil
}

// CreatePageOrder 创建支付宝电脑网站支付订单，返回跳转链接
func CreatePageOrder(tradeNo string, totalAmount float64, subject, returnUrl, notifyUrl string) (string, error) {
	client, err := getAlipayClient()
	if err != nil {
		return "", err
	}

	var p alipay.TradePagePay
	p.NotifyURL = notifyUrl
	p.ReturnURL = returnUrl
	p.Subject = subject
	p.OutTradeNo = tradeNo
	p.TotalAmount = fmt.Sprintf("%.2f", totalAmount)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	payUrl, err := client.TradePagePay(p)
	if err != nil {
		return "", fmt.Errorf("创建支付宝订单失败: %w", err)
	}

	return payUrl.String(), nil
}

// VerifyAlipayNotify 验证支付宝回调签名
func VerifyAlipayNotify(params url.Values) (bool, error) {
	client, err := getAlipayClient()
	if err != nil {
		return false, err
	}
	if err := client.VerifySign(context.Background(), params); err != nil {
		return false, err
	}
	return true, nil
}

// GetAlipayMoney 计算支付宝支付金额（人民币，单位：元）
func GetAlipayMoney(amount float64, group string) float64 {
	originalAmount := amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = amount / common.QuotaPerUnit
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	discount := 1.0
	if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(originalAmount)]; ok {
		if ds > 0 {
			discount = ds
		}
	}
	return amount * setting.AlipayUnitPrice * topupGroupRatio * discount
}

func GetAlipayMinTopup() int64 {
	minTopup := setting.AlipayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}
