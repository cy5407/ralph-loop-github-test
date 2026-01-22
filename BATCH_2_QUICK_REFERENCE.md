# Batch 2 快速參考指南

**目的**: 快速查詢和實現 Batch 2 的關鍵信息

---

## 🎯 Batch 2 目標

| 項目 | 值 |
|------|-----|
| 新增測試 | 20-25 個 |
| 新增代碼 | 800-1000 行 |
| 新增文件 | 4-5 個 |
| 預期完成時間 | 2.5-3.5 小時 |
| 最終總測試數 | 192-197 個 |

---

## 📋 實現階段清單

### ✅ 完成檢查清單

- [ ] **階段 1：重試策略** (6-8 測試)
  - [ ] RetryPolicy 介面定義
  - [ ] ExponentialBackoffRetry 實現
  - [ ] LinearBackoffRetry 實現
  - [ ] FixedIntervalRetry 實現
  - [ ] RetryMetrics 實現
  - [ ] 6-8 個單元測試

- [ ] **階段 2：故障檢測** (5-6 測試)
  - [ ] FailureDetector 介面定義
  - [ ] TimeoutDetector 實現
  - [ ] ErrorRateDetector 實現
  - [ ] HealthCheckDetector 實現
  - [ ] MultiDetector 實現
  - [ ] 5-6 個單元測試

- [ ] **階段 3：恢復機制** (4-5 測試)
  - [ ] RecoveryStrategy 介面定義
  - [ ] AutoReconnectRecovery 實現
  - [ ] SessionRestoreRecovery 實現
  - [ ] FallbackRecovery 實現
  - [ ] RecoveryCoordinator 實現
  - [ ] 4-5 個單元測試

- [ ] **階段 4：執行器集成** (4-6 測試)
  - [ ] FaultTolerantExecutor 實現
  - [ ] FaultToleranceMetrics 實現
  - [ ] FaultToleranceConfig 實現
  - [ ] 4-6 個集成測試

- [ ] **階段 5：Client 擴展** (1-2 測試)
  - [ ] RalphLoopClient SDK 字段
  - [ ] ExecuteWithFaultTolerance() 方法
  - [ ] ConfigureFaultTolerance() 方法
  - [ ] 1-2 個集成測試

- [ ] **測試和調試**
  - [ ] 所有單元測試通過
  - [ ] 所有集成測試通過
  - [ ] 代碼審查
  - [ ] 性能驗證

---

## 🔑 關鍵代碼片段

### RetryPolicy 基本框架

```go
type RetryPolicy interface {
    NextWaitDuration(attempt int) time.Duration
    MaxRetries() int
    CanRetry(attempt int, err error) bool
}

type ExponentialBackoffRetry struct {
    initialDelay time.Duration
    maxDelay     time.Duration
    multiplier   float64
    maxAttempts  int
    jitter       bool
}

func NewExponentialBackoffRetry(maxAttempts int) *ExponentialBackoffRetry {
    return &ExponentialBackoffRetry{
        initialDelay: 100 * time.Millisecond,
        maxDelay:     30 * time.Second,
        multiplier:   2.0,
        maxAttempts:  maxAttempts,
        jitter:       true,
    }
}

func (r *ExponentialBackoffRetry) NextWaitDuration(attempt int) time.Duration {
    delay := r.initialDelay * time.Duration(math.Pow(r.multiplier, float64(attempt-1)))
    if delay > r.maxDelay {
        delay = r.maxDelay
    }
    if r.jitter {
        delay = delay + time.Duration(rand.Intn(int(delay/10)))
    }
    return delay
}
```

### FailureDetector 基本框架

```go
type FailureDetector interface {
    Detect(err error, executionTime time.Duration) bool
    GetFailureType() string
    Reset()
}

type TimeoutDetector struct {
    threshold time.Duration
}

func NewTimeoutDetector(threshold time.Duration) *TimeoutDetector {
    return &TimeoutDetector{threshold: threshold}
}

func (d *TimeoutDetector) Detect(err error, executionTime time.Duration) bool {
    return executionTime > d.threshold
}

func (d *TimeoutDetector) GetFailureType() string {
    return "timeout"
}
```

### RecoveryStrategy 基本框架

```go
type RecoveryStrategy interface {
    Recover(ctx context.Context, executor *SDKExecutor, err error) error
    CanRecover(err error) bool
    Priority() int
}

type AutoReconnectRecovery struct {
    maxRetries int
    retryDelay time.Duration
}

func NewAutoReconnectRecovery(maxRetries int) *AutoReconnectRecovery {
    return &AutoReconnectRecovery{
        maxRetries: maxRetries,
        retryDelay: 100 * time.Millisecond,
    }
}

func (r *AutoReconnectRecovery) Recover(ctx context.Context, executor *SDKExecutor, err error) error {
    for i := 0; i < r.maxRetries; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(r.retryDelay):
            // 嘗試重連
            if err := executor.reconnect(); err == nil {
                return nil
            }
            r.retryDelay *= 2  // 指數退避
        }
    }
    return fmt.Errorf("reconnection failed")
}

func (r *AutoReconnectRecovery) CanRecover(err error) bool {
    return strings.Contains(err.Error(), "connection")
}

func (r *AutoReconnectRecovery) Priority() int {
    return 10  // 高優先級
}
```

### FaultTolerantExecutor 基本框架

```go
type FaultTolerantExecutor struct {
    executor    *SDKExecutor
    cliExecutor *CLIExecutor
    config      *FaultToleranceConfig
    detector    FailureDetector
    coordinator *RecoveryCoordinator
    metrics     *FaultToleranceMetrics
}

func NewFaultTolerantExecutor(
    executor *SDKExecutor,
    config *FaultToleranceConfig,
) *FaultTolerantExecutor {
    return &FaultTolerantExecutor{
        executor:    executor,
        config:      config,
        detector:    NewMultiDetector(config.FailureDetectors...),
        coordinator: NewRecoveryCoordinator(config.RecoveryStrategies...),
        metrics:     &FaultToleranceMetrics{},
    }
}

func (e *FaultTolerantExecutor) Execute(
    ctx context.Context,
    fn func(context.Context) (string, error),
) (string, error) {
    for attempt := 1; attempt <= e.config.RetryPolicy.MaxRetries(); attempt++ {
        result, err := fn(ctx)
        
        if err == nil {
            e.metrics.RecordSuccess()
            return result, nil
        }
        
        // 檢測故障
        if e.detector.Detect(err, 0) {  // 簡化，實際需要記錄執行時間
            e.metrics.RecordFailure(err)
            
            // 嘗試恢復
            recoveryErr := e.coordinator.Recover(ctx, e.executor, err)
            if recoveryErr == nil {
                e.metrics.RecordRecovery()
                continue  // 重試
            }
            
            // 恢復失敗，等待後重試
            if attempt < e.config.RetryPolicy.MaxRetries() {
                delay := e.config.RetryPolicy.NextWaitDuration(attempt)
                select {
                case <-ctx.Done():
                    return "", ctx.Err()
                case <-time.After(delay):
                    e.metrics.RecordRetry(delay)
                    continue
                }
            }
        }
        
        return "", err
    }
    
    return "", fmt.Errorf("max retries exceeded")
}

func (e *FaultTolerantExecutor) GetMetrics() *FaultToleranceMetrics {
    return e.metrics
}
```

---

## 📊 測試計畫概覽

### 階段 1 測試

```go
// retry_strategy_test.go

TestExponentialBackoffRetryBasic()
  ✓ 驗證延遲計算正確
  ✓ 驗證最大延遲限制
  ✓ 驗證退避倍數

TestExponentialBackoffRetryWithJitter()
  ✓ 驗證抖動添加
  ✓ 驗證抖動範圍

TestLinearBackoffRetryBasic()
  ✓ 驗證線性遞增
  ✓ 驗證最大延遲限制

TestFixedIntervalRetryBasic()
  ✓ 驗證固定間隔
  ✓ 驗證最大重試次數

TestRetryPolicyCanRetry()
  ✓ 驗證可重試邏輯
  ✓ 驗證不可重試的錯誤

TestRetryMetricsRecording()
  ✓ 驗證指標記錄
  ✓ 驗證成功率計算

TestRetryPoliciesComparison()
  ✓ 比較三種策略的性能
```

### 階段 2 測試

```go
// failure_detection_test.go

TestTimeoutDetectorBasic()
  ✓ 驗證超時檢測

TestErrorRateDetectorBasic()
  ✓ 驗證錯誤率計算
  ✓ 驗證滑動窗口

TestHealthCheckDetectorBasic()
  ✓ 驗證健康檢查

TestMultiDetectorCombination()
  ✓ 驗證組合檢測

TestFailureDetectorReset()
  ✓ 驗證狀態重置
```

### 階段 3 測試

```go
// recovery_mechanism_test.go

TestAutoReconnectRecoveryBasic()
  ✓ 驗證自動重連

TestSessionRestoreRecoveryBasic()
  ✓ 驗證會話恢復

TestFallbackRecoveryBasic()
  ✓ 驗證故障轉移

TestRecoveryCoordinatorPriority()
  ✓ 驗證優先級順序
```

### 階段 4 測試

```go
// fault_tolerant_executor_test.go

TestFaultTolerantExecutorBasic()
  ✓ 正常執行

TestFaultTolerantExecutorWithRetry()
  ✓ 臨時故障重試

TestFaultTolerantExecutorWithRecovery()
  ✓ 故障恢復

TestFaultTolerantExecutorCircuitBreaker()
  ✓ 熔斷器觸發

TestFaultTolerantExecutorMetrics()
  ✓ 指標記錄

TestFaultTolerantExecutorIntegration()
  ✓ 完整場景集成
```

### 階段 5 測試

```go
// client_fault_tolerance_integration_test.go

TestClientFaultToleranceIntegration()
  ✓ Client 集成
  ✓ 配置應用
  ✓ 方法調用
```

---

## ⏱️ 時間估計詳細

| 階段 | 任務 | 時間 |
|------|------|------|
| 1 | 重試策略實現 | 25-35 分鐘 |
| 2 | 故障檢測實現 | 25-35 分鐘 |
| 3 | 恢復機制實現 | 25-35 分鐘 |
| 4 | 執行器集成 | 25-35 分鐘 |
| 5 | Client 擴展 | 15-20 分鐘 |
| 6 | 測試執行和修復 | 20-30 分鐘 |
| **總計** | **~140-190 分鐘 (2.5-3.5 小時)** | |

---

## 🔗 相關文件連結

- [STAGE_8_3_BATCH_2_PLANNING.md](STAGE_8_3_BATCH_2_PLANNING.md) - 詳細規劃
- [BATCH_2_IMPLEMENTATION_FRAMEWORK.md](BATCH_2_IMPLEMENTATION_FRAMEWORK.md) - 實現框架
- [ARCHITECTURE.md](ARCHITECTURE.md) - 整體架構
- [internal/ghcopilot/sdk_executor.go](internal/ghcopilot/sdk_executor.go) - SDK 執行器實現
- [internal/ghcopilot/circuit_breaker.go](internal/ghcopilot/circuit_breaker.go) - 現有熔斷器

---

## 💡 開發建議

### 代碼組織

```
internal/ghcopilot/
├── retry_strategy.go          # 重試策略
├── retry_strategy_test.go
├── failure_detection.go        # 故障檢測
├── failure_detection_test.go
├── recovery_mechanism.go       # 恢復機制
├── recovery_mechanism_test.go
├── fault_tolerant_executor.go  # 容錯執行器
├── fault_tolerant_executor_test.go
├── client_fault_tolerance_integration_test.go
└── ... (現有文件)
```

### 測試先行 (TDD)

1. 先編寫測試用例
2. 再實現功能
3. 驗證測試通過

### 代碼審查檢查點

- [ ] 線程安全（使用 mutex）
- [ ] 錯誤處理完整
- [ ] 文檔註釋清晰
- [ ] 測試覆蓋率 >90%
- [ ] 性能無迴歸

### 提交規範

```
[Batch2] 階段X: 描述

詳細說明
- 新增功能
- 修復問題
- 改進性能

Tests: 新增 X 個測試，共通過 Y 個
```

---

## ⚠️ 常見問題

### Q1: 重試策略應該選擇哪一種？
**A**: 根據場景選擇：
- 臨時故障 → 指數退避
- 資源限制 → 線性退避
- 已知恢復時間 → 固定間隔

### Q2: 多個檢測器衝突怎麼辦？
**A**: 使用 MultiDetector 組合，任何檢測器觸發即標記為故障

### Q3: 恢復順序如何決定？
**A**: 根據 Priority() 排序，高優先級先執行

### Q4: 如何監控容錯執行？
**A**: 使用 GetMetrics() 獲取指標，監控成功率、重試率等

### Q5: 與現有熔斷器的關係？
**A**: 容錯執行器使用熔斷器作為最後防線，防止級聯故障

---

## 🚀 快速開始

1. **閱讀規劃**: [STAGE_8_3_BATCH_2_PLANNING.md](STAGE_8_3_BATCH_2_PLANNING.md)
2. **閱讀框架**: [BATCH_2_IMPLEMENTATION_FRAMEWORK.md](BATCH_2_IMPLEMENTATION_FRAMEWORK.md)
3. **階段 1**: 實現重試策略 (~35 分鐘)
4. **階段 2**: 實現故障檢測 (~35 分鐘)
5. **階段 3**: 實現恢復機制 (~35 分鐘)
6. **階段 4**: 集成執行器 (~35 分鐘)
7. **階段 5**: 擴展 Client (~20 分鐘)
8. **測試驗證**: 運行完整測試 (~30 分鐘)

---

**下一步**: 選擇階段開始實現或聯繫 AI 進行代碼審查
