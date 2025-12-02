package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver"
)

// cfgFile 用于存储命令行传入的配置文件路径
var cfgFile string

// NewAPIServerCommand 创建 apiserver 子命令
// 该命令用于启动 Web3 钱包 API 服务器，提供 RESTful API 接口
func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "Start the Go Web3 Wallet API server",
		Long: `Start the API server for the Web3 Wallet Backend.
The API server provides RESTful endpoints for user management,
wallet operations, and blockchain interactions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 调用 internal/apiserver 包中的运行逻辑
			return apiserver.Run(cfgFile)
		},
	}

	// 绑定配置文件参数到 apiserver 子命令
	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.yaml)")

	return cmd
}
