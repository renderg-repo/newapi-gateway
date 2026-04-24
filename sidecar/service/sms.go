package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type SMSProvider interface {
	Send(phone string, code string) error
}

var (
	sendRecord   sync.Map // phone -> lastSendTime
	provider     SMSProvider
	providerOnce sync.Once
)

func getProvider() SMSProvider {
	providerOnce.Do(func() {
		if !common.SMSEnabled {
			provider = nil
			return
		}
		switch common.SMSProvider {
		case "aliyun":
			provider = &AliyunSMSProvider{}
		case "tencent":
			provider = &TencentSMSProvider{}
		case "twilio":
			provider = &TwilioSMSProvider{}
		default:
			provider = nil
		}
	})
	return provider
}

func InvalidateProvider() {
	provider = nil
	providerOnce = sync.Once{}
}

func CanSendToPhone(phone string) bool {
	if last, ok := sendRecord.Load(phone); ok {
		if time.Since(last.(time.Time)) < 60*time.Second {
			return false
		}
	}
	return true
}

func SendSMS(phone string, code string) error {
	p := getProvider()
	if p == nil {
		return errors.New("sms provider not configured")
	}
	if err := p.Send(phone, code); err != nil {
		return err
	}
	sendRecord.Store(phone, time.Now())
	return nil
}

// AliyunSMSProvider implements SMSProvider for Alibaba Cloud SMS.
type AliyunSMSProvider struct{}

func (a *AliyunSMSProvider) Send(phone string, code string) error {
	accessKeyId := common.OptionMap["SmsAccessKeyId"]
	accessSecret := common.OptionMap["SmsAccessKeySecret"]
	signName := common.OptionMap["SmsSignName"]
	templateCode := common.OptionMap["SmsTemplateCode"]

	if accessKeyId == "" || accessSecret == "" || signName == "" || templateCode == "" {
		return errors.New("aliyun sms config incomplete")
	}

	// TODO: integrate Alibaba Cloud SMS SDK (github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi)
	common.SysLog(fmt.Sprintf("[AliyunSMS] send code %s to %s (sign=%s, template=%s)", code, phone, signName, templateCode))
	return nil
}

// TencentSMSProvider implements SMSProvider for Tencent Cloud SMS.
type TencentSMSProvider struct{}

func (t *TencentSMSProvider) Send(phone string, code string) error {
	accessKeyId := common.OptionMap["SmsAccessKeyId"]
	accessSecret := common.OptionMap["SmsAccessKeySecret"]
	signName := common.OptionMap["SmsSignName"]
	templateCode := common.OptionMap["SmsTemplateCode"]

	if accessKeyId == "" || accessSecret == "" || signName == "" || templateCode == "" {
		return errors.New("tencent sms config incomplete")
	}

	// TODO: integrate Tencent Cloud SMS SDK (github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms)
	common.SysLog(fmt.Sprintf("[TencentSMS] send code %s to %s (sign=%s, template=%s)", code, phone, signName, templateCode))
	return nil
}

// TwilioSMSProvider implements SMSProvider for Twilio.
type TwilioSMSProvider struct{}

func (t *TwilioSMSProvider) Send(phone string, code string) error {
	accessKeyId := common.OptionMap["SmsAccessKeyId"]
	accessSecret := common.OptionMap["SmsAccessKeySecret"]

	if accessKeyId == "" || accessSecret == "" {
		return errors.New("twilio config incomplete")
	}

	// TODO: integrate Twilio SDK (github.com/twilio/twilio-go)
	common.SysLog(fmt.Sprintf("[TwilioSMS] send code %s to %s", code, phone))
	return nil
}
