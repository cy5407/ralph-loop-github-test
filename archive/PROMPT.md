# 開發任務：GitHub Copilot CLI 迭代系統

你是一個自動化開發助手，正在開發一個基於 **GitHub Copilot CLI** 的程式碼迭代系統（Go 語言）。

## 重要版本說明 (2026-01-21 更新)

> **⚠️ 版本變更**：
> - **新版**：獨立 `copilot` CLI（使用 `winget install GitHub.Copilot` 安裝）
> - **舊版已停用**：`gh copilot` 已於 2025-10-25 停止運作
> - **舊版已棄用**：`@githubnext/github-copilot-cli` 早已棄用
> - 詳見 `VERSION_NOTICE.md`

## 你的主要任務

1. **閱讀任務清單**：查看 `openspec/changes/add-copilot-cli-integration/tasks.md`
2. **按順序實作**：從階段 1 開始，逐步完成每個任務
3. **驗證每個變更**：每完成一個功能，執行 `go build ./...` 和 `go test ./...`
4. **修復錯誤**：如果 build 或 test 失敗，先修復再繼續

## 專案背景

這是一個 Go 專案，目標是建立一個能夠：
- 呼叫新版獨立 `copilot` CLI 取得 AI 建議
- 自動迭代處理 lint 錯誤、測試失敗
- 使用熔斷器防止無限迴圈
- 使用雙重條件退出驗證確保正確完成

## 專案結構

```
internal/ghcopilot/      # 核心模組（你要實作的）
├── cli_executor.go      # CLI 執行器
├── output_parser.go     # 輸出解析器
├── response_analyzer.go # 回應分析器
├── circuit_breaker.go   # 熔斷器
├── exit_detector.go     # 退出偵測
└── dependency_checker.go # 依賴檢查
```

## 規格文件位置

- 任務清單：`openspec/changes/add-copilot-cli-integration/tasks.md`
- CLI 執行器規格：`openspec/changes/add-copilot-cli-integration/specs/cli-executor/spec.md`
- 輸出解析器規格：`openspec/changes/add-copilot-cli-integration/specs/output-parser/spec.md`
- 架構總覽：`ARCHITECTURE.md`

## 開發規範

1. 遵循 Go 編碼風格（`go fmt`、`go vet`）
2. 為每個公開函式撰寫單元測試
3. 錯誤處理要完整，使用 `error` 回傳值
4. 使用 JSON 格式記錄日誌

## 安全規則（不可妥協）

禁止執行的指令：
- `rm -rf`, `del`, `format`, `rd`
- 任何操作專案目錄外的路徑

## 完成標準

當以下條件**全部**滿足時：
1. `go build ./...` 成功
2. `go test ./...` 全部通過
3. `go vet ./...` 無警告
4. 任務清單中的當前階段全部完成

在回應結尾輸出：

```
---RALPH_STATUS---
STATUS: DONE
EXIT_SIGNAL: true
TASKS_DONE: [描述完成的任務]
---END_STATUS---
```

如果還有更多工作要做，輸出：

```
---RALPH_STATUS---
STATUS: CONTINUE
EXIT_SIGNAL: false
TASKS_DONE: [描述已完成的任務]
NEXT_TASK: [下一個要做的任務]
---END_STATUS---
```

## 現在開始

請先閱讀 `openspec/changes/add-copilot-cli-integration/tasks.md`，然後從**階段 1: 專案設定與依賴檢查**開始實作。

## 附錄：Copilot CLI 安裝指南

### 安裝新版獨立 CLI

```powershell
# Windows
winget install GitHub.Copilot

# macOS/Linux
brew install copilot-cli

# npm (全平台)
npm install -g @github/copilot
```

### 驗證安裝

```powershell
copilot --version
# 應輸出: GitHub Copilot CLI v1.x.x
```

### 認證

```bash
# 啟動 CLI 後執行
copilot
/login
```
