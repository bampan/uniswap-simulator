package pool

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/fullmath"
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

func (p *Pool) Clone() *Pool {
	return &Pool{
		Token0:       p.Token0,
		Token1:       p.Token1,
		Fee:          p.Fee,
		SqrtRatioX96: p.SqrtRatioX96.Clone(),
		Liquidity:    p.Liquidity.Clone(),
		TickSpacing:  p.TickSpacing,
		TickCurrent:  p.TickCurrent,
		//TickData returning the same. Not a full Clone
		TickData: p.TickData,
	}
}

func (p *Pool) Mint(tickLower int, tickUpper int, amount *ui.Int) {
	p.modifyPosition(tickLower, tickUpper, amount)
}

func (p *Pool) Burn(tickLower int, tickUpper int, amount *ui.Int) {
	amountMinus := new(ui.Int)
	amountMinus.Neg(amount)
	p.modifyPosition(tickLower, tickUpper, amountMinus)
}

func (p *Pool) GetOutputAmount(inputAmount *ui.Int, token string, sqrtPriceLimitX96 *ui.Int) (*ui.Int, *ui.Int) {
	zeroForOne := token == p.Token0
	return p.swap(zeroForOne, inputAmount, sqrtPriceLimitX96)
}

func (p *Pool) GetInputAmount(outputAmount *ui.Int, token string, sqrtPriceLimitX96 *ui.Int) (*ui.Int, *ui.Int) {
	zeroForOne := token == p.Token1
	return p.swap(zeroForOne, outputAmount, sqrtPriceLimitX96)
}

func (p *Pool) modifyPosition(lower int, upper int, amount *ui.Int) {
	p.TickData.UpdateTick(lower, amount, false)
	p.TickData.UpdateTick(upper, amount, true)

	if p.TickCurrent >= lower && p.TickCurrent < upper {
		p.Liquidity.Add(p.Liquidity, amount)
	}
}

// flash
// Use amounts instead of Paid
func (p *Pool) flash(amount0 *ui.Int, amount1 *ui.Int) {
	fee0 := fullmath.MulDivRoundingUp(amount0, ui.NewInt(uint64(p.Fee)), ui.NewInt(1e6))
	fee1 := fullmath.MulDivRoundingUp(amount1, ui.NewInt(uint64(p.Fee)), ui.NewInt(1e6))
	_ = fee0
	_ = fee1
}

// swap
// amountSpecified can be negative
func (p *Pool) swap(zeroForOne bool, amountSpecified *ui.Int, sqrtPriceLimitX96In *ui.Int) (*ui.Int, *ui.Int) {
	sqrtPriceLimitX96 := sqrtPriceLimitX96In.Clone()
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
		step.sqrtPriceStartX96 = state.sqrtPriceX96
		step.tickNext, step.initialized = p.TickData.NextInitializedTickWithinOneWord(state.tick, zeroForOne)

		if step.tickNext < tickmath.MinTick {
			step.tickNext = tickmath.MinTick
		} else if step.tickNext > tickmath.MaxTick {
			step.tickNext = tickmath.MaxTick
		}

		step.sqrtPriceNextX96 = tickmath.TM.GetSqrtRatioAtTick(step.tickNext)
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
		//fmt.Printf("SwapStep Out %d %d %d %d \n", state.sqrtPriceX96, step.amountIn, step.amountOut, step.feeAmount)
		//fmt.Printf("SwapStep quotient %d \n", new(ui.Int).Div(step.amountIn, step.amountOut))
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
			state.tick = tickmath.TM.GetTickAtSqrtRatio(state.sqrtPriceX96)
		}

	}
	// Update Slot0
	p.TickCurrent = state.tick
	p.Liquidity = state.liquidity
	p.SqrtRatioX96 = state.sqrtPriceX96
	amount0, amount1 := new(ui.Int), new(ui.Int)
	if zeroForOne == exactInput {
		amount0.Sub(amountSpecified, state.amountSpecifiedRemainingI)
		amount1.Set(state.amountCalculatedI)
	} else {
		amount0.Set(state.amountCalculatedI)
		amount1.Sub(amountSpecified, state.amountSpecifiedRemainingI)
	}
	return amount0, amount1
}
