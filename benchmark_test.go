package main

import (
	"testing"
	"uniswap-simulator/lib/constants"
	ppool "uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickdata"
	"uniswap-simulator/lib/tickmath"
	"uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

var transactions []transaction.Transaction

func init() {
	transactions = getTransactions()

}

func Benchmark_run(bench *testing.B) {
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96, _ := ui.FromHex("0x42919A3B4E1F2E279AB5FE196328")
	tickSpacing := constants.TickSpaces[fee]
	liquidity := ui.NewInt(0)
	tickCurrent := tickmath.TM.GetTickAtSqrtRatio(sqrtX96)
	tickData := tickdata.NewTickData(tickSpacing)
	pool := &ppool.Pool{
		token0,
		token1,
		fee,
		sqrtX96,
		liquidity,
		tickSpacing,
		tickCurrent,
		tickData,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		for _, trans := range transactions {
			switch trans.Type {
			case "mint":
				pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
			case "burn":
				pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
			case "swap":
				if trans.Amount0.Sign() > 0 {
					pool.GetOutputAmount(trans.Amount0, token0, ui.NewInt(0))
				} else if trans.Amount1.Sign() > 0 {
					pool.GetOutputAmount(trans.Amount1, token1, ui.NewInt(0))
				}
			}
		}
	}
}
