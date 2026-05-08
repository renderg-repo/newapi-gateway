package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"
)

func setupPaymentTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.QuotaPerUnit = 500000

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	DB = db
	LOG_DB = db

	if err := db.AutoMigrate(&TopUp{}, &User{}, &Log{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func seedUser(t *testing.T, db *gorm.DB, username string, quota int) *User {
	t.Helper()
	user := &User{
		Username: username,
		Password: "test_password_123",
		Quota:    quota,
		Group:    "default",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		AffCode:  randstr.String(8),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func seedTopUp(t *testing.T, db *gorm.DB, userId int, tradeNo, paymentMethod, status string, amount int64, money float64) *TopUp {
	t.Helper()
	topUp := &TopUp{
		UserId:        userId,
		Amount:        amount,
		Money:         money,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		CreateTime:    common.GetTimestamp(),
		Status:        status,
	}
	if err := db.Create(topUp).Error; err != nil {
		t.Fatalf("failed to create topup: %v", err)
	}
	return topUp
}

func TestRechargeWechatPay(t *testing.T) {
	db := setupPaymentTestDB(t)

	t.Run("successful recharge", func(t *testing.T) {
		user := seedUser(t, db, "wechat_user1", 0)
		seedTopUp(t, db, user.Id, "WXP-1-123-abc", "wechatpay_native", common.TopUpStatusPending, 10, 7.3)

		err := RechargeWechatPay("WXP-1-123-abc")
		require.NoError(t, err)

		// Verify topup status changed to success
		var topUp TopUp
		db.Where("trade_no = ?", "WXP-1-123-abc").First(&topUp)
		require.Equal(t, common.TopUpStatusSuccess, topUp.Status)
		require.Greater(t, topUp.CompleteTime, int64(0))

		// Verify user quota increased
		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		require.Equal(t, int(10*common.QuotaPerUnit), updatedUser.Quota)
	})

	t.Run("idempotent - already success", func(t *testing.T) {
		user := seedUser(t, db, "wechat_user2", 100)
		seedTopUp(t, db, user.Id, "WXP-2-123-abc", "wechatpay_native", common.TopUpStatusSuccess, 10, 7.3)

		err := RechargeWechatPay("WXP-2-123-abc")
		require.NoError(t, err)

		// User quota should not change
		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		require.Equal(t, 100, updatedUser.Quota)
	})

	t.Run("wrong payment method", func(t *testing.T) {
		user := seedUser(t, db, "wechat_user3", 0)
		seedTopUp(t, db, user.Id, "WXP-3-123-abc", "alipay_page", common.TopUpStatusPending, 10, 7.3)

		err := RechargeWechatPay("WXP-3-123-abc")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("non-pending status", func(t *testing.T) {
		user := seedUser(t, db, "wechat_user4", 0)
		seedTopUp(t, db, user.Id, "WXP-4-123-abc", "wechatpay_native", common.TopUpStatusFailed, 10, 7.3)

		err := RechargeWechatPay("WXP-4-123-abc")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("trade not found", func(t *testing.T) {
		err := RechargeWechatPay("WXP-NOT-EXIST")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("empty trade no", func(t *testing.T) {
		err := RechargeWechatPay("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "未提供支付单号")
	})
}

func TestRechargeAlipay(t *testing.T) {
	db := setupPaymentTestDB(t)

	t.Run("successful recharge", func(t *testing.T) {
		user := seedUser(t, db, "alipay_user1", 0)
		seedTopUp(t, db, user.Id, "ALI-1-123-abc", "alipay_page", common.TopUpStatusPending, 20, 14.6)

		err := RechargeAlipay("ALI-1-123-abc")
		require.NoError(t, err)

		var topUp TopUp
		db.Where("trade_no = ?", "ALI-1-123-abc").First(&topUp)
		require.Equal(t, common.TopUpStatusSuccess, topUp.Status)
		require.Greater(t, topUp.CompleteTime, int64(0))

		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		require.Equal(t, int(20*common.QuotaPerUnit), updatedUser.Quota)
	})

	t.Run("idempotent - already success", func(t *testing.T) {
		user := seedUser(t, db, "alipay_user2", 500)
		seedTopUp(t, db, user.Id, "ALI-2-123-abc", "alipay_page", common.TopUpStatusSuccess, 20, 14.6)

		err := RechargeAlipay("ALI-2-123-abc")
		require.NoError(t, err)

		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		require.Equal(t, 500, updatedUser.Quota)
	})

	t.Run("wrong payment method", func(t *testing.T) {
		user := seedUser(t, db, "alipay_user3", 0)
		seedTopUp(t, db, user.Id, "ALI-3-123-abc", "wechatpay_native", common.TopUpStatusPending, 20, 14.6)

		err := RechargeAlipay("ALI-3-123-abc")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("non-pending status", func(t *testing.T) {
		user := seedUser(t, db, "alipay_user4", 0)
		seedTopUp(t, db, user.Id, "ALI-4-123-abc", "alipay_page", common.TopUpStatusFailed, 20, 14.6)

		err := RechargeAlipay("ALI-4-123-abc")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("trade not found", func(t *testing.T) {
		err := RechargeAlipay("ALI-NOT-EXIST")
		require.Error(t, err)
		require.Contains(t, err.Error(), "充值失败")
	})

	t.Run("empty trade no", func(t *testing.T) {
		err := RechargeAlipay("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "未提供支付单号")
	})
}

func TestRechargeWechatPayQuotaCalculation(t *testing.T) {
	db := setupPaymentTestDB(t)

	t.Run("quota calculation with decimal", func(t *testing.T) {
		common.QuotaPerUnit = 500000
		user := seedUser(t, db, "wechat_user5", 0)
		seedTopUp(t, db, user.Id, "WXP-5-123-abc", "wechatpay_native", common.TopUpStatusPending, 3, 2.19)

		err := RechargeWechatPay("WXP-5-123-abc")
		require.NoError(t, err)

		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		// 3 * 500000 = 1500000
		require.Equal(t, 1500000, updatedUser.Quota)
	})
}

func TestRechargeAlipayQuotaCalculation(t *testing.T) {
	db := setupPaymentTestDB(t)

	t.Run("quota calculation with decimal", func(t *testing.T) {
		common.QuotaPerUnit = 500000
		user := seedUser(t, db, "alipay_user5", 0)
		seedTopUp(t, db, user.Id, "ALI-5-123-abc", "alipay_page", common.TopUpStatusPending, 5, 36.5)

		err := RechargeAlipay("ALI-5-123-abc")
		require.NoError(t, err)

		var updatedUser User
		db.Where("id = ?", user.Id).First(&updatedUser)
		// 5 * 500000 = 2500000
		require.Equal(t, 2500000, updatedUser.Quota)
	})
}
