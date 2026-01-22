package ghcopilot

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewPersistenceManager 測試建立新的持久化管理器
func TestNewPersistenceManager(t *testing.T) {
	tmpDir := t.TempDir()
	pm, err := NewPersistenceManager(tmpDir, false)

	if err != nil {
		t.Errorf("NewPersistenceManager() 失敗: %v", err)
	}
	if pm == nil {
		t.Error("NewPersistenceManager() 傳回 nil")
	}
	if pm.storageDir != tmpDir {
		t.Errorf("儲存目錄應為 %s，但為 %s", tmpDir, pm.storageDir)
	}
}

// TestSaveContextManagerJSON 測試以 JSON 格式儲存上下文管理器
func TestSaveContextManagerJSON(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	// 建立一個上下文管理器並添加迴圈
	cm := NewContextManager()
	cm.StartLoop(0, "測試提示詞")
	cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
		ctx.CLICommand = "copilot what-the-shell 'ls'"
		ctx.CLIOutput = "ls -la"
		ctx.ShouldContinue = true
	})
	cm.FinishLoop()

	err := pm.SaveContextManager(cm)
	if err != nil {
		t.Errorf("SaveContextManager() 失敗: %v", err)
	}

	// 驗證檔案已建立
	files, _ := pm.ListSavedContexts()
	if len(files) == 0 {
		t.Error("應該建立了儲存檔案")
	}
}

// TestSaveContextManagerGob 測試以 Gob 格式儲存上下文管理器
func TestSaveContextManagerGob(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, true)

	// 建立一個上下文管理器
	cm := NewContextManager()
	cm.StartLoop(0, "測試提示詞")
	cm.FinishLoop()

	err := pm.SaveContextManager(cm)
	if err != nil {
		t.Errorf("SaveContextManager() 失敗: %v", err)
	}

	// 驗證檔案已建立且為 .gob 格式
	files, _ := pm.ListSavedContexts()
	if len(files) == 0 {
		t.Error("應該建立了儲存檔案")
	}

	for _, f := range files {
		if filepath.Ext(f) != ".gob" {
			t.Errorf("檔案應為 .gob 格式，但為 %s", f)
		}
	}
}

// TestSaveExecutionContext 測試儲存單個執行上下文
func TestSaveExecutionContext(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	ctx := NewExecutionContext(0, "測試提示詞")
	ctx.CLICommand = "copilot what-the-shell 'ls'"
	ctx.CLIOutput = "ls -la"

	err := pm.SaveExecutionContext(ctx)
	if err != nil {
		t.Errorf("SaveExecutionContext() 失敗: %v", err)
	}

	// 驗證檔案已建立
	files, _ := pm.ListSavedContexts()
	if len(files) != 1 {
		t.Errorf("應該有 1 個檔案，但有 %d 個", len(files))
	}
}

// TestLoadExecutionContext 測試載入執行上下文
func TestLoadExecutionContext(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	// 儲存一個上下文
	originalCtx := NewExecutionContext(0, "測試提示詞")
	originalCtx.CLICommand = "copilot what-the-shell 'ls'"
	originalCtx.CLIOutput = "ls -la"
	originalCtx.CLIExitCode = 0

	pm.SaveExecutionContext(originalCtx)

	// 載入上下文
	loadedCtx, err := pm.LoadExecutionContext(originalCtx.LoopID)
	if err != nil {
		t.Errorf("LoadExecutionContext() 失敗: %v", err)
	}

	if loadedCtx.CLICommand != originalCtx.CLICommand {
		t.Errorf("CLI 命令不符: %s vs %s", loadedCtx.CLICommand, originalCtx.CLICommand)
	}
	if loadedCtx.CLIExitCode != originalCtx.CLIExitCode {
		t.Errorf("退出碼不符: %d vs %d", loadedCtx.CLIExitCode, originalCtx.CLIExitCode)
	}
}

// TestListSavedContexts 測試列出已儲存的上下文
func TestListSavedContexts(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	// 儲存多個上下文
	for i := 0; i < 3; i++ {
		ctx := NewExecutionContext(i, "測試")
		pm.SaveExecutionContext(ctx)
	}

	files, err := pm.ListSavedContexts()
	if err != nil {
		t.Errorf("ListSavedContexts() 失敗: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("應該有 3 個檔案，但有 %d 個", len(files))
	}
}

// TestExportAsJSON 測試匯出為 JSON
func TestExportAsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	cm := NewContextManager()
	cm.StartLoop(0, "測試")
	cm.FinishLoop()

	exportPath := filepath.Join(tmpDir, "export.json")
	err := pm.ExportAsJSON(cm, exportPath)
	if err != nil {
		t.Errorf("ExportAsJSON() 失敗: %v", err)
	}

	// 驗證檔案已建立
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("匯出檔案未建立")
	}

	// 驗證內容
	data, _ := os.ReadFile(exportPath)
	if len(data) == 0 {
		t.Error("匯出檔案為空")
	}
}

// TestSetMaxBackups 測試設定最多備份數量
func TestSetMaxBackups(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	pm.SetMaxBackups(5)

	if pm.maxBackups != 5 {
		t.Errorf("最多備份數應為 5，但為 %d", pm.maxBackups)
	}
}

// TestGetStorageDir 測試取得儲存目錄
func TestGetStorageDir(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	if pm.GetStorageDir() != tmpDir {
		t.Errorf("儲存目錄應為 %s，但為 %s", tmpDir, pm.GetStorageDir())
	}
}

// TestPersistenceManagerWithNilContext 測試以 nil 上下文保存失敗
func TestPersistenceManagerWithNilContext(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	err := pm.SaveContextManager(nil)
	if err == nil {
		t.Error("應該在保存 nil 上下文時傳回錯誤")
	}
}

// TestPersistenceRoundTrip 測試完整的往返保存和載入
func TestPersistenceRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPersistenceManager(tmpDir, false)

	// 建立上下文
	cm := NewContextManager()
	for i := 0; i < 2; i++ {
		cm.StartLoop(i, "測試提示詞 "+string(rune(i)))
		cm.UpdateCurrentLoop(func(ctx *ExecutionContext) {
			ctx.CLICommand = "test command"
			ctx.CLIOutput = "test output"
			ctx.CompletionScore = 50 + i*10
		})
		cm.FinishLoop()
	}

	// 儲存
	pm.SaveContextManager(cm)

	// 驗證可以列出檔案
	files, _ := pm.ListSavedContexts()
	if len(files) == 0 {
		t.Error("應該有已儲存的檔案")
	}

	// 載入最新的檔案
	if len(files) > 0 {
		filename := filepath.Join(tmpDir, files[len(files)-1])
		loadedCM, err := pm.LoadContextManager(filename)
		if err != nil {
			t.Errorf("LoadContextManager() 失敗: %v", err)
		}

		loadedHistory := loadedCM.GetLoopHistory()
		originalHistory := cm.GetLoopHistory()

		if len(loadedHistory) != len(originalHistory) {
			t.Errorf("載入的歷史長度應為 %d，但為 %d", len(originalHistory), len(loadedHistory))
		}
	}
}

// TestMultipleFormatSupport 測試多種格式支援
func TestMultipleFormatSupport(t *testing.T) {
	tmpDir := t.TempDir()

	// 使用 JSON
	pmJSON, _ := NewPersistenceManager(tmpDir+"/json", false)
	cm := NewContextManager()
	cm.StartLoop(0, "測試")
	cm.FinishLoop()
	pmJSON.SaveContextManager(cm)

	// 使用 Gob
	pmGob, _ := NewPersistenceManager(tmpDir+"/gob", true)
	pmGob.SaveContextManager(cm)

	// 驗證兩種格式都已建立
	filesJSON, _ := pmJSON.ListSavedContexts()
	filesGob, _ := pmGob.ListSavedContexts()

	if len(filesJSON) == 0 {
		t.Error("應該建立 JSON 檔案")
	}
	if len(filesGob) == 0 {
		t.Error("應該建立 Gob 檔案")
	}
}
