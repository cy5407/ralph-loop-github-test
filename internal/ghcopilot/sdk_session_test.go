package ghcopilot

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestNewSDKSessionPool 測試建立新會話池
func TestNewSDKSessionPool(t *testing.T) {
	pool := NewSDKSessionPool(100, 5*time.Minute)

	if pool == nil {
		t.Error("NewSDKSessionPool() 傳回 nil")
	}

	if pool.GetSessionCount() != 0 {
		t.Error("新建立的會話池應該是空的")
	}
}

// TestCreateSessionSuccess 測試成功建立會話
func TestCreateSessionSuccess(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	session, err := pool.CreateSession("test-session")

	if err != nil {
		t.Errorf("建立會話失敗: %v", err)
	}

	if session == nil {
		t.Error("傳回的會話為 nil")
	}

	if session.ID != "test-session" {
		t.Errorf("會話 ID 不符: 期望 'test-session'，實際 %s", session.ID)
	}

	if session.Status != SessionActive {
		t.Errorf("新會話狀態應為 Active，但為 %v", session.Status)
	}

	if pool.GetSessionCount() != 1 {
		t.Error("會話計數應為 1")
	}
}

// TestGetSessionSuccess 測試成功取得會話
func TestGetSessionSuccess(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	created, _ := pool.CreateSession("test-session")
	retrieved, err := pool.GetSession("test-session")

	if err != nil {
		t.Errorf("取得會話失敗: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Error("取得的會話不符")
	}
}

// TestGetSessionNotFound 測試取得不存在的會話
func TestGetSessionNotFound(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	_, err := pool.GetSession("nonexistent")

	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Error("應該傳回 'not found' 錯誤")
	}
}

// TestUpdateSession 測試更新會話
func TestUpdateSession(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	pool.CreateSession("test-session")

	err := pool.UpdateSession("test-session", func(s *SDKSession) error {
		s.Properties["key"] = "value"
		return nil
	})

	if err != nil {
		t.Errorf("更新會話失敗: %v", err)
	}

	session, _ := pool.GetSession("test-session")
	if session.Properties["key"] != "value" {
		t.Error("會話屬性未更新")
	}
}

// TestRemoveSession 測試移除會話
func TestRemoveSession(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	pool.CreateSession("test-session")

	if pool.GetSessionCount() != 1 {
		t.Error("應該有 1 個會話")
	}

	err := pool.RemoveSession("test-session")

	if err != nil {
		t.Errorf("移除會話失敗: %v", err)
	}

	if pool.GetSessionCount() != 0 {
		t.Error("應該沒有會話")
	}

	// 嘗試再次移除應該失敗
	err = pool.RemoveSession("test-session")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Error("再次移除應該傳回錯誤")
	}
}

// TestListSessions 測試列出會話
func TestListSessions(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	pool.CreateSession("session1")
	pool.CreateSession("session2")
	pool.CreateSession("session3")

	sessions := pool.ListSessions()

	if len(sessions) != 3 {
		t.Errorf("應該有 3 個會話，但有 %d 個", len(sessions))
	}

	// 驗證會話 ID
	ids := make(map[string]bool)
	for _, s := range sessions {
		ids[s.ID] = true
	}

	if !ids["session1"] || !ids["session2"] || !ids["session3"] {
		t.Error("會話 ID 不完整")
	}
}

// TestPoolMaxSize 測試池大小限制
func TestPoolMaxSize(t *testing.T) {
	pool := NewSDKSessionPool(3, 5*time.Minute)

	pool.CreateSession("session1")
	pool.CreateSession("session2")
	pool.CreateSession("session3")

	// 嘗試建立第四個應該失敗
	_, err := pool.CreateSession("session4")

	if err == nil || !strings.Contains(err.Error(), "full") {
		t.Error("應該傳回池滿錯誤")
	}
}

// TestPoolDuplicateSession 測試重複會話
func TestPoolDuplicateSession(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	pool.CreateSession("test-session")

	// 嘗試建立相同 ID 的會話應該失敗
	_, err := pool.CreateSession("test-session")

	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Error("應該傳回會話已存在錯誤")
	}
}

// TestSessionTimeout 測試會話逾時
func TestSessionTimeout(t *testing.T) {
	pool := NewSDKSessionPool(10, 100*time.Millisecond)

	pool.CreateSession("test-session")

	// 立即取得應該成功
	_, err := pool.GetSession("test-session")
	if err != nil {
		t.Errorf("立即取得應該成功: %v", err)
	}

	// 等待超過逾時
	time.Sleep(150 * time.Millisecond)

	// 再次取得應該失敗
	_, err = pool.GetSession("test-session")
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Error("應該傳回逾時錯誤")
	}
}

// TestCleanupExpiredSessions 測試清理過期會話
func TestCleanupExpiredSessions(t *testing.T) {
	pool := NewSDKSessionPool(10, 50*time.Millisecond)

	pool.CreateSession("session1")
	pool.CreateSession("session2")
	pool.CreateSession("session3")

	if pool.GetSessionCount() != 3 {
		t.Error("應該有 3 個會話")
	}

	// 等待過期
	time.Sleep(100 * time.Millisecond)

	// 清理
	cleaned := pool.CleanupExpiredSessions()

	if cleaned != 3 {
		t.Errorf("應該清理 3 個會話，但清理了 %d 個", cleaned)
	}

	if pool.GetSessionCount() != 0 {
		t.Error("應該沒有會話")
	}
}

// TestClearAll 測試清除所有會話
func TestClearAll(t *testing.T) {
	pool := NewSDKSessionPool(10, 5*time.Minute)

	pool.CreateSession("session1")
	pool.CreateSession("session2")
	pool.CreateSession("session3")

	if pool.GetSessionCount() != 3 {
		t.Error("應該有 3 個會話")
	}

	err := pool.ClearAll()

	if err != nil {
		t.Errorf("清除所有會話失敗: %v", err)
	}

	if pool.GetSessionCount() != 0 {
		t.Error("應該沒有會話")
	}
}

// TestSessionMetricsRecordCall 測試記錄呼叫
func TestSessionMetricsRecordCall(t *testing.T) {
	metrics := &SessionMetrics{}

	metrics.RecordCall(100*time.Millisecond, true, nil)

	if metrics.TotalCalls != 1 {
		t.Error("應該有 1 個呼叫")
	}

	if metrics.SuccessfulCalls != 1 {
		t.Error("應該有 1 個成功呼叫")
	}

	if metrics.FailedCalls != 0 {
		t.Error("應該沒有失敗呼叫")
	}
}

// TestSessionMetricsRecordError 測試記錄錯誤
func TestSessionMetricsRecordError(t *testing.T) {
	metrics := &SessionMetrics{}

	testErr := fmt.Errorf("test error")
	metrics.RecordCall(50*time.Millisecond, false, testErr)

	if metrics.TotalCalls != 1 {
		t.Error("應該有 1 個呼叫")
	}

	if metrics.FailedCalls != 1 {
		t.Error("應該有 1 個失敗呼叫")
	}

	if metrics.ErrorCount != 1 {
		t.Error("應該有 1 個錯誤")
	}

	if metrics.LastError == nil {
		t.Error("應該記錄最後的錯誤")
	}
}

// TestSessionMetricsAverageDuration 測試平均執行時間
func TestSessionMetricsAverageDuration(t *testing.T) {
	metrics := &SessionMetrics{}

	metrics.RecordCall(100*time.Millisecond, true, nil)
	metrics.RecordCall(200*time.Millisecond, true, nil)

	if metrics.AverageDuration != 150*time.Millisecond {
		t.Errorf("平均執行時間應為 150ms，但為 %v", metrics.AverageDuration)
	}
}

// TestSessionMetricsSuccessRate 測試成功率
func TestSessionMetricsSuccessRate(t *testing.T) {
	metrics := &SessionMetrics{}

	metrics.RecordCall(100*time.Millisecond, true, nil)
	metrics.RecordCall(100*time.Millisecond, true, nil)
	metrics.RecordCall(100*time.Millisecond, false, nil)

	successRate := metrics.GetSuccessRate()
	expectedRate := float64(2) / float64(3)

	if successRate != expectedRate {
		t.Errorf("成功率應為 %f，但為 %f", expectedRate, successRate)
	}
}

// TestSessionMetricsErrorRate 測試錯誤率
func TestSessionMetricsErrorRate(t *testing.T) {
	metrics := &SessionMetrics{}

	metrics.RecordCall(100*time.Millisecond, true, nil)
	metrics.RecordCall(100*time.Millisecond, true, nil)
	metrics.RecordCall(100*time.Millisecond, false, nil)

	errorRate := metrics.GetErrorRate()
	expectedRate := float64(1) / float64(3)

	if errorRate != expectedRate {
		t.Errorf("錯誤率應為 %f，但為 %f", expectedRate, errorRate)
	}
}

// TestSessionMetricsEmpty 測試空指標
func TestSessionMetricsEmpty(t *testing.T) {
	metrics := &SessionMetrics{}

	if metrics.GetSuccessRate() != 0.0 {
		t.Error("空指標的成功率應為 0")
	}

	if metrics.GetErrorRate() != 0.0 {
		t.Error("空指標的錯誤率應為 0")
	}
}

// TestPoolIntegration 測試池集成
func TestPoolIntegration(t *testing.T) {
	pool := NewSDKSessionPool(100, 1*time.Minute)

	// 建立 10 個會話
	for i := 0; i < 10; i++ {
		_, err := pool.CreateSession(fmt.Sprintf("session-%d", i))
		if err != nil {
			t.Errorf("建立會話失敗: %v", err)
		}
	}

	// 驗證計數
	if pool.GetSessionCount() != 10 {
		t.Errorf("應該有 10 個會話，但有 %d 個", pool.GetSessionCount())
	}

	// 移除一些會話
	for i := 0; i < 5; i++ {
		pool.RemoveSession(fmt.Sprintf("session-%d", i))
	}

	// 驗證計數
	if pool.GetSessionCount() != 5 {
		t.Errorf("應該有 5 個會話，但有 %d 個", pool.GetSessionCount())
	}

	// 列出剩餘會話
	sessions := pool.ListSessions()
	if len(sessions) != 5 {
		t.Errorf("應該列出 5 個會話，但列出了 %d 個", len(sessions))
	}

	// 清除所有會話
	pool.ClearAll()

	if pool.GetSessionCount() != 0 {
		t.Error("應該沒有會話")
	}
}
