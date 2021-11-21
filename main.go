package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path"
	"time"
	"uniswap-simulator/lib/constants"
	cons "uniswap-simulator/lib/constants"
	ppool "uniswap-simulator/lib/pool"
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
	start := time.Now()
	for i, trans := range transactions {

		var clonedPool *ppool.Pool
		switch trans.Type {
		case "Mint":
			if !trans.Amount.IsZero() {
				pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
			}
		case "Burn":
			if !trans.Amount.IsZero() {
				pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
			}
		case "Swap":
			amount0, amount1 := new(ui.Int), new(ui.Int)
			if trans.Amount0.Sign() > 0 {
				clonedPool = pool.Clone()
				amount0, amount1 = pool.GetOutputAmount(trans.Amount0, token0, cons.Zero)

				if trans.Amount0.Cmp(amount0) != 0 || trans.Amount1.Cmp(amount1) != 0 || pool.SqrtRatioX96.Cmp(trans.SqrtPriceX96) != 0 || pool.TickCurrent != trans.Tick {
					amount0Cloned, amount1Cloned := clonedPool.GetOutputAmount(trans.Amount0, token0, trans.SqrtPriceX96)
					//fmt.Printf("%d %d %d %d\n", trans.Amount0.SToBig(), amount0.SToBig(), trans.Amount1.SToBig(), amount1.SToBig())
					if trans.Amount0.Cmp(amount0Cloned) != 0 || trans.Amount1.Cmp(amount1Cloned) != 0 || clonedPool.SqrtRatioX96.Cmp(trans.SqrtPriceX96) != 0 || clonedPool.TickCurrent != trans.Tick {
						fmt.Println(trans)
						panic(trans)
					}
					pool = clonedPool
				}
			} else if trans.Amount1.Sign() > 0 {
				clonedPool = pool.Clone()

				amount0, amount1 = pool.GetOutputAmount(trans.Amount1, token1, cons.Zero)
				if trans.Amount0.Cmp(amount0) != 0 || trans.Amount1.Cmp(amount1) != 0 || pool.SqrtRatioX96.Cmp(trans.SqrtPriceX96) != 0 || pool.TickCurrent != trans.Tick {
					amount0Cloned, amount1Cloned := clonedPool.GetOutputAmount(trans.Amount1, token1, trans.SqrtPriceX96)
					if trans.Amount0.Cmp(amount0Cloned) != 0 || trans.Amount1.Cmp(amount1Cloned) != 0 || clonedPool.SqrtRatioX96.Cmp(trans.SqrtPriceX96) != 0 || clonedPool.TickCurrent != trans.Tick {
						//fmt.Printf("%d %d %d \n", clonedPool.SqrtRatioX96, trans.SqrtPriceX96, clonedPool.TickCurrent)
						//
						//fmt.Printf("%d %d %d %d\n", amount1Cloned, trans.Amount1, amount0Cloned, trans.Amount0)
						fmt.Println(trans)
						panic(trans)
					}
					pool = clonedPool
				}
			}
			_ = i
		}
	}
	fmt.Printf("%d\n", pool.Liquidity)
	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)

}

func getTransactions() []ent.Transaction {
	filename := "transactions2.json"
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
