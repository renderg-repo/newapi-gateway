package service

import (
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
)

const (
	qrcodeTicketTTL      = 5 * time.Minute
	qrcodeCleanupInterval = 1 * time.Minute
)

// QRCodeSession holds the state of a single QR code login attempt.
type QRCodeSession struct {
	Ticket    string       `json:"ticket"`
	Status    QRCodeStatus `json:"status"`
	WeChatId  string       `json:"wechat_id,omitempty"`
	UserID    int          `json:"user_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

var (
	ticketStore sync.Map
	cleanupOnce sync.Once
)

// CreateQRCodeSession generates a new QR code login session and returns it.
func CreateQRCodeSession() *QRCodeSession {
	startCleanupRoutine()

	ticket := common.GetUUID()
	now := time.Now()
	session := &QRCodeSession{
		Ticket:    ticket,
		Status:    QRCodeStatusPending,
		CreatedAt: now,
		ExpiresAt: now.Add(qrcodeTicketTTL),
	}
	ticketStore.Store(ticket, session)
	return session
}

// GetQRCodeSession retrieves a session by ticket.
func GetQRCodeSession(ticket string) (*QRCodeSession, bool) {
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

// UpdateQRCodeSessionStatus updates the status of a session and optionally sets related fields.
func UpdateQRCodeSessionStatus(ticket string, status QRCodeStatus, wechatId string, userID int) bool {
	session, ok := GetQRCodeSession(ticket)
	if !ok {
		return false
	}
	session.Status = status
	if wechatId != "" {
		session.WeChatId = wechatId
	}
	if userID > 0 {
		session.UserID = userID
	}
	ticketStore.Store(ticket, session)
	return true
}

// startCleanupRoutine starts a background goroutine that periodically removes expired sessions.
func startCleanupRoutine() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(qrcodeCleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				cleanupExpiredSessions()
			}
		}()
	})
}

func cleanupExpiredSessions() {
	ticketStore.Range(func(key, val any) bool {
		session := val.(*QRCodeSession)
		if time.Now().After(session.ExpiresAt) {
			ticketStore.Delete(key)
		}
		return true
	})
}
