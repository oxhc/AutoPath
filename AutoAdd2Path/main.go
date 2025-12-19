package main

import (
	"autoPath/utils"
	"fmt"
)

func main() {
	op()
	utils.PressEnterToContinue()
}

func op() {
	// 1. 获取当前exe所在目录
	exeDir, err := utils.GetExeDir()
	if err != nil {
		fmt.Printf("获取exe目录失败: %v\n", err)
		return
	}
	fmt.Printf("当前目录: %s\n", exeDir)

	// 2. 检查是否已在Path中
	exists, err := utils.IsDirInPath(exeDir)
	if err != nil {
		fmt.Printf("检查目录是否在Path中失败: %v\n", err)
		return
	}
	if exists {
		fmt.Println("目录已存在于用户Path环境变量中，无需添加")
		return
	}

	// 3. 添加到用户Path
	err = utils.AddDirToUserPath(exeDir)
	if err != nil {
		fmt.Printf("添加目录到Path失败: %v\n", err)
		return
	}
	fmt.Println("目录已成功添加到用户Path环境变量！")
	fmt.Println("注意：已打开的命令行窗口需重启才能生效，新窗口直接生效")
}
