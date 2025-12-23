package main

import (
	"autoPath/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findPythonInstallation 查找Windows系统中所有有效的Python安装路径
// 返回值：包含python.exe完整路径的字符串切片（去重后）
func findPythonInstallation() []string {
	// 存储最终的Python路径（去重）
	pythonPaths := make(map[string]struct{}) // 用map去重，value用空结构体节省内存

	// 步骤1：遍历常见的Python安装目录
	commonPaths := getCommonPythonPaths()
	for _, basePath := range commonPaths {
		// 检查基础路径是否存在
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}
		fmt.Println("> 正在搜索Python安装：", basePath)
		// 遍历基础路径下的子目录（处理Python版本号目录，如Python311）
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// 只处理文件，且文件名是python.exe或python3.exe
			if !info.IsDir() && (strings.EqualFold(info.Name(), "python.exe") || strings.EqualFold(info.Name(), "python3.exe")) {
				absPath, _ := filepath.Abs(path) // 转为绝对路径
				if strings.Contains(absPath, "Lib") {
					return filepath.SkipDir
				}
				pythonPaths[absPath] = struct{}{}
				return filepath.SkipAll
				//fmt.Printf("Found Python installation: %s\n", absPath)
			}
			return nil
		})
		if err != nil {
			//fmt.Println("> 搜索Python安装失败: ", err)
			continue
		}
	}

	// 步骤2：从PATH环境变量中查找Python
	pathList := utils.Where([]string{"python.exe", "python3.exe"})
	for _, path := range pathList {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		absPath, _ := filepath.Abs(path)
		pythonPaths[absPath] = struct{}{}
	}

	// 步骤3：将map转为切片返回
	result := make([]string, 0, len(pythonPaths))
	for path := range pythonPaths {
		result = append(result, path)
	}
	return result
}

func clearPythonPaths() {
	pythonPaths := utils.WhereWithoutFileName([]string{"python.exe", "python3.exe", "pip.exe"})
	if len(pythonPaths) == 0 {
		return
	}
	for _, path := range pythonPaths {
		fmt.Println("> 正在删除Python路径：", path)
		err := utils.DeleteDirFromUserPath(path)
		if err != nil {
			fmt.Println("> 删除Python路径失败：", err)
			return
		}
	}
	pathList, err := utils.GetPathList()
	if err != nil {
		fmt.Println("> 获取Path列表失败：", err)
		return
	}
	for _, path := range pathList {
		if strings.Contains(path, "WindowsApps") {
			fmt.Println("> 删除Path：", path)
			err := utils.DeleteDirFromUserPath(path)
			if err != nil {
				fmt.Println("> 删除Path失败：", err)
				return
			}
		}
	}
}

// getCommonPythonPaths 获取Windows下Python常见的安装根路径
func getCommonPythonPaths() []string {
	var paths []string

	// 1. 系统盘Program Files（64位）
	systemDrive := os.Getenv("SystemDrive") // 通常是C:
	if systemDrive != "" {
		paths = append(paths, filepath.Join(systemDrive, "Program Files", "Python"))
		paths = append(paths, filepath.Join(systemDrive, "Program Files (x86)", "Python")) // 32位Python
	}

	// 2. 当前用户级 AppData（优先用 LOCALAPPDATA 变量，更规范）
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		// 替代原有的 filepath.Join(userProfile, "AppData", "Local", ...)
		paths = append(paths, filepath.Join(localAppData, "Programs", "Python"))
	}

	// 3. 全局/所有用户级 AppData（补充！适配“为所有用户安装”的Python）
	allUsersProfile := os.Getenv("ALLUSERSPROFILE")
	if allUsersProfile != "" {
		paths = append(paths, filepath.Join(allUsersProfile, "Programs", "Python"))
	}

	// 4. 根目录下的Python（手动解压/绿色版）
	paths = append(paths, "C:\\Python")
	paths = append(paths, "D:\\Python")

	return paths
}

// extractPythonDirPath 提取Python所在目录和Scripts目录路径
// 参数：path - Python可执行文件的完整路径（如C:\Users\13910\AppData\Local\Programs\Python\Python311\python.exe）
// 返回值1：Python可执行文件所在的目录路径
// 返回值2：Python对应的Scripts目录路径（与Python目录同级，或在Python目录下）
func extractPythonDirPath(path string) (string, string) {
	// 1. 获取Python可执行文件所在的目录（文件路径的目录部分）
	// filepath.Dir会自动适配Windows(\)和Linux/Mac(/)路径分隔符
	pythonDir := filepath.Dir(path)

	// 2. 拼接Scripts目录路径（Scripts与python.exe同级，位于Python根目录下）
	// filepath.Join自动处理路径分隔符，保证跨平台兼容性
	scriptsDir := filepath.Join(pythonDir, "Scripts")

	return pythonDir, scriptsDir
}

func main() {
	op()
	utils.PressEnterToContinue()
}
func op() {
	pythonPaths := findPythonInstallation()
	if len(pythonPaths) == 0 {
		fmt.Println("未找到Python安装路径")
		return
	}
	fmt.Println("找到以下Python安装路径：")
	for i, path := range pythonPaths {
		fmt.Println(i+1, ":", path)
	}

	var pythonPath string
	for {
		fmt.Print("请选择要保留的Python版本(输入序号)：")
		var choice int
		_, err := fmt.Scanf("%d\n", &choice)
		if err != nil {
			fmt.Printf("> 输入的序号无效\n")
			continue
		}
		if choice < 1 || choice > len(pythonPaths) {
			fmt.Printf("> 输入的序号无效\n")
			continue
		}
		pythonPath = pythonPaths[choice-1]
		break
	}
	pythonDirPath, pythonScriptsDirPath := extractPythonDirPath(pythonPath)
	fmt.Println("> Python目录：", pythonDirPath)
	fmt.Println("> Python Scripts目录：", pythonScriptsDirPath)
	fmt.Println("正在将该Python版本添加到环境变量...")
	option := utils.Confirm(`正在清理旧的Python版本...是否继续`)
	if !option {
		fmt.Println("> 退出程序")
		return
	}
	clearPythonPaths()
	fmt.Println("> 删除完成")
	option = utils.Confirm(`是否添加该Python版本到环境变量`)
	if !option {
		fmt.Println("> 退出程序")
		return
	}
	err := utils.AddDirToUserPath(pythonDirPath)
	if err != nil {
		fmt.Println("> 添加Python路径失败：", err)
	}
	err = utils.AddDirToUserPath(pythonScriptsDirPath)
	if err != nil {
		fmt.Println("> 添加Python Scripts路径失败：", err)
	}
	fmt.Println("> 添加完成")
}
