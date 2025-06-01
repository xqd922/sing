package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

func main() {
	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		showErrorMessage("获取程序路径失败: " + err.Error())
		return
	}
	exeDir := filepath.Dir(exePath)

	// 寻找 sing-box 可执行文件
	singBoxPath := filepath.Join(exeDir, "sing-box.exe")

	// 检查 sing-box.exe 是否存在
	if _, err := os.Stat(singBoxPath); os.IsNotExist(err) {
		showErrorMessage("未找到 sing-box.exe，请将 sing-box.exe 放在同一目录")
		return
	}

	// 创建进程
	cmd := exec.Command(singBoxPath, "run")
	cmd.Dir = exeDir

	// 隐藏控制台窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	// 启动进程
	err = cmd.Start()
	if err != nil {
		showErrorMessage("启动失败: " + err.Error())
		return
	}

	// 显示成功消息
	showSuccessMessage("Sing-Box 已成功启动")
}

func showSuccessMessage(message string) {
	title := "Sing-Box 控制器"
	windows.MessageBox(0, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title), windows.MB_OK|windows.MB_ICONINFORMATION)
}

func showErrorMessage(message string) {
	title := "Sing-Box 控制器 - 错误"
	windows.MessageBox(0, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title), windows.MB_OK|windows.MB_ICONERROR)
}
