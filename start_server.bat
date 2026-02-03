@echo off
echo ========================================
echo   工地报警通知系统 - WebSocket 服务器
echo ========================================
echo.

REM 设置环境变量（可选）
set JWT_SECRET=WanFang@JWT2024#SecureKey!ChangeInProduction
set PORT=8080

echo [启动] 正在启动服务器...
echo [配置] JWT Secret: %JWT_SECRET:~0,10%...
echo [配置] 端口: %PORT%
echo.

REM 启动服务器
go run cmd/server/main.go

pause
