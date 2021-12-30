package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"strconv"
	"testing"
	cons "uniswap-simulator/lib/constants"
	ppool "uniswap-simulator/lib/pool"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

func Test_Fee(t *testing.T) {
	transactions := getTransactionsTest()
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96big, _ := new(big.Int).SetString("1350174849792634181862360983626536", 10)
	sqrtX96, _ := ui.FromBig(sqrtX96big)

	pool := ppool.NewPool(token0, token1, fee, sqrtX96)

	amountBig, _ := new(big.Int).SetString("93924580278", 10)
	amount, _ := ui.FromBig(amountBig)
	pool.MintStrategy(190880, 198880, amount)
	for _, trans := range transactions {
		switch trans.Type {
		case "Mint":
			pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
		case "Burn":
			pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
		case "Swap":
			if trans.Amount0.Sign() > 0 {
				if trans.UseX96 {
					pool.ExactInputSwap(trans.Amount0, pool.Token0, trans.SqrtPriceX96)
				} else {
					pool.ExactInputSwap(trans.Amount0, pool.Token0, cons.Zero)
				}
			} else if trans.Amount1.Sign() > 0 {
				if trans.UseX96 {
					pool.ExactInputSwap(trans.Amount1, pool.Token1, trans.SqrtPriceX96)
				} else {
					pool.ExactInputSwap(trans.Amount1, pool.Token1, cons.Zero)
				}
			}
		case "Flash":
			pool.Flash(trans.Amount0, trans.Amount1)
		}
	}
	pool.BurnStrategy(190880, 198880, amount)
	amount0, amount1 := pool.CollectStrategy(190880, 198880)
	amount0Str := fmt.Sprintf("%d", amount0)
	amount1Str := fmt.Sprintf("%d", amount1)
	liquidityStr := fmt.Sprintf("%d", pool.Liquidity)
	sqrtPriceX96 := fmt.Sprintf("%d", pool.SqrtRatioX96)

	if amount0Str != "1053517" {
		t.Errorf("amount0Str is not correct, got: %s, want: %s.", amount0Str, "1053517")
	}
	if amount1Str != "275576756661612" {
		t.Errorf("amount1Str is not correct, got: %s, want: %s.", amount1Str, "939245")
	}
	if liquidityStr != "734329717995335932" {
		t.Errorf("liquidityStr is not correct, got: %s, want: %s.", liquidityStr, "939245")
	}
	if sqrtPriceX96 != "1337536101591430553762461821094985" {
		t.Errorf("sqrtPriceX96 is not correct, got: %s, want: %s.", sqrtPriceX96, "1350174849792634181862360983626536")
	}

}

func getTransactionsTest() []ent.Transaction {
	filename := "trans.json"
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
