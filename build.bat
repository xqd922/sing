@echo off
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

echo 正在清理之前的编译...
go clean

echo 正在下载依赖...
go mod tidy

echo 正在编译应用程序...
go build -ldflags "-H windowsgui" -o singbox-gui.exe

if %errorlevel% neq 0 (
    echo 编译失败！
    exit /b %errorlevel%
)

echo 编译成功。请将 sing-box.exe 放入同一目录。 