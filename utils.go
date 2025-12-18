package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// PressEnterToContinue 等待用户按回车键
func PressEnterToContinue() {
	fmt.Print("按回车键关闭此窗口...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
