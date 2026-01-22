# 階段 8：API 設計與封裝

**完成日期**: 2026-01-21  
**狀態**: 🔄 進行中 (8.1 完成，8.2+ 待進行)

## 概覽

### 目標
設計統一的公開 API (`RalphLoopClient`)，隱藏內部複雜性，提供簡單易用的接口。

### 8.1 成果 (✅ 完成)
- RalphLoopClient 主公開 API 類別
- ClientConfig 與 ClientBuilder 模式
- 15+ 核心公開方法
- 16 個單元測試，100% 通過

---

## RalphLoopClient 設計

### 核心結構
```go
type RalphLoopClient struct {
    // 內部依賴
    executor        *CLIExecutor
    parser          *OutputParser
    analyzer        *ResponseAnalyzer
    breaker         *CircuitBreaker
    contextManager  *ContextManager
    persistenceM    *PersistenceManager
    exitDetector    *ExitDetector
    
    // 狀態管理
    initialized bool
    closed      bool
    config      *ClientConfig
}
```

### ClientConfig 配置
```go
type ClientConfig struct {
    WorkDir           string           // 工作目錄
    Timeout           time.Duration    // CLI 執行逾時
    MaxRetries        int             // 最大重試次數
    CircuitBreakerMax int             // 熔斷器閾值
    MaxContextHistory int             // 上下文歷史上限
    PersistenceDir    string          // 持久化儲存目錄
    Silent            bool            // 靜默模式
    AllowAllTools     bool            // 允許所有工具
    Model             string          // LLM 模型
}
```

### ClientBuilder 流式建構器
```go
// 使用範例
client := NewRalphLoopClient().
    WithWorkDir("/my/project").
    WithTimeout(30 * time.Second).
    WithModel("claude-3.5-sonnet").
    Build()
```

---

## 公開 API 方法清單

### 核心執行方法

#### 1. ExecuteLoop
```go
func (c *RalphLoopClient) ExecuteLoop(
    ctx context.Context, 
    prompt string
) (*LoopResult, error)
```
執行單次迴圈，傳回完整結果或錯誤。

**回傳結構**:
```go
type LoopResult struct {
    LoopIndex          int
    Success            bool
    CleanedOutput      string
    CodeBlocks         []string
    Options            map[string]string
    CompletionScore    float64
    ShouldContinue     bool
    CircuitBreakerOpen bool
    Error              error
}
```

#### 2. ExecuteUntilCompletion
```go
func (c *RalphLoopClient) ExecuteUntilCompletion(
    ctx context.Context, 
    initialPrompt string, 
    maxLoops int
) ([]*LoopResult, error)
```
執行多個迴圈直到完成或達到最大迴圈數。

---

### 狀態查詢方法

#### 3. GetStatus
```go
func (c *RalphLoopClient) GetStatus() *ClientStatus
```
取得目前客戶端狀態。

```go
type ClientStatus struct {
    Initialized         bool
    Closed              bool
    CircuitBreakerOpen  bool
    CircuitBreakerState CircuitBreakerState
    LoopsExecuted       int
    Summary             map[string]interface{}
}
```

#### 4. GetHistory
```go
func (c *RalphLoopClient) GetHistory() []*ExecutionContext
```
取得所有已執行迴圈的歷史記錄（讀取專用複本）。

#### 5. GetSummary
```go
func (c *RalphLoopClient) GetSummary() map[string]interface{}
```
取得統計摘要：總迴圈數、成功/失敗計數等。

---

### 控制和管理方法

#### 6. ResetCircuitBreaker
```go
func (c *RalphLoopClient) ResetCircuitBreaker() error
```
重置熔斷器狀態（允許繼續執行）。

#### 7. ClearHistory
```go
func (c *RalphLoopClient) ClearHistory()
```
清空執行歷史記錄。

#### 8. ExportHistory
```go
func (c *RalphLoopClient) ExportHistory(outputPath string) error
```
將完整歷史記錄匯出至文件（JSON 格式）。

#### 9. Close
```go
func (c *RalphLoopClient) Close() error
```
優雅關閉客戶端，清理資源。

---

### 建構器方法

#### 10-15. WithXxx 方法
```go
func (cb *ClientBuilder) WithWorkDir(dir string) *ClientBuilder
func (cb *ClientBuilder) WithTimeout(d time.Duration) *ClientBuilder
func (cb *ClientBuilder) WithMaxRetries(n int) *ClientBuilder
func (cb *ClientBuilder) WithMaxContextHistory(n int) *ClientBuilder
func (cb *ClientBuilder) WithCircuitBreakerMax(n int) *ClientBuilder
func (cb *ClientBuilder) WithModel(m string) *ClientBuilder
func (cb *ClientBuilder) WithSilent(s bool) *ClientBuilder
func (cb *ClientBuilder) WithAllowAllTools(b bool) *ClientBuilder
func (cb *ClientBuilder) Build() *RalphLoopClient
```

---

## 設計原則

### 1. 單一職責原則
- `RalphLoopClient` 作為統一入口點，隱藏內部模組複雜性
- 各內部模組保持獨立的職責

### 2. 封裝
- 所有內部細節通過 `ExecutionContext` 透明公開
- 不暴露低層組件，只公開高層業務對象

### 3. 流式配置
- `ClientBuilder` 模式提供流式、可讀的配置方法
- 預設值合理，允許部分自定義

### 4. 錯誤處理
- 所有 API 方法返回 `(result, error)` 元組
- 系統級錯誤通過 error 返回，不拋出異常

### 5. 生命週期管理
- `Close()` 明確指定清理時機
- 防止資源洩漏

---

## 測試覆蓋 (✅ 16/16 通過)

### 建構測試
- ✅ TestNewRalphLoopClient
- ✅ TestDefaultClientConfig
- ✅ TestClientBuilderPattern

### 狀態查詢測試
- ✅ TestGetStatus
- ✅ TestGetHistory
- ✅ TestClientGetSummary
- ✅ TestClientConfiguration

### 控制測試
- ✅ TestClearHistory
- ✅ TestClientClose
- ✅ TestResetCircuitBreaker
- ✅ TestGetStatus_CircuitBreakerOpen
- ✅ TestExecuteLoopWithoutInit
- ✅ TestExecuteLoopAfterClose

### 構建者測試
- ✅ TestBuilderMultipleSettings

---

## 後續階段規劃

### 🔄 階段 8.2：模組整合（待進行）
**目標**: 完全集成 ContextManager 與 PersistenceManager

**任務**:
- [ ] 自動持久化上下文（每個迴圈後）
- [ ] 載入歷史記錄於初始化
- [ ] 配置持久化位置
- [ ] 備份機制

### 🔄 階段 8.3：錯誤處理與重試（待進行）
**目標**: 完善錯誤處理、重試邏輯

**任務**:
- [ ] 實作重試邏輯
- [ ] 錯誤分類與恢復策略
- [ ] 登錄與診斷
- [ ] 優雅降級

### 🔄 階段 8.4：完整執行迴圈（待進行）
**目標**: 實作完整迴圈工作流

**任務**:
- [ ] 使用者交互流程
- [ ] 決策邏輯
- [ ] 退出條件整合
- [ ] 性能最佳化

---

## 程式碼統計

| 檔案 | 行數 | 描述 |
|-----|------|------|
| client.go | 330 | 主 API 實作 |
| client_test.go | 267 | 單元測試 |
| **總計** | **597** | **API 層** |

---

## 技術債務

目前已記錄的技術債務（需要在未來版本中解決）:
- Context 結構與 SDK 的潛在冗餘
- SDK 版本過舊，需要遷移至新版
- 詳見 [TECHNICAL_DEBT.md](./TECHNICAL_DEBT.md)

---

## 相關文件
- [IMPLEMENTATION_PROGRESS.md](./IMPLEMENTATION_PROGRESS.md) - 整體進度
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 系統架構
- [context.go](./internal/ghcopilot/context.go) - 上下文管理
- [persistence.go](./internal/ghcopilot/persistence.go) - 持久化
