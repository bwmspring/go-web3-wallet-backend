package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 顶级配置结构体
type Config struct {
	Server   ServerConfig       `mapstructure:"server"`
	Database DatabaseConfig     `mapstructure:"database"`
	Chains   []BlockchainConfig `mapstructure:"chains"   yaml:"chains"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	JWTSecret   string `mapstructure:"jwt_secret"`
	JWTDuration string `mapstructure:"jwt_duration"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`

	// 连接池配置
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime_minutes"`
}

// BlockchainConfig 区块链配置
type BlockchainConfig struct {
	ChainID     uint   `yaml:"chain_id"     mapstructure:"chain_id"`
	Name        string `yaml:"name"         mapstructure:"name"`
	RPCUrl      string `yaml:"rpc_url"      mapstructure:"rpc_url"`
	ExplorerUrl string `yaml:"explorer_url" mapstructure:"explorer_url"`

	// 假设 YAML 中有更复杂的结构，例如 contract_addresses:
	ContractAddresses map[string]string `yaml:"contract_addresses" mapstructure:"contract_addresses"`
	IsTestnet         bool              `yaml:"is_testnet"         mapstructure:"is_testnet"` // 对应可选字段
}

// LoadConfigFromFile 加载并解析配置文件
func LoadConfigFromFile(configPath string) (*Config, error) {
	// 设置配置文件的名称和类型
	viper.SetConfigName("config") // 文件名（不带扩展名）
	viper.SetConfigType("yaml")   // 文件类型

	// 设置查找路径
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".") // 在当前目录下查找
	}

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("配置文件未找到: %s", err)
		}
		return nil, fmt.Errorf("读取配置文件失败: %s", err)
	}

	var cfg Config
	// 将读取的内容反序列化到结构体中
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置结构失败: %s", err)
	}

	// 配置加载成功，日志将在 logger 初始化后输出
	return &cfg, nil
}
