# Ralph Loop 開發計畫

## 階段 1: 專案設定與依賴檢查

- [ ] 建立 `internal/ghcopilot/` 目錄結構
- [ ] 初始化 Go 模組 (`go mod init github.com/cy540/ralph-loop`)
- [ ] 實作依賴檢查功能 (`dependency_checker.go`)
- [ ] 撰寫依賴檢查單元測試

## 階段 2: CLI 執行器核心

- [ ] 實作基本 CLI 執行器 (`cli_executor.go`)
- [ ] 實作 `SuggestShellCommand()` 方法
- [ ] 實作錯誤處理與重試機制
- [ ] 撰寫 CLI 執行器單元測試

## 階段 3: 輸出解析器

- [ ] 建立 `CopilotSuggestion` 資料結構
- [ ] 實作 Markdown 程式碼區塊提取
- [ ] 實作多選項識別
- [ ] 撰寫輸出解析器單元測試

## 階段 4: 回應分析器

- [ ] 實作結構化輸出解析 (`---COPILOT_STATUS---`)
- [ ] 實作自然語言完成偵測
- [ ] 實作信心分數計算
- [ ] 撰寫回應分析器單元測試

## 階段 5: 熔斷器

- [ ] 實作熔斷器狀態管理 (CLOSED/HALF_OPEN/OPEN)
- [ ] 實作無進展偵測
- [ ] 實作相同錯誤偵測
- [ ] 撰寫熔斷器單元測試

## 階段 6: 退出偵測

- [ ] 實作雙重條件退出驗證
- [ ] 實作退出訊號追蹤
- [ ] 撰寫退出偵測單元測試

## 里程碑驗收

- [ ] `go build ./...` 成功
- [ ] `go test ./...` 全部通過
- [ ] `go vet ./...` 無警告
