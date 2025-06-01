# Sing-Box 控制器

## 功能特性

- 图形化界面管理 Sing-Box
- 实时状态监控
  - 内存使用
  - 协程数量
  - 网络流量统计
  - 连接数监控
- 启动/停止控制
- 日志实时显示
- 深色主题
- 仅支持 Windows 平台

## 使用说明

1. 下载 `sing.exe`
2. 将 `sing-box.exe` 放在同一目录
3. 双击 `sing.exe` 启动控制器

## 系统要求

- Windows 10 或更高版本
- Go 1.20+
- sing-box 可执行文件

## 编译

```bash
go mod tidy
go build -o sing.exe
```

## 许可

MIT 许可 