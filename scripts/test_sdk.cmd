@echo off
REM SDK PoC 測試腳本 (CMD 版本)

echo === SDK PoC 測試 ===
echo.

REM 1. 檢查 Copilot CLI 版本
echo 1. 檢查 Copilot CLI 版本...
copilot --version
if %ERRORLEVEL% NEQ 0 (
    echo X Copilot CLI 未安裝或無法執行
    exit /b 1
)

REM 2. 設定工作目錄
cd /d "C:\Users\cy540\OneDrive\桌面\Github CLI 自動跌代"

REM 3. 執行 SDK 基本連線測試
echo.
echo 2. 執行 SDK 基本連線測試...
go test -v -run TestSDKBasicConnection "./test/sdk_poc_test.go"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [32m✓ SDK 基本連線測試通過！[0m
) else (
    echo.
    echo [31mX SDK 基本連線測試失敗[0m
    echo 錯誤碼: %ERRORLEVEL%
)

REM 4. 執行 Session 建立測試
echo.
echo 3. 執行 Session 建立測試...
go test -v -run TestSDKSessionCreation "./test/sdk_poc_test.go"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [32m✓ Session 建立測試通過！[0m
) else (
    echo.
    echo [31mX Session 建立測試失敗[0m
)

REM 5. 顯示決策報告
echo.
echo 4. 執行決策測試...
go test -v -run TestSDKDecision "./test/sdk_poc_test.go"

echo.
echo === 測試完成 ===
echo.
echo 如果所有測試都通過，請執行以下命令繼續 SDK 整合：
echo go test -v ./test
pause
