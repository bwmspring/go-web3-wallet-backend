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
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

---

-- DDL for wallets table
CREATE TABLE wallets (
    -- 主键
    id BIGSERIAL PRIMARY KEY,

    -- 外键：关联到 users 表的 id
    user_id BIGINT NOT NULL,

    -- 区块链地址：42 字符，唯一且非空
    address VARCHAR(42) UNIQUE NOT NULL,

    -- 加密密钥：存储 Keystore JSON
    encrypted_key TEXT NOT NULL,

    -- 派生路径：非空
    derivation_path VARCHAR(50) NOT NULL,

    -- 钱包名称
    name VARCHAR(255),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 软删除字段
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, 

    -- 外键约束
    CONSTRAINT fk_wallets_user 
        FOREIGN KEY (user_id) 
        REFERENCES users (id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

-- 创建索引以优化查询速度
CREATE UNIQUE INDEX idx_wallets_address ON wallets (address);
CREATE INDEX idx_wallets_user_id ON wallets (user_id);
CREATE INDEX idx_wallets_deleted_at ON wallets (deleted_at);


-- psql -h localhost -p 5432 -U web3_user -d web3_wallet_db
-- DROP TABLE IF EXISTS wallets;
-- DROP TABLE IF EXISTS users;