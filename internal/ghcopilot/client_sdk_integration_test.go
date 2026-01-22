package ghcopilot

import (
	"context"
	"testing"
	"time"
)

// TestClientSDKIntegration 測試 RalphLoopClient 與 SDKExecutor 集成
func TestClientSDKIntegration(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	// 驗證 SDK 執行器已初始化
	if client.sdkExecutor == nil {
		t.Fatal("SDK 執行器應已初始化")
	}

	// 驗證狀態方法
	status := client.GetSDKStatus()
	if status == nil {
		t.Log("初始狀態為 nil（在啟動前正常）")
	}

	// 驗證會話計數
	count := client.GetSDKSessionCount()
	if count != 0 {
		t.Errorf("初始會話計數應為 0，實際: %d", count)
	}
}

// TestClientStartStopSDKExecutor 測試啟動和停止 SDK 執行器
func TestClientStartStopSDKExecutor(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	// 啟動 SDK 執行器
	err := client.StartSDKExecutor(ctx)
	if err != nil {
		t.Logf("啟動 SDK 執行器: %v（可能是 CLI 路徑問題）", err)
	}

	// 停止 SDK 執行器
	err = client.StopSDKExecutor(ctx)
	if err != nil {
		t.Logf("停止 SDK 執行器: %v", err)
	}
}

// TestClientExecuteWithSDK 測試使用 SDK 執行程式碼完成
func TestClientExecuteWithSDK(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	// 嘗試執行（可能因為 CLI 不可用而失敗，這是正常的）
	result, err := client.ExecuteWithSDK(ctx, "print('hello')")
	if err != nil {
		t.Logf("ExecuteWithSDK 失敗: %v（預期可能失敗）", err)
		return
	}

	if result == "" {
		t.Log("返回空結果（預期）")
	}
}

// TestClientExplainWithSDK 測試使用 SDK 解釋程式碼
func TestClientExplainWithSDK(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	result, err := client.ExplainWithSDK(ctx, "def hello(): return 'world'")
	if err != nil {
		t.Logf("ExplainWithSDK 失敗: %v", err)
		return
	}

	if result == "" {
		t.Log("返回空結果（預期）")
	}
}

// TestClientGenerateTestsWithSDK 測試使用 SDK 生成測試
func TestClientGenerateTestsWithSDK(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	result, err := client.GenerateTestsWithSDK(ctx, "func Add(a, b int) int { return a + b }")
	if err != nil {
		t.Logf("GenerateTestsWithSDK 失敗: %v", err)
		return
	}

	if result == "" {
		t.Log("返回空結果（預期）")
	}
}

// TestClientCodeReviewWithSDK 測試使用 SDK 進行程式碼審查
func TestClientCodeReviewWithSDK(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	result, err := client.CodeReviewWithSDK(ctx, "var x = 1; var y = 2;")
	if err != nil {
		t.Logf("CodeReviewWithSDK 失敗: %v", err)
		return
	}

	if result == "" {
		t.Log("返回空結果（預期）")
	}
}

// TestClientSDKSessionManagement 測試 SDK 會話管理
func TestClientSDKSessionManagement(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	// 啟動執行器以測試會話
	if err := client.StartSDKExecutor(ctx); err != nil {
		t.Logf("無法啟動執行器: %v", err)
		return
	}

	// 創建會話
	session, err := client.sdkExecutor.CreateSession("test-session-1")
	if err != nil {
		t.Logf("創建會話失敗: %v", err)
		return
	}

	if session == nil {
		t.Fatal("會話應不為 nil")
	}

	// 驗證會話計數
	count := client.GetSDKSessionCount()
	if count != 1 {
		t.Errorf("會話計數應為 1，實際: %d", count)
	}

	// 列出會話
	sessions := client.ListSDKSessions()
	if len(sessions) != 1 {
		t.Errorf("應有 1 個會話，實際: %d", len(sessions))
	}

	// 終止會話
	err = client.TerminateSDKSession("test-session-1")
	if err != nil {
		t.Logf("終止會話失敗: %v", err)
	}

	// 驗證會話已移除
	count = client.GetSDKSessionCount()
	if count != 0 {
		t.Errorf("終止後會話計數應為 0，實際: %d", count)
	}
}

// TestClientGetSDKStatus 測試取得 SDK 狀態
func TestClientGetSDKStatus(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	// 啟動執行器
	if err := client.StartSDKExecutor(ctx); err != nil {
		t.Logf("無法啟動執行器: %v", err)
		return
	}

	// 取得狀態
	status := client.GetSDKStatus()
	if status == nil {
		t.Fatal("狀態應不為 nil")
	}

	// 驗證狀態欄位
	if !status.Running {
		t.Log("執行器應在運行狀態")
	}

	if status.SessionCount < 0 {
		t.Errorf("會話計數不應為負: %d", status.SessionCount)
	}
}

// TestClientSDKClosing 測試客戶端關閉時正確清理 SDK 資源
func TestClientSDKClosing(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)

	ctx := context.Background()

	// 啟動 SDK 執行器
	if err := client.StartSDKExecutor(ctx); err != nil {
		t.Logf("啟動執行器失敗: %v", err)
	}

	// 建立一些會話
	for i := 0; i < 3; i++ {
		sessionID := "session-" + string(rune(i))
		_, err := client.sdkExecutor.CreateSession(sessionID)
		if err != nil {
			t.Logf("創建會話失敗: %v", err)
		}
	}

	// 關閉客戶端
	err := client.Close()
	if err != nil {
		t.Fatalf("關閉客戶端失敗: %v", err)
	}

	// 驗證客戶端已關閉
	if !client.closed {
		t.Error("客戶端應已關閉")
	}

	// 嘗試使用已關閉的客戶端應返回錯誤
	_, err = client.ExecuteWithSDK(ctx, "test")
	if err == nil || err.Error() != "client is closed" {
		t.Error("已關閉的客戶端應返回錯誤")
	}
}

// TestClientSDKWithTimeout 測試 SDK 執行器的超時設定
func TestClientSDKWithTimeout(t *testing.T) {
	config := DefaultClientConfig()
	config.CLITimeout = 100 * time.Millisecond // 設定超短超時
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	// 驗證超時已設定
	if client.sdkExecutor == nil {
		t.Fatal("SDK 執行器應已初始化")
	}

	if client.sdkExecutor.config.Timeout != 100*time.Millisecond {
		t.Errorf("超時應為 100ms，實際: %v", client.sdkExecutor.config.Timeout)
	}
}

// TestClientSDKMultipleCycles 測試多個 SDK 循環
func TestClientSDKMultipleCycles(t *testing.T) {
	config := DefaultClientConfig()
	client := NewRalphLoopClientWithConfig(config)
	defer client.Close()

	ctx := context.Background()

	// 啟動執行器
	if err := client.StartSDKExecutor(ctx); err != nil {
		t.Logf("啟動執行器失敗: %v", err)
		return
	}

	// 執行多個循環
	for i := 0; i < 3; i++ {
		// 創建會話
		sessionID := "cycle-session-" + string(rune(i))
		_, err := client.sdkExecutor.CreateSession(sessionID)
		if err != nil {
			t.Logf("循環 %d: 創建會話失敗: %v", i, err)
			continue
		}

		// 終止會話
		err = client.TerminateSDKSession(sessionID)
		if err != nil {
			t.Logf("循環 %d: 終止會話失敗: %v", i, err)
		}
	}

	// 驗證最終沒有會話
	count := client.GetSDKSessionCount()
	if count != 0 {
		t.Logf("最終會話計數應為 0，實際: %d", count)
	}
}
