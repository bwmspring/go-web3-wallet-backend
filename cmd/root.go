package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wallet-backend",
	Short: "A high-performance Go Web3 Wallet Backend",
	Long:  `The backend service for managing keys, querying blockchain data, and sending transactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认行为：如果直接执行，则显示帮助信息
		cmd.Help()
	},
}

// Execute 应用程序的入口函数
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
