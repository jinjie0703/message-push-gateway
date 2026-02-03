@echo off
echo ========================================
echo   工地报警通知系统 - WebSocket 服务器
echo ========================================
echo.

REM 设置环境变量（可选）
REM 建议通过环境变量注入密钥，不要在仓库中硬编码生产密钥。
REM 示例：
REM   set JWT_SECRET=CHANGE_ME_IN_PROD
REM   set PORT=8080

if "%PORT%"=="" set PORT=8080

echo [启动] 正在启动服务器...
echo [配置] 端口: %PORT%
echo.

REM 启动服务器
go run cmd/server/main.go

pause
