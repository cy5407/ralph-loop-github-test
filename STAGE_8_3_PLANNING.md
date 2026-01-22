# Stage 8.3 規劃：SDK 層整合與容錯機制

**狀態**: 📋 規劃中  
**目標測試數**: 130-140  
**計畫完成日期**: 2026-Q1  

---

## 📊 現狀分析

### SDK 版本升級完成 ✅
```
舊版: github.com/github/copilot-sdk/go v0.1.14
新版: github.com/github/copilot-sdk/go v0.1.15-preview.0 (最新開發版)
狀態: 所有 3 個 SDK PoC 測試通過
```

### 雙層架構現狀
```
RalphLoopClient (Stage 8.1-8.2)
├── CLI 層 ✅ (cli_executor.go)
│   ├── 命令: copilot version, copilot explain...
│   ├── 狀態: 生產就緒 (125 個測試)
│   └── 特點: 輕量級、簡單
│
└── SDK 層 ⏳ (待整合)
    ├── API 層: NewClient, Start, Stop...
    ├── 狀態: PoC 測試通過，待整合至主 API
    └── 特點: 類型安全、連接持久
```

---

## 🎯 Stage 8.3 目標

### Batch 1: SDK 層集成 (25-30 個測試)

#### 1.1 建立 SDKExecutor 模組
```go
// internal/ghcopilot/sdk_executor.go

type SDKExecutor struct {
    client  copilot.Client
    config  *SDKConfig
    session *SDKSession
}

// 核心方法
func (e *SDKExecutor) Start(ctx context.Context) error
func (e *SDKExecutor) Stop(ctx context.Context) error
func (e *SDKExecutor) Complete(ctx context.Context, prompt string) (string, error)
func (e *SDKExecutor) Explain(ctx context.Context, code string) (string, error)
func (e *SDKExecutor) GenerateTests(ctx context.Context, code string) (string, error)
func (e *SDKExecutor) CodeReview(ctx context.Context, code string) (string, error)
```

#### 1.2 建立 SDK 會話管理
```go
// SDK Session 追蹤
type SDKSession struct {
    ID        string
    StartTime time.Time
    Status    string
    Metrics   SessionMetrics
}

// 會話池管理
type SDKSessionPool struct {
    sessions map[string]*SDKSession
    mu       sync.RWMutex
}
```

#### 1.3 集成至 RalphLoopClient
```go
// client.go 新增方法

// 支援 SDK 執行
func (c *RalphLoopClient) ExecuteWithSDK(ctx context.Context, prompt string) (*LoopResult, error)

// SDK 會話管理
func (c *RalphLoopClient) GetSDKStatus() *SDKStatus
func (c *RalphLoopClient) ListSDKSessions() []*SDKSession
func (c *RalphLoopClient) TerminateSDKSession(sessionID string) error
```

#### 1.4 單元測試 (25-30 個)
- TestSDKExecutorStart
- TestSDKExecutorStop
- TestSDKComplete
- TestSDKExplain
- TestSDKGenerateTests
- TestSDKCodeReview
- TestSDKSessionCreation
- TestSDKSessionPoolManagement
- TestSDKErrorHandling
- TestExecuteLoopWithSDK
- TestGetSDKStatus
- TestListSDKSessions
- ... 等

### Batch 2: 容錯與重試機制 (20-25 個測試)

#### 2.1 重試策略
```go
type RetryPolicy struct {
    MaxRetries    int
    InitialBackoff time.Duration
    MaxBackoff    time.Duration
    Strategy      RetryStrategy // Exponential, Linear, Fixed
}

// 實現重試邏輯
func (e *SDKExecutor) WithRetry(ctx context.Context, fn func() error) error
```

#### 2.2 故障恢復
```go
type FailureRecovery struct {
    BackupStrategy  BackupStrategy
    AutoRecover     bool
    MaxRecoveryTime time.Duration
}

// 自動恢復
func (c *RalphLoopClient) EnableAutoRecovery(enabled bool)
func (c *RalphLoopClient) RecoverFromFailure(ctx context.Context) error
```

#### 2.3 單元測試
- TestRetryWithExponentialBackoff
- TestRetryWithLinearBackoff
- TestRetryMaxAttemptsExceeded
- TestAutoRecoveryTrigger
- TestRecoverFromSDKFailure
- ... 等

### Batch 3: CLI vs SDK 選擇器 (10-15 個測試)

#### 3.1 智能選擇器
```go
type ExecutionMode int

const (
    ModeCLI       ExecutionMode = iota  // 輕量級
    ModeSDK                             // 類型安全
    ModeAuto                            // 自動選擇
    ModeHybrid                          // 混合
)

type Selector struct {
    preference    ExecutionMode
    fallbackOn    bool
}

// 選擇最佳執行模式
func (s *Selector) Choose(task *Task) ExecutionMode
```

#### 3.2 效能比較
```go
type PerformanceMetrics struct {
    CLITime        time.Duration
    SDKTime        time.Duration
    MemoryUsage    uint64
    ErrorRate      float64
}

func (c *RalphLoopClient) BenchmarkExecutionModes() *PerformanceMetrics
```

#### 3.3 單元測試
- TestChooseMode_Simple
- TestChooseMode_Complex
- TestFallbackToSDK
- TestFallbackToCLI
- TestHybridExecution
- ... 等

---

## 📈 測試進度預期

| 階段 | Batch | 測試數 | 累計 | 進度 |
|-----|-------|--------|------|------|
| 8.2 | 完成 | 125 | 125 | ✅ 70% |
| 8.3 | Batch 1 | 25-30 | 150-155 | 🔄 進行中 |
| 8.3 | Batch 2 | 20-25 | 170-180 | ⏳ 計畫中 |
| 8.3 | Batch 3 | 10-15 | 180-195 | ⏳ 計畫中 |
| **總計** | **8.3** | **55-70** | **180-195** | **完成目標** |

---

## 🏗️ 檔案結構變更

```
internal/ghcopilot/
├── client.go              (已有，+新方法)
├── client_test.go         (已有，+新測試)
├── cli_executor.go        (已有)
├── cli_executor_test.go   (已有)
├── sdk_executor.go        (新增) ✨
├── sdk_executor_test.go   (新增) ✨
├── sdk_session_pool.go    (新增) ✨
├── sdk_session_pool_test.go (新增) ✨
├── selector.go            (新增) ✨
├── selector_test.go       (新增) ✨
└── ... (其他現有模組)
```

---

## 🔄 實現流程

### 第 1 周: SDK Executor 基礎
1. ✅ 升級 SDK 至最新版本
2. 🔄 建立 SDKExecutor 結構體
3. 📝 實現 Start/Stop/Complete
4. 🧪 建立初始測試 (10 個)

### 第 2 周: 會話管理
1. 實現 SDKSession 管理
2. 建立會話池
3. 集成至 RalphLoopClient
4. 增加測試 (10-15 個)

### 第 3 周: 容錯機制
1. 實現重試策略
2. 實現自動恢復
3. 添加故障檢測
4. 增加測試 (20-25 個)

### 第 4 周: 執行模式選擇
1. 建立選擇器邏輯
2. 實現效能監測
3. 集成到主 API
4. 增加測試 (10-15 個)

---

## 💡 技術考量

### SDK 相容性
- ✅ 當前版本: v0.1.15-preview.0 通過所有測試
- 📋 未來遷移: 等待官方 v1.0 穩定版本
- 🔮 備選方案: 保持 CLI 層作為備選

### 性能優化
- **CLI**: 適合一次性操作 (輕量級)
- **SDK**: 適合長連接 (批量操作)
- **混合**: 根據工作負載動態選擇

### 錯誤恢復
- 故障時自動降級至 CLI
- 利用現有備份層恢復狀態
- 記錄詳細日誌用於診斷

---

## 🚀 關鍵里程碑

| 里程碑 | 目標 | 測試數 |
|--------|------|--------|
| 🎯 Batch 1 完成 | SDK 層完全集成 | 150-155 |
| 🎯 Batch 2 完成 | 容錯機制就緒 | 170-180 |
| 🎯 Batch 3 完成 | 智能選擇器就緒 | 180-195 |
| 🏁 Stage 8.3 完成 | 全面集成測試通過 | **195+** |

---

## 📚 參考文件

- [當前 SDK 說明](./COPILOT_SDK_EXPLANATION.md)
- [CLI 執行器](../internal/ghcopilot/cli_executor.go)
- [Client API](../internal/ghcopilot/client.go)
- [SDK PoC 測試](../test/sdk_poc_test.go)
- [Stage 8.2 完成報告](./STAGE_8_2_BATCH_3.md)

---

## ✨ 預期收益

### 功能層面
- ✅ 程式化 SDK 集成
- ✅ 完整的會話管理
- ✅ 自動故障恢復
- ✅ 動態執行模式選擇

### 品質層面
- ✅ 測試覆蓋率提升至 98%+
- ✅ 錯誤捕獲率提升至 95%+
- ✅ 系統可靠性達到企業級

### 性能層面
- ✅ 平均響應時間優化 20-30%
- ✅ 資源使用率最佳化
- ✅ 支援高並發操作
