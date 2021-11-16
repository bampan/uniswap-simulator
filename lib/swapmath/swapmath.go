package swapmath

import (
	fm "uniswap-simulator/lib/fullmath"
	sqrtmath "uniswap-simulator/lib/sqrtprice_math"
	ui "uniswap-simulator/uint256"
)

var MaxFee = new(ui.Int).Exp(ui.NewInt(10), ui.NewInt(6))

func ComputeSwapStep(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemainingI *ui.Int, feePips int) (sqrtRatioNextX96, amountIn, amountOut, feeAmount *ui.Int) {
	//fmt.Printf("SwapMath Input %d %d %d %d \n", sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, amountRemainingI)
	zeroForOne := sqrtRatioCurrentX96.Cmp(sqrtRatioTargetX96) >= 0

	exactIn := amountRemainingI.Sign() >= 0

	if exactIn {
		amountRemainingLessFee := new(ui.Int).Div(new(ui.Int).Mul(amountRemainingI, new(ui.Int).Sub(MaxFee, ui.NewInt(uint64(feePips)))), MaxFee)
		if zeroForOne {
			amountIn = sqrtmath.GetAmount0Delta(sqrtRatioTargetX96, sqrtRatioCurrentX96, liquidity, true)
		} else {
			amountIn = sqrtmath.GetAmount1Delta(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, true)
		}
		if amountRemainingLessFee.Cmp(amountIn) >= 0 {
			sqrtRatioNextX96 = sqrtRatioTargetX96.Clone()
		} else {
			sqrtRatioNextX96 = sqrtmath.GetNextSqrtPriceFromInput(sqrtRatioCurrentX96, liquidity, amountRemainingLessFee, zeroForOne)
		}
	} else {
		if zeroForOne {
			amountOut = sqrtmath.GetAmount1Delta(sqrtRatioTargetX96, sqrtRatioCurrentX96, liquidity, false)
		} else {
			amountOut = sqrtmath.GetAmount0Delta(sqrtRatioCurrentX96, sqrtRatioTargetX96, liquidity, false)
		}
		if new(ui.Int).Neg(amountRemainingI).Cmp(amountOut) >= 0 {
			sqrtRatioNextX96 = sqrtRatioTargetX96.Clone()
		} else {
			sqrtRatioNextX96 = sqrtmath.GetNextSqrtPriceFromOutput(sqrtRatioCurrentX96, liquidity, new(ui.Int).Neg(amountRemainingI), zeroForOne)

		}
	}

	max := sqrtRatioTargetX96.Cmp(sqrtRatioNextX96) == 0

	if zeroForOne {
		if !(max && exactIn) {
			amountIn = sqrtmath.GetAmount0Delta(sqrtRatioNextX96, sqrtRatioCurrentX96, liquidity, true)
		}
		if !(max && !exactIn) {
			amountOut = sqrtmath.GetAmount1Delta(sqrtRatioNextX96, sqrtRatioCurrentX96, liquidity, false)
		}
	} else {
		if !(max && exactIn) {
			amountIn = sqrtmath.GetAmount1Delta(sqrtRatioCurrentX96, sqrtRatioNextX96, liquidity, true)
		}
		if !(max && !exactIn) {
			amountOut = sqrtmath.GetAmount0Delta(sqrtRatioCurrentX96, sqrtRatioNextX96, liquidity, false)
		}
	}

	if !exactIn && amountOut.Cmp(new(ui.Int).Neg(amountRemainingI)) > 0 {
		amountOut = new(ui.Int).Neg(amountRemainingI)
	}

	if exactIn && sqrtRatioNextX96.Cmp(sqrtRatioTargetX96) != 0 {
		// we didn't reach the target, so take the remainder of the maximum input as fee
		feeAmount = new(ui.Int).Sub(amountRemainingI, amountIn)
	} else {
		feeAmount = fm.MulDivRoundingUp(amountIn, ui.NewInt(uint64(feePips)), new(ui.Int).Sub(MaxFee, ui.NewInt(uint64(feePips))))
	}

	return
}
