package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func Confirm(message string) bool {
	// 创建标准输入读取器，避免缓冲区残留问题
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/N] ", message)

	for {
		// 读取整行输入（包含换行符，后续清理）
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("> 输入流异常，请重新输入：")
			continue
		}

		// 清理输入：去除首尾空格、换行符，转为大写
		choice := strings.ToUpper(strings.TrimSpace(input))

		// 有效选项判断
		switch choice {
		case "Y":
			return true
		case "N":
			return false
		default:
			// 无效选项给出明确提示
			fmt.Print("> 无效输入，请仅输入 Y 或 N：")
		}
	}
}
