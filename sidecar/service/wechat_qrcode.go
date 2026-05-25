package service

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// ---------- XML message types for WeChat event callbacks ----------

// WeChatEncryptedMessage is the top-level XML structure from WeChat.
type WeChatEncryptedMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName  string   `xml:"FromUserName"`
	CreateTime    int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content       string   `xml:"Content"`
	MsgID        int64    `xml:"MsgId"`
	AgentID      int      `xml:"AgentID"`

	// Event fields
	Event     string  `xml:"Event"`
	EventKey  string  `xml:"EventKey"`
	Ticket    string  `xml:"Ticket"`
	Latitude  float64 `xml:"Latitude"`
	Longitude float64 `xml:"Longitude"`
	Precision float64 `xml:"Precision"`
}

// ---------- QRCode session types ----------

// QRCodeStatus represents the status of a QR code login session.
type QRCodeStatus string

const (
	QRCodeStatusPending   QRCodeStatus = "pending"
	QRCodeStatusScanned   QRCodeStatus = "scanned"
	QRCodeStatusConfirmed QRCodeStatus = "confirmed"
	QRCodeStatusExpired   QRCodeStatus = "expired"
)

const (
	qrcodeTTL            = 5 * time.Minute
	cleanupInterval      = 1 * time.Minute
	accessTokenThreshold = 30 * time.Minute // refresh if TTL < 30min
	redisTicketPrefix   = "wechat:qrcode:ticket:"
	redisScenePrefix     = "wechat:qrcode:scene:"
	redisTokenKey       = "wechat:qrcode:token"
)

// QRCodeSession holds the state of a single QR code login attempt.
type QRCodeSession struct {
	Ticket    string       `json:"ticket"`
	SceneID   int          `json:"scene_id"`
	Status    QRCodeStatus `json:"status"`
	OpenID    string       `json:"openid,omitempty"`
	UserID    int          `json:"user_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// WeChatToken holds the cached access_token with expiry.
type WeChatToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

var (
	ticketStore sync.Map // key: ticket UUID, value: *QRCodeSession
	sceneStore  sync.Map // key: scene_id (int), value: *QRCodeSession
	cleanupOnce sync.Once

	accessTokenMu     sync.RWMutex
	cachedToken       atomic.Value // stores *WeChatToken
	sceneIDCounter    atomic.Int64
)

// ---------- helper functions for Redis fallback ----------

func useRedis() bool {
	return common.RedisEnabled && common.RDB != nil
}

func getRedisTicketKey(ticket string) string {
	return redisTicketPrefix + ticket
}

func getRedisSceneKey(sceneID int) string {
	return redisScenePrefix + strconv.Itoa(sceneID)
}

func saveSessionToRedis(session *QRCodeSession) error {
	if !useRedis() {
		return nil
	}

	ticketKey := getRedisTicketKey(session.Ticket)
	sceneKey := getRedisSceneKey(session.SceneID)

	// Use JSON string storage
	data, err := common.Marshal(session)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = qrcodeTTL
	}

	if err := common.RedisSet(ticketKey, string(data), ttl); err != nil {
		return err
	}
	return common.RedisSet(sceneKey, string(data), ttl)
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

func getSessionFromRedisBySceneID(sceneID int) (*QRCodeSession, bool) {
	if !useRedis() {
		return nil, false
	}

	key := getRedisSceneKey(sceneID)
	data, err := common.RedisGet(key)
	if err != nil {
		return nil, false
	}

	var session QRCodeSession
	if err := common.UnmarshalJsonStr(data, &session); err != nil {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) && session.Status == QRCodeStatusPending {
		session.Status = QRCodeStatusExpired
		_ = saveSessionToRedis(&session)
	}

	return &session, true
}

func deleteSessionFromRedis(session *QRCodeSession) {
	if !useRedis() {
		return
	}

	ticketKey := getRedisTicketKey(session.Ticket)
	sceneKey := getRedisSceneKey(session.SceneID)
	_ = common.RedisDel(ticketKey)
	_ = common.RedisDel(sceneKey)
}

func saveTokenToRedis(token *WeChatToken) error {
	if !useRedis() {
		return nil
	}

	data, err := common.Marshal(token)
	if err != nil {
		return err
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Hour // Fallback TTL
	}

	return common.RedisSet(redisTokenKey, string(data), ttl)
}

func getTokenFromRedis() (*WeChatToken, bool) {
	if !useRedis() {
		return nil, false
	}

	data, err := common.RedisGet(redisTokenKey)
	if err != nil {
		return nil, false
	}

	var token WeChatToken
	if err := common.UnmarshalJsonStr(data, &token); err != nil {
		return nil, false
	}

	return &token, true
}

// ---------- public API ----------

// VerifySignature checks the WeChat signature for message callback validation.
func VerifySignature(signature, timestamp, nonce string) bool {
	token := common.WeChatReceiveToken
	if token == "" {
		return false
	}
	parts := []string{token, timestamp, nonce}
	sort.Strings(parts)
	hash := sha1.Sum([]byte(strings.Join(parts, "")))
	expected := fmt.Sprintf("%x", hash)
	return expected == signature
}

// GetAccessToken returns a valid WeChat access_token, refreshing if needed.
func GetAccessToken() (string, error) {
	// First check in-memory cache
	if v := cachedToken.Load(); v != nil {
		t := v.(*WeChatToken)
		if time.Now().Before(t.ExpiresAt) {
			return t.AccessToken, nil
		}
	}

	// Then check Redis
	if token, ok := getTokenFromRedis(); ok {
		if time.Now().Before(token.ExpiresAt) {
			// Refresh in-memory cache
			cachedToken.Store(token)
			return token.AccessToken, nil
		}
	}

	accessTokenMu.Lock()
	defer accessTokenMu.Unlock()

	// Double-check after acquiring lock.
	if v := cachedToken.Load(); v != nil {
		t := v.(*WeChatToken)
		if time.Now().Before(t.ExpiresAt) {
			return t.AccessToken, nil
		}
	}

	appID := common.WeChatAppID
	appSecret := common.WeChatAppSecret
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("WeChat AppID or AppSecret not configured")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to request access_token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read access_token response: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	if err := common.UnmarshalJsonStr(string(body), &result); err != nil {
		return "", fmt.Errorf("failed to parse access_token response: %w", err)
	}
	if result.Errcode != 0 {
		return "", fmt.Errorf("wechat API error: %d %s", result.Errcode, result.Errmsg)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}

	// WeChat tokens expire in 7200s; refresh at 30min remaining.
	ttl := time.Duration(result.ExpiresIn) * time.Second
	if ttl > accessTokenThreshold {
		ttl -= accessTokenThreshold
	}
	token := &WeChatToken{
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(ttl),
	}
	cachedToken.Store(token)

	// Save to Redis
	_ = saveTokenToRedis(token)

	return result.AccessToken, nil
}

// CreateTempQRCode creates a temporary QR code via WeChat API.
// Returns the ticket and the QR code image URL.
func CreateTempQRCode(sceneID int) (ticket string, qrcodeURL string, err error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", accessToken)

	// QR_SCENE: temporary QR code, 600s expiry
	body := map[string]any{
		"expire_seconds": 600,
		"action_name":    "QR_SCENE",
		"action_info": map[string]any{
			"scene": map[string]any{
				"scene_id": sceneID,
			},
		},
	}

	jsonBody, err := common.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", "", fmt.Errorf("failed to create QR code: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read QR code response: %w", err)
	}

	var result struct {
		Ticket        string `json:"ticket"`
		ExpireSeconds int    `json:"expire_seconds"`
		URL           string `json:"url"`
		Errcode       int    `json:"errcode"`
		Errmsg        string `json:"errmsg"`
	}
	if err := common.UnmarshalJsonStr(string(respBody), &result); err != nil {
		return "", "", fmt.Errorf("failed to parse QR code response: %w", err)
	}
	if result.Errcode != 0 {
		return "", "", fmt.Errorf("wechat API error: %d %s", result.Errcode, result.Errmsg)
	}
	if result.Ticket == "" {
		return "", "", fmt.Errorf("empty ticket in response")
	}

	// The QR code image URL.
	imageURL := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", result.Ticket)

	return result.Ticket, imageURL, nil
}

// GenerateSceneID returns a monotonically increasing scene ID (fits int32).
func GenerateSceneID() int {
	return int(sceneIDCounter.Add(1) % 100000)
}

// CreateQRCodeSession generates a new QR code login session with the given scene ID.
func CreateQRCodeSession(sceneID int) *QRCodeSession {
	startCleanupRoutine()

	ticket := common.GetUUID()
	now := time.Now()
	session := &QRCodeSession{
		Ticket:    ticket,
		SceneID:   sceneID,
		Status:    QRCodeStatusPending,
		CreatedAt: now,
		ExpiresAt: now.Add(qrcodeTTL),
	}

	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(ticket, session)
		sceneStore.Store(sceneID, session)
	}
	return session
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
		sceneStore.Store(session.SceneID, session)
	}
	return session, true
}

// GetQRCodeSessionBySceneID retrieves a session by scene_id.
func GetQRCodeSessionBySceneID(sceneID int) (*QRCodeSession, bool) {
	if useRedis() {
		return getSessionFromRedisBySceneID(sceneID)
	}

	val, ok := sceneStore.Load(sceneID)
	if !ok {
		return nil, false
	}
	session := val.(*QRCodeSession)
	if time.Now().After(session.ExpiresAt) && session.Status == QRCodeStatusPending {
		session.Status = QRCodeStatusExpired
		ticketStore.Store(session.Ticket, session)
		sceneStore.Store(sceneID, session)
	}
	return session, true
}

// UpdateQRCodeSessionStatus updates the status of a session.
func UpdateQRCodeSessionStatus(ticket string, status QRCodeStatus, openID string, userID int) bool {
	session, ok := GetQRCodeSessionByTicket(ticket)
	if !ok {
		return false
	}

	session.Status = status
	if openID != "" {
		session.OpenID = openID
	}
	if userID > 0 {
		session.UserID = userID
	}

	if useRedis() {
		_ = saveSessionToRedis(session)
	} else {
		ticketStore.Store(ticket, session)
		sceneStore.Store(session.SceneID, session)
	}

	return true
}

// ParseWeChatMessage parses XML message bytes from WeChat callback.
func ParseWeChatMessage(data []byte) (*WeChatEncryptedMessage, error) {
	var msg WeChatEncryptedMessage
	if err := xml.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return &msg, nil
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
			sceneStore.Delete(session.SceneID)
		}
		return true
	})
}
