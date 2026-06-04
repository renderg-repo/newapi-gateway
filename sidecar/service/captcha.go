package service

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/mojocn/base64Captcha"
)

const (
	// CaptchaRedisPrefix Redis 中验证码 key 的前缀
	CaptchaRedisPrefix = "captcha:"
	// CaptchaExpiration 验证码有效期
	CaptchaExpiration = 5 * time.Minute
)

// 使用 base64Captcha 内置的 store 来处理
type captchaStore struct{}

var (
	memStore = base64Captcha.DefaultMemStore
)

type CaptchaResponse struct {
	CaptchaId string `json:"captcha_id"`
	ImageData string `json:"image_data"`
}

func GenerateCaptcha() (*CaptchaResponse, error) {
	// 创建数字验证码驱动
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, memStore)

	id, b64s, answer, err := c.Generate()
	if err != nil {
		return nil, err
	}

	// 如果 Redis 启用，也在 Redis 中存一份
	if common.RedisEnabled {
		key := CaptchaRedisPrefix + id
		_ = common.RedisSet(key, answer, CaptchaExpiration)
	}

	return &CaptchaResponse{
		CaptchaId: id,
		ImageData: b64s,
	}, nil
}

func VerifyCaptcha(captchaId, captchaCode string) bool {
	if captchaId == "" || captchaCode == "" {
		return false
	}

	// 先尝试从内存验证（优先，因为更快）
	valid := memStore.Verify(captchaId, captchaCode, true)
	if valid {
		// 如果内存验证成功，也尝试清理 Redis 中的
		if common.RedisEnabled {
			key := CaptchaRedisPrefix + captchaId
			_ = common.RedisDel(key)
		}
		return true
	}

	// 内存验证失败，尝试 Redis
	if common.RedisEnabled {
		key := CaptchaRedisPrefix + captchaId
		storedAnswer, err := common.RedisGet(key)
		if err == nil && storedAnswer != "" {
			if storedAnswer == captchaCode {
				// 验证成功，删除 Redis 中的
				_ = common.RedisDel(key)
				return true
			}
		}
	}

	return false
}

func GenerateCaptchaAuto() (*CaptchaResponse, error) {
	return GenerateCaptcha()
}

func VerifyCaptchaAuto(captchaId, captchaCode string) bool {
	return VerifyCaptcha(captchaId, captchaCode)
}
