package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"strconv"
	"uniswap-simulator/lib/constants"
	cons "uniswap-simulator/lib/constants"
	ppool "uniswap-simulator/lib/pool"
	strat "uniswap-simulator/lib/strategy"
	"uniswap-simulator/lib/strategyData"
	"uniswap-simulator/lib/tickdata"
	"uniswap-simulator/lib/tickmath"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

func main() {
	fmt.Println("Start")
	transactions := getTransactions()
	fmt.Println("Transactions: ", len(transactions))
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96big, _ := new(big.Int).SetString("1350174849792634181862360983626536", 10)
	sqrtX96, _ := ui.FromBig(sqrtX96big)
	tickSpacing := constants.TickSpaces[fee]
	liquidity := ui.NewInt(0)
	tickCurrent := tickmath.TM.GetTickAtSqrtRatio(sqrtX96)
	tickData := tickdata.NewTickData(tickSpacing)

	strategydata := &strategyData.StrategyData{
		ui.NewInt(0),
		ui.NewInt(0),
		tickdata.NewTickData(tickSpacing),
		ui.NewInt(0),
	}

	pool := &ppool.Pool{
		token0,
		token1,
		fee,
		sqrtX96,
		liquidity,
		tickSpacing,
		tickCurrent,
		tickData,
		strategydata,
	}
	startAmount0 := ui.NewInt(1_000_000)                                // 1 USDC
	startAmount1big, _ := new(big.Int).SetString("290000000000000", 10) // 290_000_000_000_000 wei ~= 1 USD worth of ETH
	startAmount1, _ := ui.FromBig(startAmount1big)

	strategy := strat.NewStrategy(startAmount0, startAmount1, pool, 10)

	//starttime := transactions[0].Timestamp
	//// 30 days
	//nextUpdate := starttime + (60 * 60 * 24 * 199)
	//fmt.Printf("NextUpdate: %d\n", nextUpdate)
	// 24 hours
	//updateInterval := 60 * 60 * 24 + starttime
	for _, trans := range transactions {
		//if i+1 == 1 {
		//	strategy.Rebalance()
		//}
		//if trans.Timestamp > nextUpdate {
		//	strategy.Rebalance()
		//	nextUpdate += updateInterval
		//}
		var amount1, amount0 *ui.Int
		switch trans.Type {

		case "Mint":
			if !trans.Amount.IsZero() {
				amount0, amount1 = strategy.Pool.MintStrategy(trans.TickLower, trans.TickUpper, trans.Amount)
				if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) {
					fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
					panic("what")
				}
			}

		case "Burn":
			if !trans.Amount.IsZero() {
				amount0, amount1 = strategy.Pool.BurnStrategy(trans.TickLower, trans.TickUpper, trans.Amount)
				if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) {
					fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
					panic("what")
				}
			}

		case "Swap":

			if trans.Amount0.Sign() > 0 {
				if trans.UseX96 {
					amount0, amount1 = strategy.Pool.GetOutputAmount(trans.Amount0, token0, trans.SqrtPriceX96)
				} else {
					amount0, amount1 = strategy.Pool.GetOutputAmount(trans.Amount0, token0, cons.Zero)
				}
			} else if trans.Amount1.Sign() > 0 {
				if trans.UseX96 {
					amount0, amount1 = strategy.Pool.GetOutputAmount(trans.Amount1, token1, trans.SqrtPriceX96)
				} else {
					amount0, amount1 = strategy.Pool.GetOutputAmount(trans.Amount1, token1, cons.Zero)
				}
			}
			if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) || !trans.SqrtPriceX96.Eq(strategy.Pool.SqrtRatioX96) || trans.Tick != strategy.Pool.TickCurrent {
				fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
				fmt.Printf("%d %d %d %d\n", trans.SqrtPriceX96, strategy.Pool.SqrtRatioX96, trans.Tick, strategy.Pool.TickCurrent)
				panic("what")
			}
		case "Flash":
			strategy.Pool.Flash(trans.Amount0, trans.Amount1)
		}
	}
	//fmt.Printf("Start_Amount0: %d Start_Amount1: %d \n", startAmount0, startAmount1)
	//amount0, amount1 := new(ui.Int), new(ui.Int)
	//strategy.BurnAll()
	//fmt.Printf("EndAmount0: %d EndAmount1: %d \n", strategy.Amount0, strategy.Amount1)
	//fmt.Printf("FeeAmount0: %d FeeAmount1: %d \n", strategy.Pool.StrategyData.FeeAmount0, strategy.Pool.StrategyData.FeeAmount1)
	//amount0.Add(strategy.Amount0, strategy.Pool.StrategyData.FeeAmount0)
	//amount1.Add(strategy.Amount1, strategy.Pool.StrategyData.FeeAmount1)
	//
	//fmt.Printf("Amount0Total: %d Amount1Total: %d \n", amount0, amount1)
	//fmt.Printf("%d \n", strategy.Pool.StrategyData.Liquidity)

}

func getTransactions() []ent.Transaction {
	filename := "transactions.json"
	filepath := path.Join("data", filename)
	file, err := os.Open(filepath)
	check(err)
	value, err := ioutil.ReadAll(file)
	check(err)
	var transactionsInput []ent.TransactionInput
	err = json.Unmarshal([]byte(value), &transactionsInput)
	check(err)
	var transactions []ent.Transaction
	for _, transIn := range transactionsInput {
		useX96, _ := strconv.ParseBool(transIn.UseX96)
		trans := ent.Transaction{
			transIn.Type,
			stringToUint256(transIn.Amount),
			stringToUint256(transIn.Amount0),
			stringToUint256(transIn.Amount1),
			transIn.ID,
			stringToUint256(transIn.SqrtPriceX96),
			transIn.Tick,
			transIn.TickLower,
			transIn.TickUpper,
			transIn.Timestamp,
			useX96,
		}
		transactions = append(transactions, trans)
	}
	return transactions
}

func stringToUint256(amount string) *ui.Int {
	bigint := new(big.Int)
	bigint.SetString(amount, 10)
	uint256, _ := ui.FromBig(bigint)
	return uint256
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
