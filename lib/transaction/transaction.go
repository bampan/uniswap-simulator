package transaction

import (
	ui "uniswap-simulator/uint256"
)

type TransactionInput struct {
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	Amount0      string `json:"amount0"`
	Amount1      string `json:"amount1"`
	ID           string `json:"id"`
	SqrtPriceX96 string `json:"sqrtPriceX96"`
	Tick         int    `json:"tick"`
	TickLower    int    `json:"tickLower"`
	TickUpper    int    `json:"tickUpper"`
	Timestamp    int    `json:"timestamp"`
}

type Transaction struct {
	Type         string
	Amount       *ui.Int
	Amount0      *ui.Int
	Amount1      *ui.Int
	ID           string
	SqrtPriceX96 *ui.Int
	Tick         int
	TickLower    int
	TickUpper    int
	Timestamp    int
}
