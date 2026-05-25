package service

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/QuantumNous/new-api/types"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
)

const (
	HookEventRelayFinished = "hook.relay.finished"
	HookEventTopUpSuccess  = "hook.topup.success"
)

// RelayHookPayload 模型调用完成事件的负载
type RelayHookPayload struct {
	Event         string `json:"event"`
	UserId        int    `json:"user_id"`
	TokenId       int    `json:"token_id"`
	ChannelId     int    `json:"channel_id"`
	OriginModel   string `json:"origin_model"`
	RequestPath   string `json:"request_path"`
	IsStream      bool   `json:"is_stream"`
	StatusCode    int    `json:"status_code"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	QuotaConsumed int    `json:"quota_consumed"`
	DurationMs    int64  `json:"duration_ms"`
	Timestamp     int64  `json:"timestamp"`
}

// TopUpHookPayload 用户充值成功事件的负载
type TopUpHookPayload struct {
	Event         string  `json:"event"`
	UserId        int     `json:"user_id"`
	TradeNo       string  `json:"trade_no"`
	PaymentMethod string  `json:"payment_method"`
	Amount        int64   `json:"amount"`
	Money         float64 `json:"money"`
	Quota         int     `json:"quota"`
	Timestamp     int64   `json:"timestamp"`
}

func init() {
	model.OnTopUpSuccess = EmitTopUpSuccessHook
}

// EmitRelayFinishedHook 触发模型调用完成钩子（fire-and-forget）
func EmitRelayFinishedHook(c *gin.Context, relayInfo *relaycommon.RelayInfo, err *types.NewAPIError) {
	if !isHookEnabled() {
		return
	}
	if relayInfo == nil {
		return
	}
	payload := buildRelayHookPayload(c, relayInfo, err)
	goSendHook(payload)
}

// EmitTopUpSuccessHook 触发充值成功钩子（fire-and-forget）
func EmitTopUpSuccessHook(topUp *model.TopUp, quota int) {
	if !isHookEnabled() {
		return
	}
	if topUp == nil {
		return
	}
	payload := TopUpHookPayload{
		Event:         HookEventTopUpSuccess,
		UserId:        topUp.UserId,
		TradeNo:       topUp.TradeNo,
		PaymentMethod: topUp.PaymentMethod,
		Amount:        topUp.Amount,
		Money:         topUp.Money,
		Quota:         quota,
		Timestamp:     time.Now().Unix(),
	}
	goSendHook(payload)
}

func isHookEnabled() bool {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return common.OptionMap["HookEnabled"] == "true" && common.OptionMap["HookUrl"] != ""
}

func getHookUrlAndSecret() (url, secret string) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return common.OptionMap["HookUrl"], common.OptionMap["HookSecret"]
}

func buildRelayHookPayload(c *gin.Context, relayInfo *relaycommon.RelayInfo, err *types.NewAPIError) RelayHookPayload {
	durationMs := time.Since(relayInfo.StartTime).Milliseconds()
	statusCode := http.StatusOK
	errorCode := ""
	errorMessage := ""
	if err != nil {
		statusCode = err.StatusCode
		errorCode = string(err.GetErrorCode())
		errorMessage = err.Error()
	}
	channelId := 0
	if relayInfo.ChannelMeta != nil {
		channelId = relayInfo.ChannelMeta.ChannelId
	}
	return RelayHookPayload{
		Event:         HookEventRelayFinished,
		UserId:        relayInfo.UserId,
		TokenId:       relayInfo.TokenId,
		ChannelId:     channelId,
		OriginModel:   relayInfo.OriginModelName,
		RequestPath:   relayInfo.RequestURLPath,
		IsStream:      relayInfo.IsStream,
		StatusCode:    statusCode,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
		QuotaConsumed: relayInfo.FinalPreConsumedQuota,
		DurationMs:    durationMs,
		Timestamp:     time.Now().Unix(),
	}
}

func goSendHook(payload interface{}) {
	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				common.SysError(fmt.Sprintf("hook delivery panic: %v", r))
			}
		}()
		if err := sendHookEvent(payload); err != nil {
			common.SysError("hook delivery failed: " + err.Error())
		}
	})
}

func sendHookEvent(payload interface{}) error {
	hookUrl, secret := getHookUrlAndSecret()
	if hookUrl == "" {
		return nil
	}

	payloadBytes, err := common.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal hook payload: %v", err)
	}

	var resp *http.Response
	if system_setting.EnableWorker() {
		workerReq := &WorkerRequest{
			URL:    hookUrl,
			Key:    system_setting.WorkerValidKey,
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: payloadBytes,
		}
		if secret != "" {
			signature := generateSignature(secret, payloadBytes)
			workerReq.Headers["X-Hook-Signature"] = signature
		}
		resp, err = DoWorkerRequest(workerReq)
		if err != nil {
			return fmt.Errorf("failed to send hook through worker: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("hook request failed with status code: %d", resp.StatusCode)
		}
	} else {
		fetchSetting := system_setting.GetFetchSetting()
		if err := common.ValidateURLWithFetchSetting(hookUrl, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
			return fmt.Errorf("hook request rejected by ssrf protection: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, hookUrl, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create hook request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if secret != "" {
			signature := generateSignature(secret, payloadBytes)
			req.Header.Set("X-Hook-Signature", signature)
		}

		client := GetHttpClient()
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send hook request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("hook request failed with status code: %d", resp.StatusCode)
		}
	}
	return nil
}
