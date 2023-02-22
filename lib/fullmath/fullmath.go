package fullmath

import (
	cons "github.com/ftchann/uniswap-simulator/lib/constants"

	ui "github.com/holiman/uint256"
)

func MulDivRoundingUp(a, b, denominator *ui.Int) *ui.Int {
	if a.IsZero() || b.IsZero() {
		return ui.NewInt(0)
	}
	result := MulDiv(a, b, denominator)
	rem := new(ui.Int).MulMod(a, b, denominator)
	if !rem.IsZero() {
		result.Add(result, cons.One)
	}
	return result
}

func MulDiv(a, b, denominator *ui.Int) *ui.Int {
	result, overflow := new(ui.Int).MulDivOverflow(a, b, denominator)
	if overflow {
		panic("mulDiv overflow")
	}
	return result
}
