package ghcopilot

import (
	"fmt"
	"sync"
	"time"
)

// SDKSession 代表一個 SDK 會話
type SDKSession struct {
	ID         string            // 會話 ID
	StartTime  time.Time         // 開始時間
	Status     SDKSessionStatus  // 狀態
	LastUsed   time.Time         // 最後使用時間
	Metrics    *SessionMetrics   // 會話指標
	Properties map[string]string // 自訂屬性
}

// SDKSessionStatus 會話狀態
type SDKSessionStatus string

const (
	SessionActive   SDKSessionStatus = "active"
	SessionIdle     SDKSessionStatus = "idle"
	SessionClosed   SDKSessionStatus = "closed"
	SessionError    SDKSessionStatus = "error"
	SessionRecovery SDKSessionStatus = "recovery"
)

// SessionMetrics 會話指標
type SessionMetrics struct {
	TotalCalls      int64         // 總呼叫次數
	SuccessfulCalls int64         // 成功呼叫次數
	FailedCalls     int64         // 失敗呼叫次數
	TotalDuration   time.Duration // 總執行時間
	AverageDuration time.Duration // 平均執行時間
	ErrorCount      int64         // 錯誤計數
	LastError       error         // 最後一個錯誤
	LastErrorTime   time.Time     // 最後錯誤時間
}

// SDKSessionPool 會話池管理
type SDKSessionPool struct {
	mu       sync.RWMutex
	sessions map[string]*SDKSession
	maxSize  int
	timeout  time.Duration // 會話逾時
}

// NewSDKSessionPool 建立新的會話池
func NewSDKSessionPool(maxSize int, timeout time.Duration) *SDKSessionPool {
	return &SDKSessionPool{
		sessions: make(map[string]*SDKSession),
		maxSize:  maxSize,
		timeout:  timeout,
	}
}

// CreateSession 建立新會話
func (p *SDKSessionPool) CreateSession(sessionID string) (*SDKSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 檢查是否已達最大大小
	if len(p.sessions) >= p.maxSize {
		return nil, fmt.Errorf("session pool full")
	}

	// 檢查會話是否已存在
	if _, exists := p.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists")
	}

	session := &SDKSession{
		ID:         sessionID,
		StartTime:  time.Now(),
		Status:     SessionActive,
		LastUsed:   time.Now(),
		Metrics:    &SessionMetrics{},
		Properties: make(map[string]string),
	}

	p.sessions[sessionID] = session
	return session, nil
}

// GetSession 取得會話
func (p *SDKSessionPool) GetSession(sessionID string) (*SDKSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 檢查是否逾時
	if time.Since(session.LastUsed) > p.timeout && session.Status == SessionActive {
		return nil, fmt.Errorf("session timeout")
	}

	return session, nil
}

// UpdateSession 更新會話
func (p *SDKSessionPool) UpdateSession(sessionID string, updateFn func(*SDKSession) error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	if err := updateFn(session); err != nil {
		return err
	}

	session.LastUsed = time.Now()
	return nil
}

// RemoveSession 移除會話
func (p *SDKSessionPool) RemoveSession(sessionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Status = SessionClosed
	delete(p.sessions, sessionID)
	return nil
}

// ListSessions 列出所有會話
func (p *SDKSessionPool) ListSessions() []*SDKSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sessions := make([]*SDKSession, 0, len(p.sessions))
	for _, session := range p.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetSessionCount 取得會話計數
func (p *SDKSessionPool) GetSessionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

// CleanupExpiredSessions 清理過期會話
func (p *SDKSessionPool) CleanupExpiredSessions() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	count := 0

	for sessionID, session := range p.sessions {
		if session.Status == SessionActive && now.Sub(session.LastUsed) > p.timeout {
			session.Status = SessionClosed
			delete(p.sessions, sessionID)
			count++
		}
	}

	return count
}

// ClearAll 清除所有會話
func (p *SDKSessionPool) ClearAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, session := range p.sessions {
		session.Status = SessionClosed
	}

	p.sessions = make(map[string]*SDKSession)
	return nil
}

// RecordCall 記錄會話呼叫
func (m *SessionMetrics) RecordCall(duration time.Duration, success bool, err error) {
	m.TotalCalls++
	m.TotalDuration += duration

	if m.TotalCalls > 0 {
		m.AverageDuration = m.TotalDuration / time.Duration(m.TotalCalls)
	}

	if success {
		m.SuccessfulCalls++
	} else {
		m.FailedCalls++
		m.ErrorCount++
		m.LastError = err
		m.LastErrorTime = time.Now()
	}
}

// GetErrorRate 取得錯誤率
func (m *SessionMetrics) GetErrorRate() float64 {
	if m.TotalCalls == 0 {
		return 0.0
	}
	return float64(m.FailedCalls) / float64(m.TotalCalls)
}

// GetSuccessRate 取得成功率
func (m *SessionMetrics) GetSuccessRate() float64 {
	if m.TotalCalls == 0 {
		return 0.0
	}
	return float64(m.SuccessfulCalls) / float64(m.TotalCalls)
}
