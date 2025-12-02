package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 代表所有子命令的根命令
var rootCmd = &cobra.Command{
	Use:   "wallet-backend",
	Short: "Wallet-Backend is the command line tool for the Web3 Wallet System.",
	Long: `Wallet-Backend contains all subcommands to run various services, 
such as 'apiserver', 'authzserver', and 'migrate'.
`,
	// 根命令只用于显示帮助信息
}

// Execute 是 main.main() 调用的唯一函数
// 它注册所有子命令并执行根命令
func Execute() {
	// 注册所有子命令
	rootCmd.AddCommand(NewAPIServerCommand())
	// 未来可以在这里添加其他子命令，例如：
	// rootCmd.AddCommand(NewAuthzServerCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
