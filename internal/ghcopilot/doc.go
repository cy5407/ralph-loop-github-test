// Package ghcopilot 提供 GitHub Copilot CLI 整合層
//
// 版本說明（2026-01-21 更新）：
//   - 本套件使用 **新版獨立** GitHub Copilot CLI (`copilot` 命令)
//   - 安裝方式：`winget install GitHub.Copilot` 或 `npm install -g @github/copilot`
//   - **舊版 `gh copilot` 已於 2025-10-25 停用**
//   - **`@githubnext/github-copilot-cli` 早已棄用**
//   - 詳見 VERSION_NOTICE.md
//
// 該套件封裝了所有與 GitHub Copilot CLI 的互動，包括：
//   - 依賴檢查（copilot CLI、認證）
//   - CLI 執行與結果捕獲（直接呼叫 `copilot` 命令）
//   - 輸出解析
//   - 上下文管理
//   - 回應分析（含完成偵測和卡住偵測）
//   - 熔斷機制（防止失控迴圈）
//   - 優雅退出決策（雙重條件驗證）
//
// 依賴關係：
//
//	應用程式 → ghcopilot → copilot CLI (獨立版) → GitHub Copilot 雲端
package ghcopilot

// Version 是目前套件的版本
const Version = "0.2.0"
