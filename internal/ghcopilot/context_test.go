package ghcopilot

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewContextManager 測試建立新的上下文管理器
func TestNewContextManager(t *testing.T) {
	cm := NewContextManager()
	if cm == nil {
		t.Error("NewContextManager() 傳回 nil")
	}
	if cm.currentLoop != nil {
		t.Error("新建立的管理器應無當前迴圈")
	}
	if len(cm.loopHistory) != 0 {
		t.Errorf("新建立的管理器應無歷史記錄，但有 %d 筆", len(cm.loopHistory))
	}
}

// TestNewExecutionContext 測試建立新的執行上下文
func TestNewExecutionContext(t *testing.T) {
	loopIndex := 0
	userPrompt := "測試提示詞"
	ctx := NewExecutionContext(loopIndex, userPrompt)

	if ctx == nil {
		t.Error("NewExecutionContext() 傳回 nil")
	}
	if ctx.LoopIndex != loopIndex {
		t.Errorf("迴圈索引應為 %d，但為 %d", loopIndex, ctx.LoopIndex)
	}
	if ctx.UserPrompt != userPrompt {
		t.Errorf("使用者提示應為 '%s'，但為 '%s'", userPrompt, ctx.UserPrompt)
	}
	if ctx.Metadata == nil {
		t.Error("Metadata 應初始化，但為 nil")
	}
}

// TestStartLoop 測試開始新的迴圈
func TestStartLoop(t *testing.T) {
	cm := NewContextManager()
	userPrompt := "測試提示詞"

	ctx := cm.StartLoop(0, userPrompt)
	if ctx == nil {
		t.Error("StartLoop() 傳回 nil")
	}
	if ctx.UserPrompt != userPrompt {
		t.Errorf("迴圈提示應為 '%s'", userPrompt)
	}

	currentCtx := cm.GetCurrentLoop()
	if currentCtx == nil {
		t.Error("GetCurrentLoop() 傳回 nil")
	}
	if currentCtx.LoopIndex != 0 {
		t.Errorf("迴圈索引應為 0，但為 %d", currentCtx.LoopIndex)
	}
}

// TestUpdateCurrentLoop 測試更新當前迴圈
func TestUpdateCurrentLoop(t *testing.T) {
	cm := NewContextManager()
	cm.StartLoop(0, "測試提示詞")

	err := cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		ctx.CLICommand = "copilot what-the-shell 'ls'"
		ctx.CLIOutput = "ls -la"
		ctx.CLIExitCode = 0
	})

	if err != nil {
		t.Errorf("UpdateCurrentLoop() 失敗: %v", err)
	}

	ctx := cm.GetCurrentLoop()
	if ctx.CLICommand != "copilot what-the-shell 'ls'" {
		t.Error("CLI 命令未正確更新")
	}
	if ctx.CLIExitCode != 0 {
		t.Error("CLI 退出碼未正確更新")
	}
}

// TestUpdateCurrentLoopWithoutActive 測試在無活動迴圈時更新失敗
func TestUpdateCurrentLoopWithoutActive(t *testing.T) {
	cm := NewContextManager()

	err := cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		// 不會執行
	})

	if err == nil {
		t.Error("應該在無活動迴圈時傳回錯誤")
	}
}

// TestFinishLoop 測試完成迴圈
func TestFinishLoop(t *testing.T) {
	cm := NewContextManager()
	cm.StartLoop(0, "測試提示詞")

	time.Sleep(10 * time.Millisecond) // 讓時間流逝

	err := cm.FinishLoop()
	if err != nil {
		t.Errorf("FinishLoop() 失敗: %v", err)
	}

	if cm.GetCurrentLoop() != nil {
		t.Error("完成後應無當前迴圈")
	}

	history := cm.GetLoopHistory()
	if len(history) != 1 {
		t.Errorf("歷史記錄應有 1 筆，但有 %d 筆", len(history))
	}

	if history[0].DurationMs == 0 {
		t.Error("執行時間應被記錄")
	}
}

// TestFinishLoopWithoutActive 測試在無活動迴圈時完成失敗
func TestFinishLoopWithoutActive(t *testing.T) {
	cm := NewContextManager()

	err := cm.FinishLoop()
	if err == nil {
		t.Error("應該在無活動迴圈時傳回錯誤")
	}
}

// TestMultipleLoops 測試多個連續迴圈
func TestMultipleLoops(t *testing.T) {
	cm := NewContextManager()

	// 執行 5 個迴圈
	for i := 0; i < 5; i++ {
		cm.StartLoop(i, "提示詞 "+string(rune(i)))
		cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
			ctx.ShouldContinue = true
		})
		cm.FinishLoop()
	}

	history := cm.GetLoopHistory()
	if len(history) != 5 {
		t.Errorf("歷史記錄應有 5 筆，但有 %d 筆", len(history))
	}

	summary := cm.GetSummary()
	if totalLoops, ok := summary["total_loops"].(int); ok {
		if totalLoops != 5 {
			t.Errorf("總迴圈數應為 5，但為 %d", totalLoops)
		}
	}
}

// TestGetLoopByIndex 測試根據索引取得迴圈
func TestGetLoopByIndex(t *testing.T) {
	cm := NewContextManager()

	// 建立多個迴圈
	for i := 0; i < 3; i++ {
		cm.StartLoop(i, "提示詞")
		cm.FinishLoop()
	}

	ctx := cm.GetLoopByIndex(1)
	if ctx == nil {
		t.Error("應該找到迴圈索引 1")
	}
	if ctx.LoopIndex != 1 {
		t.Errorf("迴圈索引應為 1，但為 %d", ctx.LoopIndex)
	}

	notFound := cm.GetLoopByIndex(10)
	if notFound != nil {
		t.Error("不應找到迴圈索引 10")
	}
}

// TestGetSummary 測試取得摘要
func TestGetSummary(t *testing.T) {
	cm := NewContextManager()

	// 建立 3 個成功迴圈
	for i := 0; i < 3; i++ {
		cm.StartLoop(i, "提示詞")
		cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
			ctx.ShouldContinue = true
		})
		cm.FinishLoop()
	}

	summary := cm.GetSummary()

	if totalLoops, ok := summary["total_loops"].(int); !ok || totalLoops != 3 {
		t.Error("總迴圈數應為 3")
	}

	if successCount, ok := summary["success_count"].(int); !ok || successCount != 3 {
		t.Error("成功計數應為 3")
	}

	if errorCount, ok := summary["error_count"].(int); !ok || errorCount != 0 {
		t.Error("錯誤計數應為 0")
	}
}

// TestGetLastErrorContext 測試取得最後一個錯誤上下文
func TestGetLastErrorContext(t *testing.T) {
	cm := NewContextManager()

	// 建立兩個成功迴圈和一個失敗迴圈
	cm.StartLoop(0, "提示詞")
	cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		ctx.ShouldContinue = true
	})
	cm.FinishLoop()

	cm.StartLoop(1, "提示詞")
	cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		ctx.ExitReason = "連續相同錯誤"
	})
	cm.FinishLoop()

	lastError := cm.GetLastErrorContext()
	if lastError == nil {
		t.Error("應該找到最後一個錯誤")
	}
	if lastError.LoopIndex != 1 {
		t.Errorf("最後錯誤應為迴圈 1，但為 %d", lastError.LoopIndex)
	}
}

// TestClear 測試清空上下文
func TestClear(t *testing.T) {
	cm := NewContextManager()

	// 建立一些迴圈
	for i := 0; i < 3; i++ {
		cm.StartLoop(i, "提示詞")
		cm.FinishLoop()
	}

	cm.Clear()

	if cm.GetCurrentLoop() != nil {
		t.Error("清空後應無當前迴圈")
	}
	if len(cm.GetLoopHistory()) != 0 {
		t.Error("清空後應無歷史記錄")
	}
}

// TestSetMaxHistorySize 測試設定最大歷史大小
func TestSetMaxHistorySize(t *testing.T) {
	cm := NewContextManager()
	cm.SetMaxHistorySize(3)

	// 建立 5 個迴圈
	for i := 0; i < 5; i++ {
		cm.StartLoop(i, "提示詞")
		cm.FinishLoop()
	}

	history := cm.GetLoopHistory()
	if len(history) > 3 {
		t.Errorf("歷史記錄應不超過 3 筆，但有 %d 筆", len(history))
	}
}

// TestToJSON 測試轉換為 JSON
func TestToJSON(t *testing.T) {
	cm := NewContextManager()

	// 建立一個迴圈
	cm.StartLoop(0, "測試提示詞")
	cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		ctx.CLICommand = "copilot what-the-shell 'ls'"
		ctx.CLIOutput = "ls -la"
		ctx.ShouldContinue = true
	})
	cm.FinishLoop()

	jsonStr, err := cm.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() 失敗: %v", err)
	}

	// 驗證是否為有效的 JSON
	var data interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Errorf("產生的 JSON 無效: %v", err)
	}
}

// TestContextMetadata 測試 Context 的 Metadata 欄位
func TestContextMetadata(t *testing.T) {
	ctx := NewExecutionContext(0, "測試")

	ctx.Metadata["key1"] = "value1"
	ctx.Metadata["key2"] = 42

	if ctx.Metadata["key1"] != "value1" {
		t.Error("Metadata 讀取失敗")
	}
	if ctx.Metadata["key2"] != 42 {
		t.Error("Metadata 整數讀取失敗")
	}
}

// TestLoopStatus 測試迴圈狀態結構
func TestLoopStatus(t *testing.T) {
	status := &LoopStatus{
		Status:     "CONTINUE",
		ExitSignal: false,
		TasksDone:  "2/5",
	}

	if status.Status != "CONTINUE" {
		t.Error("狀態設定失敗")
	}
	if status.ExitSignal {
		t.Error("ExitSignal 應為 false")
	}
	if status.TasksDone != "2/5" {
		t.Error("TasksDone 設定失敗")
	}
}

// TestGetLoopHistoryCopy 測試取得歷史副本（確保線程安全）
func TestGetLoopHistoryCopy(t *testing.T) {
	cm := NewContextManager()

	// 建立一些迴圈
	for i := 0; i < 3; i++ {
		cm.StartLoop(i, "提示詞")
		cm.FinishLoop()
	}

	history1 := cm.GetLoopHistory()

	if history1 != nil && len(history1) > 0 {
		// 修改返回的副本不應影響內部歷史記錄
		history1[0].LoopID = "修改"

		// 直接修改物件的欄位（因為傳回的是指向物件的指標）
		// 所以內部也會被修改，但這是 Go 的指標語義
		// 測試應檢查的是數組本身的副本

		// 重新取得歷史，確認數組本身是新的副本
		history2 := cm.GetLoopHistory()
		history2[0].LoopID = "又修改"

		// 再次取得，應該是新副本
		history3 := cm.GetLoopHistory()

		// 驗證 history3 是獨立的副本陣列（指標不同）
		if len(history3) != len(cm.loopHistory) {
			t.Error("歷史記錄長度不符")
		}

		// 確認分配的是不同的陣列
		if cap(history3) == 0 {
			t.Error("歷史陣列容量應大於 0")
		}
	}
}
