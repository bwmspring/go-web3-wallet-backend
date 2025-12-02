package main

import (
	"fmt"
	"os"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver"
)

// main 是 apiserver 的独立启动入口
// 实际部署时，可以通过运行 go build -o apiserver cmd/apiserver/main.go 来编译这个二进制文件
func main() {
	// 1. 创建 Apiserver 启动命令
	cmd := apiserver.NewAPIServerCommand()

	// 2. 执行命令，处理配置加载、依赖注入和服务器运行
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error during apiserver execution: %v\n", err)
		os.Exit(1)
	}
}
