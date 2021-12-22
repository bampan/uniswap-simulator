package liquidity_amounts

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/fullmath"
	ui "uniswap-simulator/uint256"
)

func getLiquidityForAmount0(sqrtRatioAX96, sqrtRatioBX96, amount0 *ui.Int) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	intermediate := fullmath.MulDiv(sqrtRatioAX96, sqrtRatioBX96, cons.Q96)
	return fullmath.MulDiv(amount0, intermediate, new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96))
}

func getLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1 *ui.Int) *ui.Int {
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
		liquidity = getLiquidityForAmount0(sqrtRatioAX96, sqrtRatioBX96, amount0)
	} else if sqrtRatioX96.Cmp(sqrtRatioBX96) < 0 {
		liquidity0 := getLiquidityForAmount0(sqrtRatioX96, sqrtRatioBX96, amount0)
		liquidity1 := getLiquidityForAmount1(sqrtRatioAX96, sqrtRatioX96, amount1)

		if liquidity0.Cmp(liquidity1) < 0 {
			liquidity = liquidity0
		} else {
			liquidity = liquidity1
		}
	} else {
		liquidity = getLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1)
	}
	return liquidity
}

func getAmount0ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity *ui.Int) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	temp1 := new(ui.Int).Lsh(liquidity, 96)
	temp2 := new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96)
	temp3 := fullmath.MulDiv(temp1, temp2, sqrtRatioBX96)
	return new(ui.Int).Div(temp3, sqrtRatioAX96)
}

func getAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity *ui.Int) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	temp1 := new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96)
	return fullmath.MulDiv(liquidity, temp1, cons.Q96)
}

func GetAmountsForLiquidity(sqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, liquidity *ui.Int) (*ui.Int, *ui.Int) {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) > 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}
	amount0, amount1 := new(ui.Int), new(ui.Int)
	if sqrtRatioX96.Cmp(sqrtRatioAX96) <= 0 {
		amount0 = getAmount0ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity)
	} else if sqrtRatioX96.Cmp(sqrtRatioBX96) < 0 {
		amount0 = getAmount0ForLiquidity(sqrtRatioX96, sqrtRatioBX96, liquidity)
		amount1 = getAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioX96, liquidity)
	} else {
		amount1 = getAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity)
	}
	return amount0, amount1
}
