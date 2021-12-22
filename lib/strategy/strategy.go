package strategy

import (
	"uniswap-simulator/lib/pool"
	ui "uniswap-simulator/uint256"
)

type Position struct {
	amount    *ui.Int
	tickLower int
	tickUpper int
}

type Strategy interface {
	Init() (*ui.Int, *ui.Int)
	Rebalance() (*ui.Int, *ui.Int)
	BurnAll() (*ui.Int, *ui.Int)
	GetPool() *pool.Pool
	GetAmounts() (*ui.Int, *ui.Int)
}
