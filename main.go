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
	ppool "uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickdata"
	"uniswap-simulator/lib/tickmath"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

func main() {
	fmt.Println("Start")
	transactions := getTransactions()

	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96, _ := ui.FromHex("0x42919A3B4E1F2E279AB5FE196328")
	tickSpacing := constants.TickSpaces[fee]
	liquidity := ui.NewInt(0)
	tickCurrent := tickmath.GetTickAtSqrtRatio(sqrtX96)
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
	for _, trans := range transactions {
		//fmt.Printf("%+v\n", pool.TickData)
		//fmt.Printf("%d\n", pool.Liquidity)
		switch trans.Type {
		case "mint":
			pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
		case "burn":
			pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
		case "swap":
			var outputamount *ui.Int
			if trans.Amount0.Sign() > 0 {
				outputamount = pool.GetOutputAmount(trans.Amount0, token0, ui.NewInt(0))
				//fmt.Printf("%d\n", outputamount.SToBig())
				//fmt.Printf("%d\n", trans.Amount1.SToBig())
			} else if trans.Amount1.Sign() > 0 {
				outputamount = pool.GetOutputAmount(trans.Amount1, token1, ui.NewInt(0))
				//fmt.Printf("%d\n", outputamount.SToBig())
				//fmt.Printf("%d\n", trans.Amount0.SToBig())
			}
			_ = outputamount
		}
		//if i+1 >= 30 {
		//	break
		//}
	}
	fmt.Printf("%d\n", pool.Liquidity)
	elapsed := time.Since(start)
	log.Printf("took %s", elapsed)

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
