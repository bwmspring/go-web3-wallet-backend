package crypto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

// KeyManager 定义了密钥管理工具的接口
type KeyManager interface {
	GenerateMnemonic() (string, error)
	DeriveKeyFromMnemonic(mnemonic string, path string) (privateKeyHex string, address string, err error)
	EncryptPrivateKey(privateKeyHex string, password string) (keystoreJSON string, err error)
	DecryptKeystore(keystoreJSON string, password string) (privateKeyHex string, err error)
}

// keyManager 是 KeyManager 接口的实际实现结构体
type keyManager struct{}

// NewKeyManager 是构造函数，返回导出的接口类型
func NewKeyManager() KeyManager {
	return &keyManager{}
}

// GenerateMnemonic 生成一个新的BIP-39助记词 (128位熵，12个单词)
func (m *keyManager) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate random entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}
	return mnemonic, nil
}

// DeriveKeyFromMnemonic 根据助记词和派生路径，生成私钥和地址
// 采用标准的 go-ethereum-hdwallet 库，确保 BIP-44 规范的正确实现。
func (m *keyManager) DeriveKeyFromMnemonic(
	mnemonic string,
	path string,
) (privateKeyHex string, address string, err error) {

	// 1. 创建 HD 钱包对象
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", "", fmt.Errorf("failed to create HD wallet from mnemonic: %w", err)
	}

	// 2. 派生路径
	derivationPath := hdwallet.MustParseDerivationPath(path)

	// 3. 派生账户
	account, err := wallet.Derive(derivationPath, true) // true表示使用硬化派生
	if err != nil {
		return "", "", fmt.Errorf("failed to derive account from path '%s': %w", path, err)
	}

	// 4. 提取私钥和地址
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return "", "", fmt.Errorf("failed to get private key for account: %w", err)
	}

	address = account.Address.Hex()
	privateKeyHex = hexutil.Encode(crypto.FromECDSA(privateKey))

	return privateKeyHex, address, nil
}

// EncryptPrivateKey 用密码将私钥加密为Keystore JSON
func (m *keyManager) EncryptPrivateKey(privateKeyHex string, password string) (keystoreJSON string, err error) {

	// 1. 预处理私钥并转换为 ECDSA
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKeyBytes, err := hexutil.Decode("0x" + privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %w", err)
	}
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	keyToEncrypt := &keystore.Key{
		// Address 字段必须是 common.Address 类型，通过公钥派生
		Address: crypto.PubkeyToAddress(privateKey.PublicKey),
		// PrivateKey 字段必须是 *ecdsa.PrivateKey 类型
		PrivateKey: privateKey,

		// 通常 Key 结构体还需要 ID 字段，这里简化，依赖 EncryptKey 内部处理
		// ID: uuid.New(),
	}

	// 2. 使用 go-ethereum 标准 Keystore (Scrypt KDF) 进行加密
	keyjson, err := keystore.EncryptKey(
		keyToEncrypt,
		password,
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)
	if err != nil {
		return "", fmt.Errorf("keystore encryption failed: %w", err)
	}

	return string(keyjson), nil
}

// DecryptKeystore 用密码将Keystore JSON解密为私钥
func (m *keyManager) DecryptKeystore(keystoreJSON string, password string) (privateKeyHex string, err error) {

	// 1. 解密
	key, err := keystore.DecryptKey([]byte(keystoreJSON), password)
	if err != nil {
		if strings.Contains(err.Error(), "authentication failed") {
			return "", errors.New("wallet password incorrect") // Service 层将判断并返回 ErrPasswordIncorrect
		}
		return "", fmt.Errorf("keystore decryption failed: %w", err)
	}

	// 2. 转换为 Hex 字符串
	privateKeyHex = hexutil.Encode(crypto.FromECDSA(key.PrivateKey))

	return privateKeyHex, nil
}
