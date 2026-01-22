# Stage 8.3 Batch 2 規劃：容錯機制實現

**文件日期**: 2026-01-22  
**Batch 狀態**: 📋 規劃中  
**目標測試數**: 20-25 個  
**預期完成後總測試數**: 190-197 個  
**預期完成時間**: 2-3 小時

---

## 1. 概述

Batch 2 將實現 SDK 執行層的**容錯和恢復機制**，提高系統的可靠性和穩定性。

### 核心目標
- ✅ 實現多種重試策略（Exponential, Linear, Fixed）
- ✅ 實現自動故障偵測和恢復
- ✅ 實現熔斷器保護機制
- ✅ 實現故障轉移邏輯
- ✅ 提供詳細的故障日誌和監控

### 依賴關係
- Batch 1 完成：SDKExecutor, SDKSession, RalphLoopClient 集成 ✅
- 現有模組：CircuitBreaker, ContextManager, PersistenceManager

---

## 2. 架構設計

### 2.1 核心模組組成

```
Batch 2 容錯機制
├── 重試策略層 (Retry Strategies)
│   ├── RetryPolicy 介面
│   ├── ExponentialBackoffRetry (指數退避)
│   ├── LinearBackoffRetry (線性退避)
│   └── FixedIntervalRetry (固定間隔)
├── 故障檢測層 (Fault Detection)
│   ├── FailureDetector 介面
│   ├── TimeoutDetector (逾時檢測)
│   ├── ErrorRateDetector (錯誤率檢測)
│   └── HealthCheckDetector (健康檢查)
├── 恢復機制層 (Recovery Mechanism)
│   ├── RecoveryStrategy 介面
│   ├── AutoReconnectRecovery (自動重連)
│   ├── SessionRestoreRecovery (會話恢復)
│   └── FallbackRecovery (故障轉移)
└── 容錯協調器 (Fault Tolerance Coordinator)
    ├── FaultToleranceExecutor
    └── 整合所有故障偵測和恢復機制
```

### 2.2 設計原則

1. **策略模式**：每種重試和恢復方式都是獨立策略
2. **可組合性**：策略可組合以應對複雜場景
3. **監控友善**：所有故障和恢復都有完整日誌
4. **線程安全**：所有操作都是原子性的
5. **性能考慮**：最小化重試開銷

---

## 3. 實現計畫

### 3.1 第一階段：重試策略實現 (6-8 個測試)

#### 文件：`retry_strategy.go`

**結構體定義**：

```go
// RetryPolicy 定義重試策略的介面
type RetryPolicy interface {
    // NextWaitDuration 返回下一次重試前的等待時間
    NextWaitDuration(attempt int) time.Duration
    // MaxRetries 返回最大重試次數
    MaxRetries() int
    // CanRetry 檢查是否可以重試
    CanRetry(attempt int, err error) bool
}

// ExponentialBackoffRetry 指數退避重試策略
type ExponentialBackoffRetry struct {
    initialDelay time.Duration  // 初始延遲 (預設: 100ms)
    maxDelay     time.Duration  // 最大延遲 (預設: 30s)
    multiplier   float64        // 乘數 (預設: 2.0)
    maxAttempts  int            // 最大嘗試次數
    jitter       bool           // 是否添加抖動
}

// LinearBackoffRetry 線性退避重試策略
type LinearBackoffRetry struct {
    initialDelay time.Duration  // 初始延遲
    increment    time.Duration  // 增量
    maxDelay     time.Duration  // 最大延遲
    maxAttempts  int            // 最大嘗試次數
}

// FixedIntervalRetry 固定間隔重試策略
type FixedIntervalRetry struct {
    interval    time.Duration
    maxAttempts int
}

// RetryMetrics 重試指標
type RetryMetrics struct {
    TotalAttempts      int64         // 總嘗試次數
    SuccessfulAttempts int64         // 成功次數
    FailedAttempts     int64         // 失敗次數
    TotalDelay         time.Duration // 總等待時間
    LastError          error
    LastRetryTime      time.Time
}
```

**核心方法**：

```go
// ExponentialBackoffRetry
- NewExponentialBackoffRetry(maxAttempts int) *ExponentialBackoffRetry
- (r *ExponentialBackoffRetry) NextWaitDuration(attempt int) time.Duration
- (r *ExponentialBackoffRetry) MaxRetries() int
- (r *ExponentialBackoffRetry) CanRetry(attempt int, err error) bool

// LinearBackoffRetry  
- NewLinearBackoffRetry(maxAttempts int) *LinearBackoffRetry
- (r *LinearBackoffRetry) NextWaitDuration(attempt int) time.Duration
- (r *LinearBackoffRetry) MaxRetries() int
- (r *LinearBackoffRetry) CanRetry(attempt int, err error) bool

// FixedIntervalRetry
- NewFixedIntervalRetry(interval time.Duration, maxAttempts int) *FixedIntervalRetry
- (r *FixedIntervalRetry) NextWaitDuration(attempt int) time.Duration
- (r *FixedIntervalRetry) MaxRetries() int
- (r *FixedIntervalRetry) CanRetry(attempt int, err error) bool

// RetryMetrics
- (m *RetryMetrics) RecordAttempt(success bool, err error, delay time.Duration)
- (m *RetryMetrics) GetSuccessRate() float64
- (m *RetryMetrics) GetAverageDelay() time.Duration
```

**測試計畫** (6-8 個):
- TestExponentialBackoffRetryBasic
- TestExponentialBackoffRetryWithJitter
- TestLinearBackoffRetryBasic
- TestFixedIntervalRetryBasic
- TestRetryPolicyCanRetry
- TestRetryMetricsRecording
- TestRetryPoliciesComparison

---

### 3.2 第二階段：故障檢測層 (5-6 個測試)

#### 文件：`failure_detection.go`

**結構體定義**：

```go
// FailureDetector 故障檢測介面
type FailureDetector interface {
    // Detect 檢測是否發生故障
    Detect(err error, executionTime time.Duration) bool
    // GetFailureType 獲得故障類型
    GetFailureType() string
    // Reset 重置檢測狀態
    Reset()
}

// TimeoutDetector 逾時檢測器
type TimeoutDetector struct {
    threshold time.Duration
    count     int64
}

// ErrorRateDetector 錯誤率檢測器
type ErrorRateDetector struct {
    windowSize      int
    errorThreshold  float64 // 預設: 0.5 (50%)
    errors          []error
    mu              sync.RWMutex
}

// HealthCheckDetector 健康檢查檢測器
type HealthCheckDetector struct {
    checkInterval  time.Duration
    unhealthyCount int
    maxUnhealthy   int
    lastCheckTime  time.Time
}

// MultiDetector 組合式檢測器
type MultiDetector struct {
    detectors []FailureDetector
}
```

**核心方法**：

```go
// TimeoutDetector
- NewTimeoutDetector(threshold time.Duration) *TimeoutDetector
- (d *TimeoutDetector) Detect(err error, executionTime time.Duration) bool
- (d *TimeoutDetector) GetFailureType() string

// ErrorRateDetector
- NewErrorRateDetector(windowSize int, threshold float64) *ErrorRateDetector
- (d *ErrorRateDetector) Detect(err error, executionTime time.Duration) bool
- (d *ErrorRateDetector) GetErrorRate() float64

// HealthCheckDetector
- NewHealthCheckDetector(interval time.Duration) *HealthCheckDetector
- (d *HealthCheckDetector) Detect(err error, executionTime time.Duration) bool

// MultiDetector
- NewMultiDetector(detectors ...FailureDetector) *MultiDetector
- (d *MultiDetector) Detect(err error, executionTime time.Duration) bool
```

**測試計畫** (5-6 個):
- TestTimeoutDetectorBasic
- TestErrorRateDetectorBasic
- TestHealthCheckDetectorBasic
- TestMultiDetectorCombination
- TestFailureDetectorReset

---

### 3.3 第三階段：恢復機制層 (4-5 個測試)

#### 文件：`recovery_mechanism.go`

**結構體定義**：

```go
// RecoveryStrategy 恢復策略介面
type RecoveryStrategy interface {
    // Recover 執行恢復操作
    Recover(ctx context.Context, executor *SDKExecutor, err error) error
    // CanRecover 檢查是否可以恢復
    CanRecover(err error) bool
    // Priority 優先級（高優先級先執行）
    Priority() int
}

// AutoReconnectRecovery 自動重連恢復
type AutoReconnectRecovery struct {
    maxRetries  int
    retryDelay  time.Duration
}

// SessionRestoreRecovery 會話恢復
type SessionRestoreRecovery struct {
    backupManager *SessionBackupManager
}

// FallbackRecovery 故障轉移恢復
type FallbackRecovery struct {
    fallbackExecutor *CLIExecutor
}

// RecoveryCoordinator 恢復協調器
type RecoveryCoordinator struct {
    strategies []RecoveryStrategy
    mu         sync.RWMutex
}
```

**核心方法**：

```go
// AutoReconnectRecovery
- NewAutoReconnectRecovery(maxRetries int) *AutoReconnectRecovery
- (r *AutoReconnectRecovery) Recover(ctx context.Context, executor *SDKExecutor, err error) error
- (r *AutoReconnectRecovery) CanRecover(err error) bool

// SessionRestoreRecovery
- NewSessionRestoreRecovery() *SessionRestoreRecovery
- (r *SessionRestoreRecovery) Recover(ctx context.Context, executor *SDKExecutor, err error) error

// FallbackRecovery
- NewFallbackRecovery(fallback *CLIExecutor) *FallbackRecovery
- (r *FallbackRecovery) Recover(ctx context.Context, executor *SDKExecutor, err error) error

// RecoveryCoordinator
- NewRecoveryCoordinator() *RecoveryCoordinator
- (c *RecoveryCoordinator) AddStrategy(strategy RecoveryStrategy)
- (c *RecoveryCoordinator) Recover(ctx context.Context, executor *SDKExecutor, err error) error
- (c *RecoveryCoordinator) GetRecoveryMetrics() RecoveryMetrics
```

**測試計畫** (4-5 個):
- TestAutoReconnectRecoveryBasic
- TestSessionRestoreRecoveryBasic
- TestFallbackRecoveryBasic
- TestRecoveryCoodinatorPriority

---

### 3.4 第四階段：容錯執行器集成 (4-6 個測試)

#### 文件：`fault_tolerant_executor.go`

**結構體定義**：

```go
// FaultToleranceConfig 容錯配置
type FaultToleranceConfig struct {
    RetryPolicy        RetryPolicy
    FailureDetectors   []FailureDetector
    RecoveryStrategies []RecoveryStrategy
    EnableCircuitBreaker bool
    CircuitBreakerThreshold int
    LogLevel           string
}

// FaultTolerantExecutor 容錯執行器
type FaultTolerantExecutor struct {
    executor       *SDKExecutor
    cliExecutor    *CLIExecutor
    config         *FaultToleranceConfig
    detector       FailureDetector
    coordinator    *RecoveryCoordinator
    metrics        *FaultToleranceMetrics
    mu             sync.RWMutex
    circuitBreaker *CircuitBreaker
}

// FaultToleranceMetrics 容錯指標
type FaultToleranceMetrics struct {
    TotalExecutions      int64
    SuccessfulExecutions int64
    FailedExecutions     int64
    RecoveredExecutions  int64
    CircuitBreakerTrips  int64
    TotalRetries         int64
    AverageRetryCount    float64
}
```

**核心方法**：

```go
// FaultTolerantExecutor
- NewFaultTolerantExecutor(executor *SDKExecutor, config *FaultToleranceConfig) *FaultTolerantExecutor
- (e *FaultTolerantExecutor) Execute(ctx context.Context, fn func(context.Context) (string, error)) (string, error)
- (e *FaultTolerantExecutor) ExecuteComplete(ctx context.Context, prompt string) (string, error)
- (e *FaultTolerantExecutor) ExecuteExplain(ctx context.Context, code string) (string, error)
- (e *FaultTolerantExecutor) ExecuteGenerateTests(ctx context.Context, code string) (string, error)
- (e *FaultTolerantExecutor) ExecuteCodeReview(ctx context.Context, code string) (string, error)
- (e *FaultTolerantExecutor) GetMetrics() *FaultToleranceMetrics
- (e *FaultTolerantExecutor) SetRetryPolicy(policy RetryPolicy)
- (e *FaultTolerantExecutor) AddFailureDetector(detector FailureDetector)
- (e *FaultTolerantExecutor) AddRecoveryStrategy(strategy RecoveryStrategy)
```

**測試計畫** (4-6 個):
- TestFaultTolerantExecutorBasic
- TestFaultTolerantExecutorWithRetry
- TestFaultTolerantExecutorWithRecovery
- TestFaultTolerantExecutorCircuitBreaker
- TestFaultTolerantExecutorMetrics
- TestFaultTolerantExecutorIntegration

---

### 3.5 第五階段：RalphLoopClient 擴展 (1-2 個測試)

#### 文件：修改 `client.go`

**新增結構體**：

```go
// RalphLoopClient 新增字段
type RalphLoopClient struct {
    // ... 現有字段 ...
    
    // 容錯相關字段
    faultToleranceExecutor *FaultTolerantExecutor
    faultToleranceConfig   *FaultToleranceConfig
}
```

**新增方法**：

```go
// 容錯執行方法
- (c *RalphLoopClient) ExecuteWithFaultTolerance(ctx context.Context, prompt string) (string, error)
- (c *RalphLoopClient) ConfigureFaultTolerance(config *FaultToleranceConfig)
- (c *RalphLoopClient) GetFaultToleranceMetrics() *FaultToleranceMetrics
- (c *RalphLoopClient) GetRetryPolicy() RetryPolicy
- (c *RalphLoopClient) SetRetryPolicy(policy RetryPolicy)
```

**測試計畫** (1-2 個):
- TestClientFaultToleranceIntegration

---

## 4. 測試戰略

### 4.1 測試組織

```
test/fault_tolerance/
├── retry_strategy_test.go (6-8 個測試)
├── failure_detection_test.go (5-6 個測試)
├── recovery_mechanism_test.go (4-5 個測試)
├── fault_tolerant_executor_test.go (4-6 個測試)
└── client_fault_tolerance_integration_test.go (1-2 個測試)
```

### 4.2 測試類型

1. **單元測試**：測試各個模組的獨立功能
2. **集成測試**：測試模組間的協作
3. **場景測試**：模擬真實失敗場景
4. **性能測試**：驗證重試和恢復的性能開銷

### 4.3 測試場景

| 場景 | 描述 | 預期結果 |
|------|------|---------|
| 正常執行 | 無錯誤 | 直接返回結果 |
| 臨時錯誤 | 一次性錯誤後恢復 | 重試後成功 |
| 持續錯誤 | 多次失敗 | 耗盡重試後返回錯誤 |
| 逾時 | 執行超時 | 檢測逾時並重試 |
| 高錯誤率 | 錯誤率超過閾值 | 觸發熔斷 |
| 故障轉移 | SDK 失敗 | 轉向 CLI 執行 |
| 會話丟失 | 會話無效 | 恢復會話 |

---

## 5. 實現優先級

### 優先級 1（關鍵）
- ✅ RetryPolicy 介面和實現
- ✅ FaultTolerantExecutor 基本框架
- ✅ AutoReconnectRecovery

### 優先級 2（重要）
- ✅ FailureDetector 框架
- ✅ RecoveryCoordinator
- ✅ RalphLoopClient 集成

### 優先級 3（增強）
- ✅ 多種故障檢測器實現
- ✅ 進階恢復策略
- ✅ 監控和指標

---

## 6. 預期結果

### 代碼指標
| 項目 | 預期值 |
|------|--------|
| 新增代碼行數 | 800-1000 行 |
| 新增文件 | 4-5 個 |
| 新增測試 | 20-25 個 |

### 測試指標
| 項目 | 預期值 |
|------|--------|
| 通過率 | 100% |
| 代碼覆蓋 | >90% |
| 執行時間 | <10 秒 |

### 完成後總體進度
| 項目 | 當前 | 完成後 | 進度 |
|------|------|--------|------|
| 測試總數 | 172 | 192-197 | ✅ 95%+ |
| Stage 8.3 完成 | 50% | 85% | ✅ 重大進展 |

---

## 7. 實現時間表

| 階段 | 時間估計 | 優先級 |
|------|---------|--------|
| 第一階段：重試策略 | 30-40 分鐘 | P1 |
| 第二階段：故障檢測 | 30-40 分鐘 | P2 |
| 第三階段：恢復機制 | 30-40 分鐘 | P2 |
| 第四階段：執行器集成 | 30-40 分鐘 | P1 |
| 第五階段：Client 擴展 | 15-20 分鐘 | P2 |
| 測試和調試 | 20-30 分鐘 | P1 |
| **總計** | **2.5-3.5 小時** | - |

---

## 8. 後續計畫

### Batch 2 完成後
- Batch 3：執行模式選擇器（10-15 個測試）
- Stage 8.3 最終驗證和文檔化

### 集成到主專案
- 更新 README 和使用文檔
- 性能基準測試
- 生產環境驗證

---

## 9. 參考資料

- **現有代碼**：CircuitBreaker (internal/ghcopilot/circuit_breaker.go)
- **相關文件**：STAGE_8_3_PLANNING.md, ARCHITECTURE.md
- **SDK 文件**：Copilot SDK 官方文檔

---

**下一步**: 確認規劃無誤後，開始實現第一階段（重試策略）
