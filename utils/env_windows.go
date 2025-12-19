package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

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
	//fmt.Printf("out: %s", output)

	// 解码GBK输出为UTF-8（解决中文乱码）
	utf8Output, err := GbkToUtf8(output)
	if err != nil {
		return "", fmt.Errorf("解码reg输出失败: %w", err)
	}

	//fmt.Printf("utf8: %s", utf8Output)

	// 解析reg输出，兼容REG_SZ和REG_EXPAND_SZ类型
	regOutput := string(utf8Output)
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

	//fmt.Printf("userPath: %s\n", userPath)

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
