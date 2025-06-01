package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	singBoxCmd      *exec.Cmd
	singBoxCmdMutex sync.Mutex
	logOutput       *widget.Label
)

func showUserGuide(a fyne.App, w fyne.Window) {
	guide := "使用说明：\n" +
		"1. 请确保 sing-box.exe 与本程序在同一目录\n" +
		"2. 可从 https://github.com/SagerNet/sing-box/releases 下载 sing-box.exe\n" +
		"3. 下载 sing-box-windows-amd64.tar.gz 并解压\n" +
		"4. 将 sing-box.exe 放入本程序所在目录\n" +
		"5. 点击'启动'按钮运行 sing-box"

	dialog.ShowInformation("使用指南", guide, w)
}

func startSingBox() {
	singBoxCmdMutex.Lock()
	defer singBoxCmdMutex.Unlock()

	// 如果已经有运行的进程，先停止
	if singBoxCmd != nil && singBoxCmd.Process != nil {
		stopSingBox()
	}

	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		log.Println("获取可执行文件路径失败:", err)
		logOutput.SetText("获取可执行文件路径失败: " + err.Error())
		return
	}
	exeDir := filepath.Dir(exePath)

	// 寻找 sing-box 可执行文件
	singBoxPath := filepath.Join(exeDir, "sing-box.exe")

	// 检查 sing-box.exe 是否存在
	if _, err := os.Stat(singBoxPath); os.IsNotExist(err) {
		logOutput.SetText("错误：未找到 sing-box.exe，请查看使用指南")
		return
	}

	singBoxCmd = exec.Command(singBoxPath, "run")
	singBoxCmd.Dir = exeDir

	// 异步启动并捕获输出
	go func() {
		output, err := singBoxCmd.CombinedOutput()
		if err != nil {
			logOutput.SetText(fmt.Sprintf("启动失败: %v\n%s", err, string(output)))
		} else {
			logOutput.SetText(string(output))
		}
	}()
}

func stopSingBox() {
	singBoxCmdMutex.Lock()
	defer singBoxCmdMutex.Unlock()

	if singBoxCmd != nil && singBoxCmd.Process != nil {
		err := singBoxCmd.Process.Kill()
		if err != nil {
			logOutput.SetText("停止失败: " + err.Error())
		} else {
			logOutput.SetText("已停止 sing-box")
		}
		singBoxCmd = nil
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("Sing-Box 控制器")
	w.Resize(fyne.NewSize(400, 300))

	logOutput = widget.NewLabel("准备就绪")

	startBtn := widget.NewButton("启动 Sing-Box", func() {
		startSingBox()
	})

	stopBtn := widget.NewButton("停止 Sing-Box", func() {
		stopSingBox()
	})

	guideBtn := widget.NewButton("使用指南", func() {
		showUserGuide(a, w)
	})

	content := container.NewVBox(
		startBtn,
		stopBtn,
		guideBtn,
		logOutput,
	)

	w.SetContent(content)
	w.ShowAndRun()
}
