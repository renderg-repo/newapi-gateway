package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

type SMSProvider interface {
	Send(phone string, code string) error
}

var (
	sendRecord sync.Map // phone -> lastSendTime
)

func getProvider() SMSProvider {
	if !common.SMSEnabled {
		return nil
	}
	switch common.SMSProvider {
	case "aliyun":
		return &AliyunSMSProvider{}
	case "tencent":
		return &TencentSMSProvider{}
	case "twilio":
		return &TwilioSMSProvider{}
	default:
		return nil
	}
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

	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessSecret)
	if err != nil {
		return fmt.Errorf("aliyun sms client init failed: %w", err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = signName
	request.TemplateCode = templateCode

	templateVar := common.OptionMap["SmsTemplateVar"]
	if templateVar == "" {
		templateVar = "code"
	}
	request.TemplateParam = fmt.Sprintf(`{"%s":"%s"}`, templateVar, code)

	response, err := client.SendSms(request)
	if err != nil {
		return fmt.Errorf("aliyun sms send failed: %w", err)
	}
	if response.Code != "OK" {
		return fmt.Errorf("aliyun sms send error: %s - %s", response.Code, response.Message)
	}

	common.SysLog(fmt.Sprintf("[AliyunSMS] send code %s to %s success, requestId=%s", code, phone, response.RequestId))
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
