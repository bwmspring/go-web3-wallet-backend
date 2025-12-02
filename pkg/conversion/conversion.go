package conversion

import (
	"errors"
	"math/big"

	"github.com/shopspring/decimal"
)

// 以太坊单位：1 Ether = 10^18 Wei
var ethPrecision = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)

// WeiToEther 将 Wei (大整数) 转换为 Ether (人类可读的字符串，使用 decimal 类型保持精度)
// Wei is a *big.Int, Ether is a string representation of decimal.
func WeiToEther(wei *big.Int) decimal.Decimal {
	// 将 Wei 转换为 decimal
	weiDecimal := decimal.NewFromBigInt(wei, 0)

	// 除以 10^18 (ethPrecision)，得到 Ether
	// 我们使用 1e18 的 decimal 表示来除，避免 *big.Int 和 decimal 混合计算
	ethPrecisionDecimal := decimal.NewFromBigInt(ethPrecision, 0)

	// 计算结果
	ether := weiDecimal.Div(ethPrecisionDecimal)

	// 返回结果，通常保留 18 位精度或更多
	return ether
}

// ToWei 将 Ether (字符串或 decimal) 转换为 Wei (*big.Int)
// amountDecimal is the amount in Ether (e.g., "1.5"), Wei is a *big.Int
func ToWei(amountDecimal decimal.Decimal) (*big.Int, error) {

	// 乘以 10^18 (ethPrecision)
	weiDecimal := amountDecimal.Mul(decimal.NewFromBigInt(ethPrecision, 0))

	// 检查是否有小数部分（即 Wei 应该是一个整数）
	if weiDecimal.String() != weiDecimal.Floor().String() {
		// 这是理论上不应该发生的，除非输入的 Ether 小数位数过多导致精度溢出
		return nil, errors.New("wei conversion resulted in non-integer value")
	}

	// 转换为 *big.Int
	wei := weiDecimal.BigInt()

	return wei, nil
}
