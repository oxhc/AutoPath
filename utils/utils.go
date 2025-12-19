package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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

func GbkToUtf8(gbkBytes []byte) (string, error) {
	// 空输入直接返回空
	if len(gbkBytes) == 0 {
		return "", nil
	}
	// 创建GBK解码器
	decoder := simplifiedchinese.GBK.NewDecoder()
	// 转换全部字节
	result, _, err := transform.Bytes(decoder, gbkBytes)
	if err != nil {
		return "", fmt.Errorf("GBK解码失败: %w", err)
	}
	return string(result), nil
}

// 修复：UTF-8转GBK（完整处理字符串）
func Utf8ToGbk(utf8Str string) ([]byte, error) {
	if utf8Str == "" {
		return []byte{}, nil
	}
	encoder := simplifiedchinese.GBK.NewEncoder()
	result, _, err := transform.Bytes(encoder, []byte(utf8Str))
	if err != nil {
		return nil, fmt.Errorf("UTF-8编码为GBK失败: %w", err)
	}
	return result, nil
}
