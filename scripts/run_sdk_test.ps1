# SDK PoC Test Script

Write-Host "=== SDK PoC Testing ===" -ForegroundColor Cyan

# 1. Check Copilot CLI version
Write-Host "`n1. Checking Copilot CLI version..." -ForegroundColor Yellow
copilot --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Copilot CLI not found" -ForegroundColor Red
    exit 1
}

# 2. Set working directory
$projectDir = "C:\Users\cy540\OneDrive\桌面\Github CLI 自動跌代"
Set-Location $projectDir

# 3. Run SDK basic connection test
Write-Host "`n2. Running SDK basic connection test..." -ForegroundColor Yellow
go test -v -run TestSDKBasicConnection './test/sdk_poc_test.go'

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nSUCCESS: SDK basic connection test passed!" -ForegroundColor Green
} else {
    Write-Host "`nFAILED: SDK basic connection test failed" -ForegroundColor Red
    Write-Host "Error code: $LASTEXITCODE" -ForegroundColor Red
}

# 4. Run Session creation test
Write-Host "`n3. Running Session creation test..." -ForegroundColor Yellow
go test -v -run TestSDKSessionCreation './test/sdk_poc_test.go'

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nSUCCESS: Session creation test passed!" -ForegroundColor Green
} else {
    Write-Host "`nFAILED: Session creation test failed" -ForegroundColor Red
}

# 5. Run decision test
Write-Host "`n4. Running decision test..." -ForegroundColor Yellow
go test -v -run TestSDKDecision './test/sdk_poc_test.go'

Write-Host "`n=== Testing Complete ===" -ForegroundColor Cyan
Write-Host "`nIf all tests passed, run: go test -v ./test" -ForegroundColor Green
