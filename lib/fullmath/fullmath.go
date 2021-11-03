package fullmath

import (
	"math/big"
	ui "uniswap-simulator/uint256"
)

func MulDivRoundingUp(a, b, denominator *ui.Int) *ui.Int {

	product := new(big.Int).Mul(a.ToBig(), b.ToBig())
	dm_big := denominator.ToBig()
	result := new(big.Int).Div(product, dm_big)
	if new(big.Int).Rem(product, dm_big).Cmp(big.NewInt(0)) != 0 {
		result.Add(result, big.NewInt(1))
	}
	ret, _ := ui.FromBig(result)
	return ret
}
