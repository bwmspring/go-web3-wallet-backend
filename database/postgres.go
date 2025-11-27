package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// DB 是全局导出的 GORM 数据库连接实例
var DB *gorm.DB

// InitDatabase 初始化数据库连接和配置
func InitDatabase(cfg *config.Config) {
	// 构造 PostgreSQL DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
	)

	// GORM 配置
	gormConfig := &gorm.Config{
		// 启用软删除
		SkipDefaultTransaction: true,
		// 命名策略：禁用复数表名 (但我们已经接受了 GORM 的默认复数约定，这里保留默认即可)
		NamingStrategy: schema.NamingStrategy{
			// TablePrefix: "t_", // 可以在此添加表前缀
			SingularTable: false, // 保持 GORM 默认：使用复数表名
		},
		// Logger: 可以在此集成 Zap Logger 到 GORM
	}

	var err error
	// 尝试连接数据库
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		// 使用 Zap Logger 报告致命错误
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Logger.Fatal("Failed to get generic database object", zap.Error(err))
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	// SetConnMaxLifetime 设置连接可重用的最大时间
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
}

// GetDB 返回全局的 GORM 数据库连接实例，供 Repository 层使用。
func GetDB() *gorm.DB {
	// 在生产环境中，此处可能需要检查 DB 是否为 nil，但由于我们在 InitDatabase 中使用了 Fatal，
	// 理论上如果程序运行到这里，DB 必定已成功初始化。
	return DB
}
