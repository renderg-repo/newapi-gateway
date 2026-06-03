package controller

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	mainController "github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	sidecarService "github.com/QuantumNous/new-api/sidecar/service"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

const PaymentMethodAlipay = "alipay_page"

type AlipayRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

func getAlipayMinTopup() float64 {
	minTopup := setting.AlipayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		return float64(minTopup) * common.QuotaPerUnit
	}
	return float64(minTopup)
}

func getAlipayMoney(amount float64, group string) float64 {
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

func RequestAlipay(c *gin.Context) {
	if !setting.AlipayEnabled {
		common.ApiErrorI18n(c, i18n.MsgPaymentAlipayNotEnabled)
		return
	}

	var req AlipayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	minTopup := getAlipayMinTopup()
	if req.Amount < minTopup-0.0001 {
		common.ApiErrorI18n(c, i18n.MsgPaymentAmountBelowMin, gin.H{"Min": fmt.Sprintf("%.2f", minTopup)})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgPaymentGetUserGroupError)
		return
	}

	payMoney := getAlipayMoney(req.Amount, group)
	if payMoney < 0.01 {
		common.ApiErrorI18n(c, i18n.MsgPaymentAmountTooLow)
		return
	}

	// Token 模式下归一化 Amount
	var amount int64
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amountFloat := req.Amount / common.QuotaPerUnit
		if amountFloat < 1 {
			amountFloat = 1
		}
		amount = int64(math.Round(amountFloat))
	} else {
		amount = int64(math.Round(req.Amount))
	}

	tradeNo := fmt.Sprintf("ALI-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))

	// 创建本地订单
	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodAlipay,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		log.Printf("支付宝创建本地订单失败: %v", err)
		common.ApiErrorI18n(c, i18n.MsgPaymentCreateFailed)
		return
	}

	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/alipay/notify"
	returnUrl := system_setting.ServerAddress + "/dashboard/billing"

	payUrl, err := sidecarService.CreatePageOrder(tradeNo, payMoney, "AI额度充值", returnUrl, notifyUrl)
	if err != nil {
		log.Printf("支付宝创建订单失败: %v", err)
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		common.ApiErrorI18n(c, i18n.MsgPaymentStartFailed)
		return
	}

	log.Printf("支付宝订单创建成功 - 用户: %d, 订单: %s, 金额: %.2f 元", id, tradeNo, payMoney)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_url":  payUrl,
			"trade_no": tradeNo,
			"amount":   payMoney,
		},
	})
}

func AlipayNotify(c *gin.Context) {
	var params map[string]string
	if c.Request.Method == "POST" {
		if err := c.Request.ParseForm(); err != nil {
			_, _ = c.Writer.Write([]byte("fail"))
			return
		}
		params = make(map[string]string)
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	} else {
		params = make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	if len(params) == 0 {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	urlValues := c.Request.PostForm
	if c.Request.Method != "POST" {
		urlValues = c.Request.URL.Query()
	}

	valid, err := sidecarService.VerifyAlipayNotify(urlValues)
	if err != nil || !valid {
		log.Printf("支付宝回调签名验证失败: %v", err)
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}

	tradeNo := params["out_trade_no"]
	if tradeNo == "" {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	mainController.LockOrder(tradeNo)
	defer mainController.UnlockOrder(tradeNo)

	if err := model.RechargeAlipay(tradeNo); err != nil {
		log.Printf("支付宝充值处理失败: %v, 订单: %s", err, tradeNo)
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	log.Printf("支付宝充值成功 - 订单: %s", tradeNo)
	_, _ = c.Writer.Write([]byte("success"))
}

func AlipayReturn(c *gin.Context) {
	// 支付宝同步返回，仅做页面跳转
	c.Redirect(http.StatusFound, system_setting.ServerAddress+"/dashboard/billing")
}
