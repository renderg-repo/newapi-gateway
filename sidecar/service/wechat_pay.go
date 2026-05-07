package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

const (
	WechatPayUnifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
)

// WechatPayUnifiedOrderRequest 微信统一下单请求参数
type WechatPayUnifiedOrderRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppId          string   `xml:"appid"`
	MchId          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int64    `xml:"total_fee"`
	SpbillCreateIp string   `xml:"spbill_create_ip"`
	NotifyUrl      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
}

// WechatPayUnifiedOrderResponse 微信统一下单响应
type WechatPayUnifiedOrderResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	AppId      string   `xml:"appid"`
	MchId      string   `xml:"mch_id"`
	NonceStr   string   `xml:"nonce_str"`
	Sign       string   `xml:"sign"`
	ResultCode string   `xml:"result_code"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
	PrepayId   string   `xml:"prepay_id"`
	CodeUrl    string   `xml:"code_url"`
	TradeType  string   `xml:"trade_type"`
}

// WechatPayNotify 微信支付回调通知
type WechatPayNotify struct {
	XMLName        xml.Name `xml:"xml"`
	ReturnCode     string   `xml:"return_code"`
	ReturnMsg      string   `xml:"return_msg"`
	AppId          string   `xml:"appid"`
	MchId          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	ResultCode     string   `xml:"result_code"`
	ErrCode        string   `xml:"err_code"`
	OpenId         string   `xml:"openid"`
	IsSubscribe    string   `xml:"is_subscribe"`
	TradeType      string   `xml:"trade_type"`
	BankType       string   `xml:"bank_type"`
	TotalFee       int64    `xml:"total_fee"`
	CashFee        int64    `xml:"cash_fee"`
	TransactionId  string   `xml:"transaction_id"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TimeEnd        string   `xml:"time_end"`
}

func getWechatPayConfig() (appId, mchId, apiKey string, enabled bool) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return common.OptionMap["WechatPayAppId"],
		common.OptionMap["WechatPayMchId"],
		common.OptionMap["WechatPayApiKey"],
		common.OptionMap["WechatPayEnabled"] == "true"
}

func GetWechatPayUnitPrice() float64 {
	return setting.WechatPayUnitPrice
}

func GetWechatPayMinTopUp() int64 {
	return int64(setting.WechatPayMinTopUp)
}

// GenerateNonceStr 生成随机字符串
func GenerateNonceStr(length int) string {
	return common.GetRandomString(length)
}

// WechatPaySign 生成微信支付 MD5 签名
func WechatPaySign(params map[string]string, apiKey string) string {
	// 按 key 排序拼接
	var keys []string
	for k := range params {
		if k == "sign" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	// 简单排序，确保稳定
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
	}
	b.WriteString("&key=")
	b.WriteString(apiKey)

	h := md5.New()
	h.Write([]byte(b.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// CreateNativeOrder 创建微信 Native 支付订单，返回扫码链接
func CreateNativeOrder(tradeNo string, totalFee int64, body, notifyUrl string) (string, error) {
	appId, mchId, apiKey, enabled := getWechatPayConfig()
	if !enabled || appId == "" || mchId == "" || apiKey == "" {
		return "", errors.New("微信支付未配置")
	}

	nonceStr := GenerateNonceStr(32)
	params := map[string]string{
		"appid":            appId,
		"mch_id":           mchId,
		"nonce_str":        nonceStr,
		"body":             body,
		"out_trade_no":     tradeNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       notifyUrl,
		"trade_type":       "NATIVE",
	}
	sign := WechatPaySign(params, apiKey)

	reqXML := fmt.Sprintf(`<xml>
<appid><![CDATA[%s]]></appid>
<mch_id><![CDATA[%s]]></mch_id>
<nonce_str><![CDATA[%s]]></nonce_str>
<sign><![CDATA[%s]]></sign>
<body><![CDATA[%s]]></body>
<out_trade_no><![CDATA[%s]]></out_trade_no>
<total_fee>%d</total_fee>
<spbill_create_ip><![CDATA[127.0.0.1]]></spbill_create_ip>
<notify_url><![CDATA[%s]]></notify_url>
<trade_type><![CDATA[NATIVE]]></trade_type>
</xml>`, appId, mchId, nonceStr, sign, body, tradeNo, totalFee, notifyUrl)

	resp, err := http.Post(WechatPayUnifiedOrderURL, "application/xml", bytes.NewBufferString(reqXML))
	if err != nil {
		return "", fmt.Errorf("请求微信支付统一下单失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取微信支付响应失败: %w", err)
	}

	var orderResp WechatPayUnifiedOrderResponse
	if err := xml.Unmarshal(respBody, &orderResp); err != nil {
		return "", fmt.Errorf("解析微信支付响应失败: %w", err)
	}

	if orderResp.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("微信支付返回失败: %s", orderResp.ReturnMsg)
	}
	if orderResp.ResultCode != "SUCCESS" {
		return "", fmt.Errorf("微信支付下单失败: %s - %s", orderResp.ErrCode, orderResp.ErrCodeDes)
	}

	return orderResp.CodeUrl, nil
}

// VerifyWechatPayNotify 验证微信支付回调签名
func VerifyWechatPayNotify(notify *WechatPayNotify) bool {
	_, _, apiKey, enabled := getWechatPayConfig()
	if !enabled || apiKey == "" {
		return false
	}

	params := map[string]string{
		"appid":         notify.AppId,
		"mch_id":        notify.MchId,
		"nonce_str":     notify.NonceStr,
		"result_code":   notify.ResultCode,
		"openid":        notify.OpenId,
		"is_subscribe":  notify.IsSubscribe,
		"trade_type":    notify.TradeType,
		"bank_type":     notify.BankType,
		"total_fee":     fmt.Sprintf("%d", notify.TotalFee),
		"cash_fee":      fmt.Sprintf("%d", notify.CashFee),
		"transaction_id": notify.TransactionId,
		"out_trade_no":  notify.OutTradeNo,
		"time_end":      notify.TimeEnd,
	}

	// 移除空值
	cleaned := make(map[string]string)
	for k, v := range params {
		if v != "" {
			cleaned[k] = v
		}
	}

	expectedSign := WechatPaySign(cleaned, apiKey)
	return notify.Sign == expectedSign
}

// BuildWechatPayNotifyResponse 构建微信支付回调响应
func BuildWechatPayNotifyResponse(success bool) string {
	if success {
		return "<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>"
	}
	return "<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[FAIL]]></return_msg></xml>"
}

// GetWechatPayMoney 计算微信支付金额（人民币，单位：分）
func GetWechatPayMoney(amount float64, group string) int64 {
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
	payMoney := amount * setting.WechatPayUnitPrice * topupGroupRatio * discount
	// 转换为分
	return int64(payMoney * 100)
}

func GetWechatPayMinTopup() int64 {
	minTopup := setting.WechatPayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}
