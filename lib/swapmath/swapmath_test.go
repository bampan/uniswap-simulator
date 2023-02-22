package swapmath

import (
	"fmt"
	"math/big"
	"testing"

	ui "github.com/holiman/uint256"
)

func Test(t *testing.T) {

	current_big, _ := new(big.Int).SetString("1344919684864506912172695223877090", 10)
	current, _ := ui.FromBig(current_big)
	target_big, _ := new(big.Int).SetString("1346938477169594858818217023321238", 10)
	target, _ := ui.FromBig(target_big)
	liquidity_big, _ := new(big.Int).SetString("731344820973715931", 10)
	liquidity, _ := ui.FromBig(liquidity_big)
	amountRemaining_big, _ := new(big.Int).SetString("26412237337162431364", 10)
	amountRemaining, _ := ui.FromBig(amountRemaining_big)
	sqrtPriceX96, amountIn, amountOut, feeAmount := ComputeSwapStep(current, target, liquidity, amountRemaining, 500)
	fmt.Printf("%d %d %d %d \n", sqrtPriceX96, amountIn, amountOut, feeAmount)
}
