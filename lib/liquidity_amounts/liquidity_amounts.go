package liquidity_amounts

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/fullmath"
	ui "uniswap-simulator/uint256"
)

func GetLiquidityForAmount0(sqrtRatioAX96, sqrtRatioBX96, amount0 *ui.Int) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	intermediate := fullmath.MulDiv(sqrtRatioAX96, sqrtRatioBX96, cons.Q96)
	return fullmath.MulDiv(amount0, intermediate, new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96))
}

func GetLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1 *ui.Int) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	return fullmath.MulDiv(amount1, cons.Q96, new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96))
}

func GetLiquidityForAmount(sqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, amount0, amount1 *ui.Int) (liquidity *ui.Int) {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	if sqrtRatioX96.Cmp(sqrtRatioAX96) <= 0 {
		liquidity = GetLiquidityForAmount0(sqrtRatioAX96, sqrtRatioBX96, amount0)
	} else if sqrtRatioX96.Cmp(sqrtRatioBX96) < 0 {
		liquidity0 := GetLiquidityForAmount0(sqrtRatioX96, sqrtRatioBX96, amount0)
		liquidity1 := GetLiquidityForAmount1(sqrtRatioAX96, sqrtRatioX96, amount1)

		if liquidity0.Cmp(liquidity1) < 0 {
			liquidity = liquidity0
		} else {
			liquidity = liquidity1
		}
	} else {
		liquidity = GetLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1)
	}
	return liquidity
}
