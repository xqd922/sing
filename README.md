# Sing-Box 启动器

## 使用说明

1. 下载 `singbox-gui.exe`
2. 将 `sing-box.exe` 放在同一目录
3. 双击 `singbox-gui.exe` 启动

## 功能

- 一键启动 Sing-Box
- 隐藏控制台窗口
- 启动错误提示
- 仅支持 Windows

## 下载

1. 从 [Releases](../../releases) 下载最新版本
2. 确保同目录有 `sing-box.exe`

## 编译

```bash
go build -ldflags "-H windowsgui" -o singbox-gui.exe
```

## 许可

MIT 许可 