package main

import (
	"math/big"
	"testing"
	"uniswap-simulator/lib/executor"
	ppool "uniswap-simulator/lib/pool"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

var (
	transactions   []ent.Transaction
	pool           *ppool.Pool
	startAmount0   *ui.Int
	startAmount1   *ui.Int
	startTime      int
	updateInterval int
)

func init() {

	transactions = getTransactions()
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96big, _ := new(big.Int).SetString("1350174849792634181862360983626536", 10)
	sqrtX96, _ := ui.FromBig(sqrtX96big)

	pool = ppool.NewPool(token0, token1, fee, sqrtX96)

	startAmount0 = ui.NewInt(1_000_000) // 1 USDC
	// From the Price One month in
	startAmount1big, _ := new(big.Int).SetString("366874042000000", 10) // 366874042000000 wei ~= 1 USD worth of ETH
	startAmount1, _ = ui.FromBig(startAmount1big)

	startTime = transactions[0].Timestamp + 60*60*24*30
	updateInterval = 60 * 60 * 24
}

func Benchmark_run(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		strategy := strat.NewConstantIntervallStrategy(startAmount0, startAmount1, pool, 10)
		excecution := executor.CreateExecution(strategy, startTime, updateInterval, transactions)

		excecution.Run()
	}

}
