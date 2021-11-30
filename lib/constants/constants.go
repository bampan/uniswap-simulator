package constants

import (
	"math/big"
	ui "uniswap-simulator/uint256"
)

var (
	NegativeOne, _ = ui.FromBig(big.NewInt(-1))
	Zero           = new(ui.Int)
	One            = new(ui.Int).SetOne()
	MaxUint256, _  = ui.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	// used in liquidity amount math
	Q128, _ = ui.FromHex("0x100000000000000000000000000000000")
	Q96     = new(ui.Int).Exp(ui.NewInt(2), ui.NewInt(96))
	Q192    = new(ui.Int).Exp(Q96, ui.NewInt(2))
	E6      = new(ui.Int).Exp(ui.NewInt(10), ui.NewInt(6))
	E18     = new(ui.Int).Exp(ui.NewInt(10), ui.NewInt(18))
)

var TickSpaces = map[int]int{
	500:   10,
	3000:  60,
	10000: 200,
}
