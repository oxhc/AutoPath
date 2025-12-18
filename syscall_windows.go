//go:build windows
// +build windows

package main

import (
	"fmt"
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

// sendEnvironmentChangeMessage 发送Windows系统消息，通知环境变量已更新
func sendEnvironmentChangeMessage() error {
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
