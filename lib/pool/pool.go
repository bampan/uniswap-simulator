package pool

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/swapmath"
	td "uniswap-simulator/lib/tickdata"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

type StepComputations struct {
	sqrtPriceStartX96 *ui.Int
	tickNext          int
	initialized       bool
	sqrtPriceNextX96  *ui.Int
	amountIn          *ui.Int
	amountOut         *ui.Int
	feeAmount         *ui.Int
}

type stateStruct struct {
	amountSpecifiedRemainingI *ui.Int
	amountCalculatedI         *ui.Int
	sqrtPriceX96              *ui.Int
	tick                      int
	liquidity                 *ui.Int
}

type Pool struct {
	Token0       string
	Token1       string
	Fee          int
	SqrtRatioX96 *ui.Int
	Liquidity    *ui.Int
	TickSpacing  int
	TickCurrent  int
	TickData     *td.TickData
}

func (p *Pool) Mint(tickLower int, tickUpper int, amount *ui.Int) {
	p.modifyPosition(tickLower, tickUpper, amount)
}

func (p *Pool) Burn(tickLower int, tickUpper int, amount *ui.Int) {
	amountMinus := new(ui.Int)
	amountMinus.ChangeSign(amount)
	p.modifyPosition(tickLower, tickUpper, amountMinus)
}

func (p *Pool) GetOutputAmount(inputAmount *ui.Int, token string, sqrtPriceLimitX96 *ui.Int) *ui.Int {
	zeroForOne := token == p.Token0
	return p.swap(zeroForOne, inputAmount, sqrtPriceLimitX96)
}

func (p *Pool) modifyPosition(lower int, upper int, amount *ui.Int) {
	p.TickData.UpdateTick(lower, amount, false)
	p.TickData.UpdateTick(upper, amount, true)

	if p.TickCurrent >= lower && p.TickCurrent < upper {
		p.Liquidity.Add(p.Liquidity, amount)
	}
}

// swap
// amountSpecified can be negative
func (p *Pool) swap(zeroForOne bool, amountSpecified *ui.Int, sqrtPriceLimitX96 *ui.Int) *ui.Int {
	if sqrtPriceLimitX96.IsZero() {
		if zeroForOne {
			sqrtPriceLimitX96.Add(tickmath.MinSqrtRatio, cons.One)
		} else {
			sqrtPriceLimitX96.Sub(tickmath.MaxSqrtRatio, cons.One)
		}
	}

	exactInput := amountSpecified.Sign() >= 0

	state := stateStruct{
		amountSpecified.Clone(),
		ui.NewInt(0),
		p.SqrtRatioX96.Clone(),
		p.TickCurrent,
		p.Liquidity.Clone(),
	}

	//start while loop
	for !state.amountSpecifiedRemainingI.IsZero() && state.sqrtPriceX96.Cmp(sqrtPriceLimitX96) != 0 {
		var step StepComputations
		step.sqrtPriceStartX96 = state.sqrtPriceX96.Clone()
		step.tickNext, step.initialized = p.TickData.NextInitializedTickWithinOneWord(state.tick, zeroForOne)

		if step.tickNext < tickmath.MinTick {
			step.tickNext = tickmath.MinTick
		} else if step.tickNext > tickmath.MaxTick {
			step.tickNext = tickmath.MaxTick
		}

		step.sqrtPriceNextX96 = tickmath.GetSqrtRatioAtTick(step.tickNext)
		var targetValue *ui.Int
		if zeroForOne {
			if step.sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) < 0 {
				targetValue = sqrtPriceLimitX96
			} else {
				targetValue = step.sqrtPriceNextX96
			}
		} else {
			if step.sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) > 0 {
				targetValue = sqrtPriceLimitX96
			} else {
				targetValue = step.sqrtPriceNextX96
			}
		}

		state.sqrtPriceX96, step.amountIn, step.amountOut, step.feeAmount =
			swapmath.ComputeSwapStep(state.sqrtPriceX96,
				targetValue, state.liquidity, state.amountSpecifiedRemainingI, p.Fee)
		//fmt.Printf("%d %d %d %d\n", state.sqrtPriceX96, step.amountIn, step.amountOut, step.feeAmount)
		if exactInput {
			state.amountSpecifiedRemainingI.Sub(state.amountSpecifiedRemainingI, new(ui.Int).Add(step.amountIn, step.feeAmount))
			state.amountCalculatedI.Sub(state.amountCalculatedI, step.amountOut)
		} else {
			state.amountSpecifiedRemainingI = new(ui.Int).Add(state.amountSpecifiedRemainingI, step.amountOut)
			state.amountCalculatedI = new(ui.Int).Add(state.amountCalculatedI, new(ui.Int).Add(step.amountIn, step.feeAmount))
		}

		if state.sqrtPriceX96.Cmp(step.sqrtPriceNextX96) == 0 {
			if step.initialized {
				liquidityNet := p.TickData.GetTick(step.tickNext).LiquidityNet
				if zeroForOne {
					state.liquidity = state.liquidity.Sub(state.liquidity, liquidityNet)
				} else {
					state.liquidity = state.liquidity.Add(state.liquidity, liquidityNet)
				}

			}
			if zeroForOne {
				state.tick = step.tickNext - 1
			} else {
				state.tick = step.tickNext
			}
		} else if state.sqrtPriceX96.Cmp(step.sqrtPriceStartX96) != 0 {
			state.tick = tickmath.GetTickAtSqrtRatio(state.sqrtPriceX96)
		}

	}
	p.TickCurrent = state.tick
	p.Liquidity = state.liquidity
	p.SqrtRatioX96 = state.sqrtPriceX96
	return state.amountCalculatedI
}
