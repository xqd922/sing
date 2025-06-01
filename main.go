package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SingBoxConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FilePath    string `json:"file_path"`
	IsActive    bool   `json:"is_active"`
}

type SingBoxManager struct {
	app             fyne.App
	window          fyne.Window
	cmd             *exec.Cmd
	cmdMutex        sync.Mutex
	statusLabel     *widget.Label
	trafficLabel    *widget.Label
	connectionLabel *widget.Label
	logView         *widget.TextGrid
	stopChan        chan struct{}
	configList      []*SingBoxConfig
	configListView  *widget.List
	activeConfig    *SingBoxConfig
}

type SingBoxStatus struct {
	Memory        string `json:"memory"`
	Goroutines    int    `json:"goroutines"`
	Inbound       int    `json:"inbound"`
	Outbound      int    `json:"outbound"`
	UplinkSpeed   string `json:"uplink_speed"`
	DownlinkSpeed string `json:"downlink_speed"`
	UplinkTotal   string `json:"uplink_total"`
	DownlinkTotal string `json:"downlink_total"`
}

func NewSingBoxManager() *SingBoxManager {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Sing")
	w.Resize(fyne.NewSize(800, 600))

	manager := &SingBoxManager{
		app:      a,
		window:   w,
		stopChan: make(chan struct{}),
	}
	manager.setupUI()
	manager.loadConfigs()
	return manager
}

func (m *SingBoxManager) setupUI() {
	// 状态栏
	m.statusLabel = widget.NewLabel("未运行")
	m.trafficLabel = widget.NewLabel("流量: 无")
	m.connectionLabel = widget.NewLabel("连接: 无")

	// 日志视图
	m.logView = widget.NewTextGrid()
	m.logView.SetMinRowsVisible(20)

	// 控制按钮
	startBtn := widget.NewButtonWithIcon("启动", theme.MediaPlayIcon(), m.startSingBox)
	startBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButtonWithIcon("停止", theme.MediaStopIcon(), m.stopSingBox)
	stopBtn.Importance = widget.DangerImportance

	// 侧边栏
	sideBar := container.NewVBox(
		widget.NewLabel("Dashboard"),
		widget.NewButton("概览", func() {}),
		widget.NewButton("组", func() {}),
		widget.NewButton("日志", func() {}),
		widget.NewButton("配置", func() {}),
		widget.NewButton("设置", func() {}),
	)

	// 状态面板
	statusPanel := container.NewVBox(
		widget.NewLabel("状态"),
		m.statusLabel,
		m.trafficLabel,
		m.connectionLabel,
	)

	// 主内容区
	content := container.NewHSplit(
		sideBar,
		container.NewVSplit(
			statusPanel,
			widget.NewScrollContainer(m.logView),
		),
	)

	// 底部按钮栏
	bottomBar := container.NewHBox(
		startBtn,
		stopBtn,
	)

	// 设置窗口内容
	mainContent := container.NewBorder(
		nil,
		bottomBar,
		nil,
		nil,
		content,
	)

	m.window.SetContent(mainContent)
}

func (m *SingBoxManager) loadConfigs() {
	configDir := filepath.Join(m.getExeDir(), "config", "profiles")

	// 确保配置目录存在
	os.MkdirAll(configDir, os.ModePerm)

	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		log.Println("读取配置目录失败:", err)
		return
	}

	m.configList = []*SingBoxConfig{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			config := &SingBoxConfig{
				Name:        file.Name(),
				FilePath:    filepath.Join(configDir, file.Name()),
				Description: "用户配置文件",
			}
			m.configList = append(m.configList, config)
		}
	}

	// 如果没有配置文件，创建默认配置
	if len(m.configList) == 0 {
		defaultConfig := &SingBoxConfig{
			Name:        "default.json",
			FilePath:    filepath.Join(configDir, "default.json"),
			Description: "默认配置",
		}
		m.createDefaultConfig(defaultConfig)
		m.configList = append(m.configList, defaultConfig)
	}

	// 设置第一个配置为活跃配置
	m.configList[0].IsActive = true
	m.activeConfig = m.configList[0]
}

func (m *SingBoxManager) createDefaultConfig(config *SingBoxConfig) {
	defaultContent := `{
	"inbounds": [],
	"outbounds": [],
	"route": {}
}`
	err := ioutil.WriteFile(config.FilePath, []byte(defaultContent), 0644)
	if err != nil {
		log.Println("创建默认配置失败:", err)
	}
}

func (m *SingBoxManager) setupConfigUI() {
	m.configListView = widget.NewList(
		func() int { return len(m.configList) },
		func() fyne.CanvasObject {
			return widget.NewLabel("配置模板")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			config := m.configList[id]
			label.SetText(config.Name)
			if config.IsActive {
				label.TextStyle = fyne.TextStyle{Bold: true}
			}
		},
	)

	m.configListView.OnSelected = func(id widget.ListItemID) {
		// 取消之前的活跃状态
		for _, cfg := range m.configList {
			cfg.IsActive = false
		}

		// 设置新的活跃配置
		m.configList[id].IsActive = true
		m.activeConfig = m.configList[id]

		// 刷新列表显示
		m.configListView.Refresh()
	}

	addConfigBtn := widget.NewButtonWithIcon("添加配置", theme.ContentAddIcon(), func() {
		m.showAddConfigDialog()
	})

	editConfigBtn := widget.NewButtonWithIcon("编辑配置", theme.DocumentIcon(), func() {
		if m.activeConfig != nil {
			m.openConfigFile(m.activeConfig.FilePath)
		}
	})

	configPanel := container.NewBorder(
		nil,
		container.NewHBox(addConfigBtn, editConfigBtn),
		nil,
		nil,
		m.configListView,
	)

	// 将配置面板添加到主界面
	// 这里需要根据具体的 UI 布局进行调整
}

func (m *SingBoxManager) showAddConfigDialog() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("输入配置文件名（.json）")

	dialog.ShowForm("添加新配置", "确定", "取消",
		[]*widget.FormItem{
			widget.NewFormItem("配置名称", entry),
		},
		func(b bool) {
			if !b {
				return
			}

			configName := entry.Text
			if !strings.HasSuffix(configName, ".json") {
				configName += ".json"
			}

			configDir := filepath.Join(m.getExeDir(), "config", "profiles")
			configPath := filepath.Join(configDir, configName)

			// 创建空配置文件
			err := ioutil.WriteFile(configPath, []byte("{}"), 0644)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			newConfig := &SingBoxConfig{
				Name:     configName,
				FilePath: configPath,
			}
			m.configList = append(m.configList, newConfig)
			m.configListView.Refresh()
		},
		m.window)
}

func (m *SingBoxManager) openConfigFile(path string) {
	// 使用系统默认编辑器打开配置文件
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	if err != nil {
		dialog.ShowError(err, m.window)
	}
}

func (m *SingBoxManager) getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func (m *SingBoxManager) startSingBox() {
	if m.activeConfig == nil {
		dialog.ShowError(fmt.Errorf("未选择配置文件"), m.window)
		return
	}

	singBoxPath := filepath.Join(m.getExeDir(), "sing-box.exe")

	// 使用选定的配置文件启动
	cmd := exec.Command(singBoxPath, "run", "-c", m.activeConfig.FilePath)
	m.cmd = cmd

	// 捕获输出
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		m.updateStatus("创建输出管道失败", false)
		return
	}

	// 启动进程
	if err := m.cmd.Start(); err != nil {
		m.updateStatus("启动失败", false)
		return
	}

	// 更新状态
	m.updateStatus("运行中", true)

	// 异步读取输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			m.appendLog(line)
		}
	}()

	// 定期获取状态
	go m.monitorStatus()
}

func (m *SingBoxManager) stopSingBox() {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if m.cmd != nil && m.cmd.Process != nil {
		err := m.cmd.Process.Kill()
		if err != nil {
			m.updateStatus("停止失败", false)
		} else {
			m.updateStatus("已停止", false)
		}
		m.cmd = nil
		close(m.stopChan)
		m.stopChan = make(chan struct{})
	}
}

func (m *SingBoxManager) monitorStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			status, err := m.fetchSingBoxStatus()
			if err == nil {
				m.updateTrafficStatus(status)
			}
		}
	}
}

func (m *SingBoxManager) fetchSingBoxStatus() (*SingBoxStatus, error) {
	// 模拟获取状态，实际应替换为真实的 API 调用
	return &SingBoxStatus{
		Memory:        "24 MB",
		Goroutines:    198,
		Inbound:       55,
		Outbound:      49,
		UplinkSpeed:   "860 B/s",
		DownlinkSpeed: "7.2 kB/s",
		UplinkTotal:   "303 kB",
		DownlinkTotal: "1.1 MB",
	}, nil
}

func (m *SingBoxManager) updateStatus(status string, running bool) {
	m.app.Driver().CanvasForObject(m.statusLabel).Run(func() {
		m.statusLabel.SetText(status)
	})
}

func (m *SingBoxManager) updateTrafficStatus(status *SingBoxStatus) {
	m.app.Driver().CanvasForObject(m.trafficLabel).Run(func() {
		m.trafficLabel.SetText(fmt.Sprintf(
			"内存: %s | 协程: %d\n上行: %s | 下行: %s\n总上行: %s | 总下行: %s",
			status.Memory, status.Goroutines,
			status.UplinkSpeed, status.DownlinkSpeed,
			status.UplinkTotal, status.DownlinkTotal,
		))
		m.connectionLabel.SetText(fmt.Sprintf(
			"入站连接: %d | 出站连接: %d",
			status.Inbound, status.Outbound,
		))
	})
}

func (m *SingBoxManager) appendLog(message string) {
	m.app.Driver().CanvasForObject(m.logView).Run(func() {
		m.logView.SetText(m.logView.Text() + message + "\n")
	})
}

func (m *SingBoxManager) Run() {
	m.window.ShowAndRun()
}

func main() {
	manager := NewSingBoxManager()
	manager.Run()
}
