# Sing-Box GUI 控制器

## 下载

### 方法 1：从 GitHub Releases 下载
1. 访问 [Releases 页面](../../releases)
2. 下载最新的 `singbox-gui-windows-amd64.exe`

### 方法 2：自行编译
1. 克隆仓库
2. 安装 Go 1.20+
3. 运行 `go build -ldflags "-H windowsgui" -o singbox-gui.exe`

## 使用说明

1. 下载 sing-box 可执行文件
   - 访问 https://github.com/SagerNet/sing-box/releases
   - 下载 `sing-box-windows-amd64.tar.gz`
   - 解压并将 `sing-box.exe` 放入与 `singbox-gui.exe` 相同的目录

2. 双击 `singbox-gui.exe` 启动
3. 使用界面上的"启动"和"停止"按钮控制 Sing-Box

## 编译说明

### 依赖要求
1. Go 1.20 或更高版本
2. MinGW-w64（可选，用于本地编译）
   - 下载地址：https://sourceforge.net/projects/mingw-w64/
   - 选择 x86_64-posix-seh 版本

### 编译步骤
```bash
# 克隆仓库
git clone https://github.com/yourusername/singbox-gui.git
cd singbox-gui

# 安装依赖
go mod tidy
go get -u fyne.io/fyne/v2
go get -u github.com/go-gl/gl
go get -u github.com/go-gl/glfw

# 编译
go build -ldflags "-H windowsgui" -o singbox-gui.exe
```

## 许可
基于 MIT 许可发布 