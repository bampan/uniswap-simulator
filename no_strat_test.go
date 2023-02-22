package main

import (
	"fmt"
	"math/big"
	"testing"

	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	ppool "github.com/ftchann/uniswap-simulator/lib/pool"

	ui "github.com/holiman/uint256"
)

func Test(t *testing.T) {
	transactions := getTransactions()
	token0 := "USDC"
	token1 := "WETH"
	fee := 500
	sqrtX96big, _ := new(big.Int).SetString("1350174849792634181862360983626536", 10)
	sqrtX96, _ := ui.FromBig(sqrtX96big)

	pool := ppool.NewPool(token0, token1, fee, sqrtX96)

	for _, trans := range transactions {
		var amount0, amount1 *ui.Int
		switch trans.Type {
		case "Mint":
			amount0, amount1 = pool.MintStrategy(trans.TickLower, trans.TickUpper, trans.Amount)
			if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) {
				fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
				t.Errorf("Not passing sanity check")
			}

		case "Burn":
			amount0, amount1 = pool.BurnStrategy(trans.TickLower, trans.TickUpper, trans.Amount)

			if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) {
				fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
				t.Errorf("Not passing sanity check")
			}
		case "Swap":

			if trans.Amount0.Sign() > 0 {
				if trans.UseX96 {
					amount0, amount1 = pool.ExactInputSwap(trans.Amount0, pool.Token0, trans.SqrtPriceX96)
				} else {
					amount0, amount1 = pool.ExactInputSwap(trans.Amount0, pool.Token0, cons.Zero)
				}
			} else if trans.Amount1.Sign() > 0 {
				if trans.UseX96 {
					amount0, amount1 = pool.ExactInputSwap(trans.Amount1, pool.Token1, trans.SqrtPriceX96)
				} else {
					amount0, amount1 = pool.ExactInputSwap(trans.Amount1, pool.Token1, cons.Zero)
				}
			}
			if !trans.Amount1.Eq(amount1) || !trans.Amount0.Eq(amount0) || !trans.SqrtPriceX96.Eq(pool.SqrtRatioX96) || trans.Tick != pool.TickCurrent {
				fmt.Printf("%d %d %d %d\n", trans.Amount1, amount1, trans.Amount0, amount0)
				fmt.Printf("%d %d %d %d\n", trans.SqrtPriceX96, pool.SqrtRatioX96, trans.Tick, pool.TickCurrent)
				pool.TickData.Print()
				t.Errorf("Not passing sanity check")
			}

		case "Flash":
			pool.Flash(trans.Amount0, trans.Amount1)
		}
	}
	if pool.SqrtRatioX96.ToBig().String() != "1204434112404346008547779933205831" {
		t.Errorf("Liquidity incorrect")
	}
	if pool.Liquidity.ToBig().String() != "30215871190049079976" {
		t.Errorf("Liquidity incorrect")
	}
	if pool.TickCurrent != 192593 {
		t.Errorf("Tick incorrect")
	}

}
