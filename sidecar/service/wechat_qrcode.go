package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// QRCodeStatus represents the status of a QR code login session.
type QRCodeStatus string

const (
	QRCodeStatusPending   QRCodeStatus = "pending"
	QRCodeStatusScanned   QRCodeStatus = "scanned"
	QRCodeStatusConfirmed QRCodeStatus = "confirmed"
	QRCodeStatusExpired   QRCodeStatus = "expired"
	QRCodeStatusNeedBind  QRCodeStatus = "need_bind"
)

const (
	qrcodeTTL           = 5 * time.Minute
	cleanupInterval     = 1 * time.Minute
	redisTicketPrefix   = "wechat:qrcode:ticket:"
	redisScenePrefix    = "wechat:qrcode:scene:"
)

// QRCodeSession holds the state of a single QR code login attempt.
type QRCodeSession struct {
	Ticket     string        `json:"ticket"`
	Status     QRCodeStatus  `json:"status"`
	OpenID     string        `json:"openid,omitempty"`
	UserID     int           `json:"user_id,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
}

var (
	ticketStore sync.Map // key: ticket UUID, value: *QRCodeSession
	cleanupOnce sync.Once
)

// ---------- helper functions for Redis fallback ----------

func useRedis() bool {
	return common.RedisEnabled && common.RDB != nil
}

func getRedisTicketKey(ticket string) string {
	return redisTicketPrefix + ticket
}

func saveSessionToRedis(session *QRCodeSession) error {
	if !useRedis() {
		return nil
	}

	ticketKey := getRedisTicketKey(session.Ticket)

	// Use JSON string storage
	data, err := common.Marshal(session)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = qrcodeTTL
	}

	return common.RedisSet(ticketKey, string(data), ttl)
}

func getSessionFromRedisByTicket(ticket string) (*QRCodeSession, bool) {
	if !useRedis() {
		return nil, false
	}

	key := getRedisTicketKey(ticket)
	data, err := common.RedisGet(key)
	if err != nil {
		return nil, false
	}

	var session QRCodeSession
	if err := common.UnmarshalJsonStr(data, &session); err != nil {
		return nil, false
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) && session.Status == QRCodeStatusPending {
		session.Status = QRCodeStatusExpired
		// Update the status in Redis
		_ = saveSessionToRedis(&session)
	}

	return &session, true
}

func deleteSessionFromRedis(session *QRCodeSession) {
	if !useRedis() {
		return
	}

	ticketKey := getRedisTicketKey(session.Ticket)
	_ = common.RedisDel(ticketKey)
}

// ---------- public API ----------

// CreateQRCodeSession creates a new QR code login session by calling HPC system API.
func CreateQRCodeSession() (*QRCodeSession, error) {
	startCleanupRoutine()

	if common.WeChatHpcServerAddress == "" {
		return nil, fmt.Errorf("WeChat HPC server address not configured")
	}

	// Call HPC system API to get QR code
	url := fmt.Sprintf("%s/weixin/getQrCode?type=3", common.WeChatHpcServerAddress)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call HPC API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HPC API response: %w", err)
	}

	// Parse HPC response
	var hpcResp struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
		Message string       `json:"message"`
	}
	if err := common.UnmarshalJsonStr(string(body), &hpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse HPC API response: %w", err)
	}

	if hpcResp.Code != 200 {
		return nil, fmt.Errorf("HPC API error: %s", hpcResp.Message)
	}

	// Parse the data field which contains the WeChat API response
	var wechatResp struct {
		Ticket string `json:"ticket"`
	}
	if err := common.UnmarshalJsonStr(string(hpcResp.Data), &wechatResp); err != nil {
		return nil, fmt.Errorf("failed to parse WeChat response: %w", err)
	}

	if wechatResp.Ticket == "" {
		return nil, fmt.Errorf("no ticket in HPC API response")
	}

	// Create session with ticket from HPC
	now := time.Now()
	session := &QRCodeSession{
		Ticket:    wechatResp.Ticket,
		Status:    QRCodeStatusPending,
		CreatedAt: now,
		ExpiresAt: now.Add(qrcodeTTL),
	}

	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(session.Ticket, session)
	}

	return session, nil
}

// GetQRCodeSessionByTicket retrieves a session by ticket UUID.
func GetQRCodeSessionByTicket(ticket string) (*QRCodeSession, bool) {
	if useRedis() {
		return getSessionFromRedisByTicket(ticket)
	}

	val, ok := ticketStore.Load(ticket)
	if !ok {
		return nil, false
	}
	session := val.(*QRCodeSession)
	if time.Now().After(session.ExpiresAt) && session.Status == QRCodeStatusPending {
		session.Status = QRCodeStatusExpired
		ticketStore.Store(ticket, session)
	}
	return session, true
}

// CheckQRCodeStatus polls HPC system to check QR code status.
func CheckQRCodeStatus(ticket string) (*QRCodeSession, error) {
	if common.WeChatHpcServerAddress == "" {
		return nil, fmt.Errorf("WeChat HPC server address not configured")
	}

	// Call HPC system API to check QR code status
	url := fmt.Sprintf("%s/weixin/checkQrCode", common.WeChatHpcServerAddress)
	reqBody := map[string]string{"ticket": ticket}
	jsonBody, err := common.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call HPC API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HPC API response: %w", err)
	}

	// Parse HPC response
	var hpcResp struct {
		Code int    `json:"code"`
		Data struct {
			Username string `json:"username"`
			Token    string `json:"token"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := common.UnmarshalJsonStr(string(body), &hpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse HPC API response: %w", err)
	}

	// Get current session
	session, ok := GetQRCodeSessionByTicket(ticket)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	// Update session status based on HPC response
	switch hpcResp.Code {
	case 200:
		// Success - already logged in, mark as confirmed
		session.Status = QRCodeStatusConfirmed
		// We don't get openid from HPC, just mark status
		// Username is available in hpcResp.Data.Username if needed
	case 900:
		// Waiting for scan
		session.Status = QRCodeStatusPending
	default:
		// Other codes - treat as expired or failed
		session.Status = QRCodeStatusExpired
	}

	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(ticket, session)
	}

	return session, nil
}

// SaveSession saves session to redis or memory.
func SaveSession(session *QRCodeSession) {
	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(session.Ticket, session)
	}
}

// UpdateSessionByCallback updates session with openid from HPC system callback.
func UpdateSessionByCallback(ticket string, openid string) (bool, error) {
	session, ok := GetQRCodeSessionByTicket(ticket)
	if !ok {
		return false, fmt.Errorf("session not found for ticket: %s", ticket)
	}

	session.OpenID = openid
	session.Status = QRCodeStatusConfirmed

	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(ticket, session)
	}

	return true, nil
}

// ---------- private helpers ----------

func startCleanupRoutine() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				cleanupExpiredSessions()
			}
		}()
	})
}

func cleanupExpiredSessions() {
	now := time.Now()

	if useRedis() {
		// Redis handles expiration automatically, no need to cleanup
		return
	}

	ticketStore.Range(func(key, val any) bool {
		session := val.(*QRCodeSession)
		if now.After(session.ExpiresAt) {
			ticketStore.Delete(key)
		}
		return true
	})
}
