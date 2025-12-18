package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// 系统调用相关常量（Windows）
const (
	hwndBroadcast   = 0xFFFF
	wmSettingchange = 0x001A
	smtoAbortifhung = 0x0002
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout = user32.NewProc("SendMessageTimeoutW")
)

// GetExeDir 获取当前exe所在的绝对目录（处理符号链接）
func GetExeDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取exe路径失败: %w", err)
	}
	// 解析符号链接，获取真实路径
	absExePath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("解析符号链接失败: %w", err)
	}
	// 提取目录并清理路径
	exeDir := filepath.Clean(filepath.Dir(absExePath))
	return exeDir, nil
}

// readUserPathFromReg 从注册表读取用户级Path环境变量
// 返回值：path字符串（空字符串表示无该键值），错误
func readUserPathFromReg() (string, error) {
	if runtime.GOOS != "windows" {
		return "", errors.New("仅支持Windows系统")
	}

	// 执行reg查询命令
	cmd := exec.Command("reg", "query", "HKCU\\Environment", "/v", "Path")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 检查是否是"找不到键值"的错误（退出码1，输出包含"找不到项目"）
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 && strings.Contains(strings.ToLower(string(output)), "找不到项目") {
			return "", nil // 无Path键值，返回空字符串
		}
		return "", fmt.Errorf("查询注册表失败: %w, 输出: %s", err, string(output))
	}

	// 解析reg输出，兼容REG_SZ和REG_EXPAND_SZ类型
	regOutput := string(output)
	re := regexp.MustCompile(`Path\s+(REG_SZ|REG_EXPAND_SZ)\s+(.*)`)
	matches := re.FindStringSubmatch(regOutput)
	if len(matches) < 3 {
		return "", nil // 解析失败，视为空Path
	}
	return strings.TrimSpace(matches[2]), nil
}

// IsDirInPath 检查指定目录是否已存在于用户Path环境变量中（从注册表读取）
func IsDirInPath(targetDir string) (bool, error) {
	if runtime.GOOS != "windows" {
		return false, errors.New("仅支持Windows系统")
	}

	// 从注册表读取原始用户Path
	userPath, err := readUserPathFromReg()
	if err != nil {
		return false, fmt.Errorf("从注册表读取Path失败: %w", err)
	}
	if userPath == "" {
		return false, nil // Path为空，目录肯定不存在
	}

	// 拆分Path为多个目录（Windows用;分隔）
	pathDirs := strings.Split(userPath, ";")
	// 统一路径格式（转小写、替换/为\、清理路径）
	targetDir = filepath.Clean(strings.ToLower(targetDir))

	for _, dir := range pathDirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		// 统一待检查目录的格式
		cleanDir := filepath.Clean(strings.ToLower(dir))
		//fmt.Printf("%v vs %v\n", cleanDir, targetDir)
		if cleanDir == targetDir {
			return true, nil
		}
	}
	return false, nil
}

// AddDirToUserPath 将目录添加到用户级Path环境变量（完全基于注册表操作）
func AddDirToUserPath(targetDir string) error {
	if runtime.GOOS != "windows" {
		return errors.New("仅支持Windows系统")
	}

	// 1. 检查是否已存在
	exists, err := IsDirInPath(targetDir)
	if err != nil {
		return fmt.Errorf("检查目录是否存在失败: %w", err)
	}
	if exists {
		return nil // 已存在，无需添加
	}

	// 2. 从注册表读取当前Path
	currentPath, err := readUserPathFromReg()
	if err != nil {
		return fmt.Errorf("读取当前Path失败: %w", err)
	}

	// 3. 拼接新Path
	var newPath string
	if currentPath == "" {
		newPath = targetDir // Path为空，直接设置为目标目录
	} else {
		newPath = fmt.Sprintf("%s;%s", currentPath, targetDir) // 追加到现有Path
	}

	// 4. 写入注册表（用户级Environment）
	// reg add "HKCU\Environment" /v Path /t REG_EXPAND_SZ /d "新Path值" /f
	cmd := exec.Command("reg", "add", "HKCU\\Environment", "/v", "Path", "/t", "REG_EXPAND_SZ", "/d", newPath, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("写入注册表失败: %w, 输出: %s", err, string(output))
	}

	// 5. 发送系统消息，让环境变量生效
	err = sendEnvironmentChangeMessage()
	if err != nil {
		return fmt.Errorf("发送环境变量更新消息失败: %w", err)
	}

	return nil
}

// sendEnvironmentChangeMessage 发送Windows系统消息，通知环境变量已更新
func sendEnvironmentChangeMessage() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// 转换字符串为LPCWSTR
	lpParam, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return err
	}

	// 调用SendMessageTimeoutW
	ret, _, err := sendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingchange),
		0,
		uintptr(unsafe.Pointer(lpParam)),
		uintptr(smtoAbortifhung),
		uintptr(5000), // 超时5秒
		0,
	)
	if ret == 0 {
		return fmt.Errorf("SendMessageTimeout失败: %w", err)
	}
	return nil
}

func PressEnterToContinue() {
	fmt.Print("按回车键关闭此窗口...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func op() {
	// 1. 获取当前exe所在目录
	exeDir, err := GetExeDir()
	if err != nil {
		fmt.Printf("获取exe目录失败: %v\n", err)
		return
	}
	fmt.Printf("当前目录: %s\n", exeDir)

	// 2. 检查是否已在Path中
	exists, err := IsDirInPath(exeDir)
	if err != nil {
		fmt.Printf("检查目录是否在Path中失败: %v\n", err)
		return
	}
	if exists {
		fmt.Println("目录已存在于用户Path环境变量中，无需添加")
		return
	}

	// 3. 添加到用户Path
	err = AddDirToUserPath(exeDir)
	if err != nil {
		fmt.Printf("添加目录到Path失败: %v\n", err)
		return
	}
	fmt.Println("目录已成功添加到用户Path环境变量！")
	fmt.Println("注意：已打开的命令行窗口需重启才能生效，新窗口直接生效")
}

func main() {
	op()
	PressEnterToContinue()
}
