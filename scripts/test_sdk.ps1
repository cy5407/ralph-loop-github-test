#!/usr/bin/env pwsh
# SDK PoC 測試腳本

Write-Host "=== SDK PoC 測試 ===" -ForegroundColor Cyan

# 1. 檢查 Copilot CLI 版本
Write-Host "`n1. 檢查 Copilot CLI 版本..." -ForegroundColor Yellow
copilot --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Copilot CLI 未安裝或無法執行" -ForegroundColor Red
    exit 1
}

# 2. 設定工作目錄
$projectDir = "C:\Users\cy540\OneDrive\桌面\Github CLI 自動跌代"
Set-Location $projectDir

# 3. 執行 SDK 基本連線測試
Write-Host "`n2. 執行 SDK 基本連線測試..." -ForegroundColor Yellow
go test -v -run TestSDKBasicConnection './test/sdk_poc_test.go'

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ SDK 基本連線測試通過！" -ForegroundColor Green
} else {
    Write-Host "`n❌ SDK 基本連線測試失敗" -ForegroundColor Red
    Write-Host "錯誤碼: $LASTEXITCODE" -ForegroundColor Red
}

# 4. 執行 Session 建立測試
Write-Host "`n3. 執行 Session 建立測試..." -ForegroundColor Yellow
go test -v -run TestSDKSessionCreation './test/sdk_poc_test.go'

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Session 建立測試通過！" -ForegroundColor Green
} else {
    Write-Host "`n❌ Session 建立測試失敗" -ForegroundColor Red
}

# 5. 顯示決策報告
Write-Host "`n4. 執行決策測試..." -ForegroundColor Yellow
go test -v -run TestSDKDecision './test/sdk_poc_test.go'

Write-Host "`n=== 測試完成 ===" -ForegroundColor Cyan
Write-Host "`n如果所有測試都通過，請執行以下命令繼續 SDK 整合：" -ForegroundColor Green
Write-Host "go test -v ./test" -ForegroundColor White
