package sqrtprice_math

import (
	cons "uniswap-simulator/lib/constants"
	fm "uniswap-simulator/lib/fullmath"
	ui "uniswap-simulator/uint256"
)

var MaxUint160 = new(ui.Int).Sub(new(ui.Int).Exp(ui.NewInt(2), ui.NewInt(160)), cons.One)

func multiplyIn256(x, y *ui.Int) *ui.Int {
	product := new(ui.Int).Mul(x, y)
	return new(ui.Int).And(product, cons.MaxUint256)
}

func addIn256(x, y *ui.Int) *ui.Int {
	sum := new(ui.Int).Add(x, y)
	return new(ui.Int).And(sum, cons.MaxUint256)
}

func GetAmount0Delta(sqrtRatioAX96, sqrtRatioBX96, liquidity *ui.Int, roundUp bool) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) >= 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	numerator1 := new(ui.Int).Lsh(liquidity, 96)
	numerator2 := new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96)

	if roundUp {
		return fm.MulDivRoundingUp(fm.MulDivRoundingUp(numerator1, numerator2, sqrtRatioBX96), cons.One, sqrtRatioAX96)
	}
	res := new(ui.Int)
	res.Mul(numerator1, numerator2)
	res.Div(res, sqrtRatioBX96)
	res.Div(res, sqrtRatioAX96)
	return res
}

func GetAmount1Delta(sqrtRatioAX96, sqrtRatioBX96, liquidity *ui.Int, roundUp bool) *ui.Int {
	if sqrtRatioAX96.Cmp(sqrtRatioBX96) >= 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	if roundUp {
		return fm.MulDivRoundingUp(liquidity, new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96), cons.Q96)
	}
	return new(ui.Int).Div(new(ui.Int).Mul(liquidity, new(ui.Int).Sub(sqrtRatioBX96, sqrtRatioAX96)), cons.Q96)
}

func GetNextSqrtPriceFromInput(sqrtPX96, liquidity, amountIn *ui.Int, zeroForOne bool) *ui.Int {
	if zeroForOne {
		return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountIn, true)
	}
	return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountIn, true)
}

func GetNextSqrtPriceFromOutput(sqrtPX96, liquidity, amountOut *ui.Int, zeroForOne bool) *ui.Int {
	if zeroForOne {
		return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountOut, false)
	}
	return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountOut, false)
}

func getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amount *ui.Int, add bool) *ui.Int {
	if amount.IsZero() {
		return sqrtPX96
	}

	numerator1 := new(ui.Int).Lsh(liquidity, 96)
	if add {
		product := multiplyIn256(amount, sqrtPX96)
		if new(ui.Int).Div(product, amount).Eq(sqrtPX96) {
			denominator := addIn256(numerator1, product)
			if denominator.Cmp(numerator1) >= 0 {
				ans := fm.MulDivRoundingUp(numerator1, sqrtPX96, denominator)
				return ans
			}
		}
		return fm.MulDivRoundingUp(numerator1, cons.One, new(ui.Int).Add(new(ui.Int).Div(numerator1, sqrtPX96), amount))
	} else {
		product := multiplyIn256(amount, sqrtPX96)
		denominator := new(ui.Int).Sub(numerator1, product)
		return fm.MulDivRoundingUp(numerator1, sqrtPX96, denominator)
	}
}

func getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amount *ui.Int, add bool) *ui.Int {
	if add {
		var quotient *ui.Int
		if amount.Cmp(MaxUint160) <= 0 {
			quotient = new(ui.Int).Div(new(ui.Int).Lsh(amount, 96), liquidity)
		} else {
			quotient = new(ui.Int).Div(new(ui.Int).Mul(amount, cons.Q96), liquidity)
		}
		return new(ui.Int).Add(sqrtPX96, quotient)
	}

	quotient := fm.MulDivRoundingUp(amount, cons.Q96, liquidity)

	return new(ui.Int).Sub(sqrtPX96, quotient)
}
