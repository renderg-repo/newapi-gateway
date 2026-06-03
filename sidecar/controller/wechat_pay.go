package controller

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	mainController "github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	sidecarService "github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

const PaymentMethodWechatPay = "wechatpay_native"

type WechatPayRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

func getWechatPayMinTopup() int64 {
	minTopup := setting.WechatPayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}

func getWechatPayMoney(amount float64, group string) int64 {
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

func RequestWechatPay(c *gin.Context) {
	if !setting.WechatPayEnabled {
		common.ApiErrorI18n(c, i18n.MsgPaymentWechatNotEnabled)
		return
	}

	var req WechatPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	minTopup := getWechatPayMinTopup()
	if req.Amount < minTopup {
		common.ApiErrorI18n(c, i18n.MsgPaymentAmountBelowMin, gin.H{"Min": fmt.Sprintf("%d", minTopup)})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPaymentGetUserGroupError)
		return
	}

	payMoneyFen := getWechatPayMoney(float64(req.Amount), group)
	if payMoneyFen < 1 {
		common.ApiErrorI18n(c, i18n.MsgPaymentAmountTooLow)
		return
	}

	// Token 模式下归一化 Amount
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = int64(float64(req.Amount) / common.QuotaPerUnit)
		if amount < 1 {
			amount = 1
		}
	}

	tradeNo := fmt.Sprintf("WXP-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))

	// 创建本地订单
	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         float64(payMoneyFen) / 100,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodWechatPay,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		log.Printf("微信支付创建本地订单失败: %v", err)
		common.ApiErrorI18n(c, i18n.MsgPaymentCreateFailed)
		return
	}

	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/wechatpay/notify"

	codeUrl, err := sidecarService.CreateNativeOrder(tradeNo, payMoneyFen, "AI额度充值", notifyUrl)
	if err != nil {
		log.Printf("微信支付统一下单失败: %v", err)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		common.ApiErrorI18n(c, i18n.MsgPaymentStartFailed)
		return
	}

	log.Printf("微信支付订单创建成功 - 用户: %d, 订单: %s, 金额: %.2f 元", id, tradeNo, float64(payMoneyFen)/100)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"code_url":  codeUrl,
			"trade_no":  tradeNo,
			"amount":    float64(payMoneyFen) / 100,
		},
	})
}

func WechatPayNotify(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("微信支付回调读取 body 失败: %v", err)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(false)))
		return
	}

	var notify sidecarService.WechatPayNotify
	if err := xml.Unmarshal(bodyBytes, &notify); err != nil {
		log.Printf("微信支付回调解析 XML 失败: %v", err)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(false)))
		return
	}

	if notify.ReturnCode != "SUCCESS" {
		log.Printf("微信支付回调返回失败: %s", notify.ReturnMsg)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(false)))
		return
	}

	if !sidecarService.VerifyWechatPayNotify(&notify) {
		log.Printf("微信支付回调签名验证失败: %s", notify.OutTradeNo)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(false)))
		return
	}

	if notify.ResultCode != "SUCCESS" {
		log.Printf("微信支付回调业务失败: %s - %s", notify.ErrCode, notify.OutTradeNo)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(true)))
		return
	}

	tradeNo := notify.OutTradeNo
	mainController.LockOrder(tradeNo)
	defer mainController.UnlockOrder(tradeNo)

	if err := model.RechargeWechatPay(tradeNo); err != nil {
		log.Printf("微信支付充值处理失败: %v, 订单: %s", err, tradeNo)
		c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(false)))
		return
	}

	log.Printf("微信支付充值成功 - 订单: %s", tradeNo)
	c.Data(http.StatusOK, "application/xml", []byte(sidecarService.BuildWechatPayNotifyResponse(true)))
}
