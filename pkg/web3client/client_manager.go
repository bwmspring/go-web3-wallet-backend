package web3client

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// ClientManager 定义了区块链客户端的接口，负责管理与不同链的连接。
type ClientManager interface {
	GetClient(chainID uint) (*ethclient.Client, error)
	GetBalanceByAddress(ctx context.Context, chainID uint, address string) (*big.Int, error)
}

// clientManager 实现了 ClientManager 接口
type clientManager struct {
	// connections 存储 {ChainID: *ethclient.Client} 的连接池
	connections map[uint]*ethclient.Client
	mu          sync.RWMutex // 保护 connections map
}

// NewClientManager 构造函数，根据配置切片初始化并连接所有配置的区块链节点。
// 仅依赖于 []config.BlockchainConfig，保持模块的内聚性。
func NewClientManager(chainConfigs []config.BlockchainConfig) (ClientManager, error) {
	manager := &clientManager{
		connections: make(map[uint]*ethclient.Client),
	}

	// 核心修正：遍历配置切片，初始化连接
	for _, chainCfg := range chainConfigs {
		chainID := chainCfg.ChainID
		rpcURL := chainCfg.RPCUrl

		// 忽略没有配置 URL 的链
		if rpcURL == "" {
			logger.Logger.Warn("RPC URL is empty, skipping chain initialization", zap.Uint("chain_id", chainID))
			continue
		}

		if err := manager.initializeClient(chainID, rpcURL); err != nil {
			// 初始化失败是致命错误，返回给 cmd/server 决定是否终止应用
			logger.Logger.Error("Failed to initialize RPC client",
				zap.Uint("chain_id", chainID),
				zap.String("url", rpcURL),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to initialize client for chain %d: %w", chainID, err)
		}
	}

	return manager, nil
}

// initializeClient 建立单个 RPC 连接并存入连接池
func (m *clientManager) initializeClient(chainID uint, rpcURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC URL %s: %w", rpcURL, err)
	}

	m.connections[chainID] = client
	logger.Logger.Info("Successfully connected to RPC", zap.Uint("chain_id", chainID), zap.String("url", rpcURL))
	return nil
}

// GetClient 根据 chainID 从连接池获取 *ethclient.Client 实例
func (m *clientManager) GetClient(chainID uint) (*ethclient.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.connections[chainID]
	if !ok {
		return nil, errors.New("chain ID not found in configuration or failed to connect")
	}
	return client, nil
}

// GetBalanceByAddress 查询指定地址在指定链上的余额
func (m *clientManager) GetBalanceByAddress(ctx context.Context, chainID uint, address string) (*big.Int, error) {
	client, err := m.GetClient(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for chain %d: %w", chainID, err)
	}

	// 设置查询超时 (保护系统资源)
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	addr := common.HexToAddress(address)

	// 查询余额 (nil 表示查询最新的区块余额)
	balance, err := client.BalanceAt(timeoutCtx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance for address %s: %w", address, err)
	}

	return balance, nil
}
