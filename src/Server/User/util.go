package userServe

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileDataMkdir 创建目录
// 假设FileServer/bin/main.exe运行Mkdir("./FileData", username)
// 则其创建目录在该目录上一级的FileData文件夹下
// 即最终创建：FileServer/FileData/username-2023-7-6
// FileDataMkdir("./FileData", username)
func FileDataMkdir(basePath string, username string) string {
	//	1.获取当前时间,并且格式化时间
	folderName := username + "-" + time.Now().Format("2006-01-02")
	folderPath := filepath.Join(basePath, folderName)
	// 使用mkdirall会创建多层级目录
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		fmt.Println("Mkdir MkdirAll ERROR:", err)
		return ""
	}
	return folderPath
}
