# Stage 8.3 Batch 2 實現框架詳解

**文件日期**: 2026-01-22  
**用途**: 開發者參考和實現指南

---

## 架構視圖

```
┌─────────────────────────────────────────────────────────────────┐
│                    RalphLoopClient (應用層)                       │
├─────────────────────────────────────────────────────────────────┤
│  ExecuteWithFaultTolerance() → SetRetryPolicy() → GetMetrics()   │
├─────────────────────────────────────────────────────────────────┤
│         FaultTolerantExecutor (容錯協調層)                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Execute(fn) → Detect() → Retry() → Recover() → Result   │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│  SDKExecutor        │ RetryPolicy      │ FailureDetector         │
│  (執行層)           │ (重試策略)       │ (故障檢測)              │
├─────────────────────────────────────────────────────────────────┤
│  RecoveryCoordinator │ RecoveryStrategy │ CircuitBreaker          │
│  (恢復協調)         │ (恢復策略)       │ (熔斷器)                │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. 重試策略詳解

### 1.1 指數退避 (Exponential Backoff)

```go
重試次數 │ 延遲時間 (ms)
─────────┼───────────────
   1     │ 100
   2     │ 200
   3     │ 400
   4     │ 800
   5     │ 1600
   ...   │ cap at 30000ms (30s)

延遲公式：min(initialDelay * (multiplier ^ attempt), maxDelay)
         + random(0, jitter)
```

**使用場景**：
- 網絡臨時故障
- 服務器短時間過載
- SDK 連接超時

**實現關鍵點**：
```go
type ExponentialBackoffRetry struct {
    initialDelay time.Duration  // 100ms
    maxDelay     time.Duration  // 30s
    multiplier   float64        // 2.0
    maxAttempts  int            // 3-5
    jitter       bool           // true
}

NextWaitDuration(attempt):
  delay = initialDelay * (multiplier ^ attempt)
  if delay > maxDelay: delay = maxDelay
  if jitter: delay += random(0, delay * 0.1)
  return delay
```

### 1.2 線性退避 (Linear Backoff)

```go
重試次數 │ 延遲時間 (ms)
─────────┼───────────────
   1     │ 100
   2     │ 200
   3     │ 300
   4     │ 400
   5     │ 500
   ...   │ cap at 5000ms (5s)

延遲公式：min(initialDelay + (increment * attempt), maxDelay)
```

**使用場景**：
- 資源限制恢復（如連接池耗盡）
- 逐漸增加負載
- 可預測的恢復時間

### 1.3 固定間隔 (Fixed Interval)

```go
重試次數 │ 延遲時間 (ms)
─────────┼───────────────
   1     │ 500
   2     │ 500
   3     │ 500
   4     │ 500
   5     │ 500

延遲公式：固定 interval 時間
```

**使用場景**：
- 已知恢復時間的故障
- 簡單的重試邏輯
- 測試和調試

---

## 2. 故障檢測詳解

### 2.1 逾時檢測 (Timeout Detection)

```go
檢測邏輯：
  if executionTime > threshold:
      count++
      if count >= consecutiveThreshold:
          return TIMEOUT_FAULT
```

**配置範例**：
```go
detector := NewTimeoutDetector(5 * time.Second)
// 執行時間超過 5s 視為逾時故障
```

### 2.2 錯誤率檢測 (Error Rate Detection)

```go
檢測邏輯：
  維護滑動窗口 (N=10 次執行)
  errorRate = failedAttempts / totalAttempts
  if errorRate > threshold (50%):
      return ERROR_RATE_FAULT
```

**配置範例**：
```go
detector := NewErrorRateDetector(
    windowSize: 10,          // 檢查最近 10 次
    threshold: 0.5,          // 錯誤率 > 50%
)
```

### 2.3 健康檢查檢測 (Health Check Detection)

```go
檢測邏輯：
  每 interval 時間執行一次健康檢查
  if unhealthy_count > maxUnhealthy:
      return HEALTH_FAULT
  else:
      reset unhealthy_count
```

**配置範例**：
```go
detector := NewHealthCheckDetector(
    checkInterval: 30 * time.Second,
    maxUnhealthy: 3,
)
```

### 2.4 多檢測器組合

```go
detectors := []FailureDetector{
    NewTimeoutDetector(5 * time.Second),
    NewErrorRateDetector(10, 0.5),
    NewHealthCheckDetector(30 * time.Second),
}

multiDetector := NewMultiDetector(detectors...)

// 任何檢測器檢測到故障即返回 true
if multiDetector.Detect(err, duration):
    // 發生故障
```

---

## 3. 恢復機制詳解

### 3.1 自動重連恢復 (Auto-Reconnect)

```
流程：
  1. 檢測連接丟失
  2. 嘗試重新建立連接
  3. 驗證連接有效性
  4. 恢復執行
  5. 記錄恢復過程

實現：
  max_retries: 3
  retry_delay: 100ms to 1s (指數退避)
```

**適用故障**：
- 網絡連接中斷
- SDK 客戶端失效
- 會話過期

### 3.2 會話恢復 (Session Restore)

```
流程：
  1. 偵測會話無效
  2. 檢查備份會話
  3. 恢復會話狀態
  4. 驗證會話可用性
  5. 繼續執行

實現要點：
  - 定期備份活動會話狀態
  - 保留會話元數據
  - 支援快速恢復
```

**適用故障**：
- 會話超時
- 會話丟失
- 會話狀態不一致

### 3.3 故障轉移恢復 (Fallback)

```
流程（SDK → CLI 轉移）：
  1. SDK 執行失敗
  2. 觸發故障轉移
  3. 使用 CLI 執行器執行
  4. 解析 CLI 結果
  5. 返回統一格式結果

配置：
  enable_fallback: true
  fallback_executor: CLIExecutor
```

**適用故障**：
- SDK 完全失效
- SDK 不可恢復錯誤
- 需要強制使用 CLI

### 3.4 恢復優先級

```
優先級   │ 恢復策略           │ 適用場景
─────────┼──────────────────┼──────────────────
  高     │ AutoReconnect    │ 臨時連接問題
  中     │ SessionRestore   │ 會話相關問題
  低     │ Fallback         │ SDK 完全失效
```

---

## 4. 執行流程詳解

### 4.1 完整執行流程圖

```
┌─────────────────────┐
│ ExecuteWithSDK()    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────────────────────┐
│ FaultTolerantExecutor.Execute()     │
│ ├─ Attempt = 1                      │
│ └─ MaxAttempts = 5                  │
└──────────┬──────────────────────────┘
           │
    ┌──────▼──────┐
    │ Try Execute │
    └──────┬──────┘
           │
     ┌─────▼─────┐
     │ Success?  │
     └─┬───┬─────┘
      ┌┘   └──No──┐
     Yes          │
      │           ▼
      │    ┌──────────────────┐
      │    │ Detect Failure   │
      │    └──┬────────────┬──┘
      │       │            │
      │     No             Yes
      │       │            │
      │    ┌──▼──┐    ┌────▼────────┐
      │    │ OK  │    │ Attempt < 5?│
      │    └─────┘    └──┬────┬─────┘
      │                  │    │
      │                Yes    No
      │                  │    │
      │              ┌───▼──┐ │
      │              │ Wait │ │
      │              └────┬─┘ │
      │                   │   │
      │            ┌──────┘   │
      │            │          ▼
      │            │    ┌──────────────┐
      │            │    │ Try Recover  │
      │            │    └──┬────────┬──┘
      │            │       │        │
      │            │     No        Yes
      │            │       │        │
      │            │    ┌──▼──┐ ┌──▼──┐
      │            │    │ Err │ │ OK  │
      │            │    └─────┘ └──┬──┘
      │            │              │
      │            │           Retry
      │            │              │
      └────────────┴──────┬───────┘
                         │
                  ┌──────▼──────┐
                  │ Return Result│
                  └──────────────┘
```

### 4.2 重試邏輯偽代碼

```python
function Execute(fn):
  metrics = new Metrics()
  
  for attempt = 1 to maxAttempts:
    try:
      result = fn()
      metrics.recordSuccess()
      return result
      
    catch error:
      startTime = now()
      
      // 故障檢測
      if detector.Detect(error, duration):
        metrics.recordFailure(error)
        
        // 嘗試恢復
        if attempt < maxAttempts:
          recoveryError = coordinator.Recover()
          
          if recoveryError:
            // 恢復失敗，等待後重試
            delay = retryPolicy.NextWaitDuration(attempt)
            sleep(delay)
            metrics.recordRetry(delay)
            continue
          else:
            // 恢復成功，重新執行
            metrics.recordRecovery()
            continue
      
      // 不可恢復的錯誤
      metrics.recordUnrecoverable()
      throw error
  
  // 耗盡所有重試
  throw maxAttemptsExceeded()
```

---

## 5. 配置範例

### 5.1 保守配置（優先穩定性）

```go
config := &FaultToleranceConfig{
    RetryPolicy: NewExponentialBackoffRetry(3),
    FailureDetectors: []FailureDetector{
        NewTimeoutDetector(10 * time.Second),
    },
    RecoveryStrategies: []RecoveryStrategy{
        NewAutoReconnectRecovery(2),
    },
    EnableCircuitBreaker: true,
    CircuitBreakerThreshold: 5,
}
```

### 5.2 激進配置（優先吞吐量）

```go
config := &FaultToleranceConfig{
    RetryPolicy: NewLinearBackoffRetry(5),
    FailureDetectors: []FailureDetector{
        NewTimeoutDetector(3 * time.Second),
        NewErrorRateDetector(10, 0.3),
    },
    RecoveryStrategies: []RecoveryStrategy{
        NewAutoReconnectRecovery(3),
        NewSessionRestoreRecovery(),
        NewFallbackRecovery(cliExecutor),
    },
    EnableCircuitBreaker: true,
    CircuitBreakerThreshold: 3,
}
```

### 5.3 均衡配置（默認推薦）

```go
config := &FaultToleranceConfig{
    RetryPolicy: NewExponentialBackoffRetry(5),
    FailureDetectors: []FailureDetector{
        NewTimeoutDetector(5 * time.Second),
        NewErrorRateDetector(10, 0.5),
    },
    RecoveryStrategies: []RecoveryStrategy{
        NewAutoReconnectRecovery(3),
        NewSessionRestoreRecovery(),
    },
    EnableCircuitBreaker: true,
    CircuitBreakerThreshold: 4,
}
```

---

## 6. 性能考慮

### 6.1 內存開銷

| 組件 | 平均大小 | 說明 |
|------|---------|------|
| RetryPolicy | <1 KB | 配置對象 |
| FailureDetector | 2-5 KB | 包含歷史窗口 |
| RecoveryStrategy | 1-2 KB | 配置對象 |
| FaultToleranceMetrics | 100 B | 計數器 |
| **每次執行開銷** | **<10 KB** | - |

### 6.2 時間開銷

| 操作 | 平均時間 | 備註 |
|------|---------|------|
| 重試決策 | <1 ms | 簡單邏輯 |
| 故障檢測 | <1 ms | O(1) 或 O(window_size) |
| 恢復執行 | 100-1000 ms | 取決於恢復策略 |
| 等待延遲 | 100-30000 ms | 指數退避 |

### 6.3 優化建議

1. **重試策略緩存**：避免重複計算延遲
2. **檢測器聯合短路**：任何檢測器返回 true 即中止
3. **指標采樣**：降低高頻記錄的開銷
4. **異步恢復**：某些恢復策略可異步執行

---

## 7. 錯誤場景模擬

### 7.1 測試場景設置

```go
// 模擬臨時故障
func NewTemporaryErrorSimulator() ErrorSimulator {
    return &ErrorSimulator{
        failCount: 2,  // 前 2 次失敗
        failErr: fmt.Errorf("temporary error"),
    }
}

// 模擬持續故障
func NewPersistentErrorSimulator() ErrorSimulator {
    return &ErrorSimulator{
        alwaysFail: true,
        failErr: fmt.Errorf("persistent error"),
    }
}

// 模擬超時
func NewTimeoutSimulator() ErrorSimulator {
    return &ErrorSimulator{
        delay: 10 * time.Second,
    }
}

// 模擬高錯誤率
func NewHighErrorRateSimulator() ErrorSimulator {
    return &ErrorSimulator{
        errorRate: 0.8,  // 80% 失敗率
    }
}
```

### 7.2 驗證場景

```go
scenarios := []struct {
    name     string
    simulator ErrorSimulator
    expected TestResult
}{
    {"正常執行", nil, SUCCESS},
    {"臨時故障", TemporaryError, SUCCESS_WITH_RETRY},
    {"持續故障", PersistentError, FAILED},
    {"超時", Timeout, SUCCESS_WITH_RECOVERY},
    {"高錯誤率", HighErrorRate, FALLBACK_TRIGGERED},
}
```

---

## 8. 監控指標

### 8.1 關鍵指標

```go
type FaultToleranceMetrics struct {
    // 執行統計
    TotalExecutions      int64   // 總執行次數
    SuccessfulExecutions int64   // 成功次數
    FailedExecutions     int64   // 失敗次數
    RecoveredExecutions  int64   // 恢復成功的執行
    
    // 重試統計
    TotalRetries         int64   // 總重試次數
    SuccessfulRetries    int64   // 成功重試次數
    AverageRetryCount    float64 // 平均重試次數
    MaxRetryCount        int     // 最大重試次數
    
    // 恢復統計
    RecoveryAttempts     int64   // 恢復嘗試次數
    SuccessfulRecoveries int64   // 成功恢復次數
    FailedRecoveries     int64   // 失敗恢復次數
    
    // 熔斷器統計
    CircuitBreakerTrips  int64   // 熔斷器觸發次數
    
    // 性能指標
    AverageExecutionTime time.Duration
    P95ExecutionTime     time.Duration
    P99ExecutionTime     time.Duration
}
```

### 8.2 監控儀表板

```
執行統計：
  成功率: 98.5% (197/200)
  平均執行時間: 245ms
  P95 延遲: 1.2s
  P99 延遲: 3.5s

重試統計：
  重試率: 4.2% (8/190)
  平均重試次數: 1.3
  最大重試次數: 5

恢復統計：
  恢復成功率: 85% (6/7)
  自動重連: 4 次
  會話恢復: 2 次
  故障轉移: 1 次

熔斷器：
  觸發次數: 0
  狀態: CLOSED
```

---

**下一步**: 確認框架無誤，開始實現 Batch 2
