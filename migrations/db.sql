-- DDL for users table
CREATE TABLE users (
    -- 主键
    id BIGSERIAL PRIMARY KEY, 

    -- 用户名：唯一且非空
    username VARCHAR(50) UNIQUE NOT NULL,

    -- 密码哈希：非空
    password_hash TEXT NOT NULL,

    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 软删除字段
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 创建索引以优化查询速度
CREATE UNIQUE INDEX idx_users_username ON users (username);

---


-- psql -h localhost -p 5432 -U web3_user -d web3_wallet_db
-- DROP TABLE IF EXISTS wallets;
-- DROP TABLE IF EXISTS users;

-- 创建 wallets 表
CREATE TABLE wallets (
    -- GORM 默认字段
    id               BIGSERIAL PRIMARY KEY,
    created_at       TIMESTAMP WITH TIME ZONE,
    updated_at       TIMESTAMP WITH TIME ZONE,
    deleted_at       TIMESTAMP WITH TIME ZONE,

    -- 核心关联字段
    user_id          BIGINT NOT NULL,

    -- 关联到助记词表 (MnemonicSeed)
    mnemonic_id      BIGINT NOT NULL,

    -- 钱包元数据
    chain_id         BIGINT NOT NULL,
    name             VARCHAR(100) NOT NULL,
    address          VARCHAR(42) NOT NULL, -- 以太坊地址长度为 42 (0x + 40 hex chars)

    -- 安全信息
    encrypted_key    TEXT NOT NULL,        -- Keystore JSON，使用 TEXT 类型存储大文本
    derivation_path  VARCHAR(255) NOT NULL, -- BIP-44 路径

    -- 索引和约束
    -- 确保地址的唯一性，地址在不同链上通常是相同的，但这里我们假设地址在全球唯一或在用户级别唯一。
    -- 如果需要支持多链相同地址，则应是 UNIQUE(user_id, chain_id) 或 UNIQUE(address, chain_id)。
    -- 暂时保持 address 独立唯一索引。
    CONSTRAINT unique_address UNIQUE (address),

    -- 用户ID索引
    CONSTRAINT idx_wallets_user_id INDEX (user_id),

    -- 🌟 外键约束：关联到 mnemonic_seeds 表
    CONSTRAINT fk_wallets_mnemonic
        FOREIGN KEY (mnemonic_id)
        REFERENCES mnemonic_seeds(id)
        ON DELETE RESTRICT -- 保证助记词未被删除时，关联的钱包不能被删除
);

-- 为常用查询字段创建索引
CREATE INDEX idx_wallets_chain_id ON wallets (chain_id);