package pool

import (
	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	"github.com/ftchann/uniswap-simulator/lib/fullmath"
	"github.com/ftchann/uniswap-simulator/lib/invariant"
	"github.com/ftchann/uniswap-simulator/lib/position"
	"github.com/ftchann/uniswap-simulator/lib/sqrtprice_math"
	"github.com/ftchann/uniswap-simulator/lib/swapmath"
	td "github.com/ftchann/uniswap-simulator/lib/tickdata"
	"github.com/ftchann/uniswap-simulator/lib/tickmath"

	ui "github.com/holiman/uint256"
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
	feeGrowthGlobalX128       *ui.Int
	liquidity                 *ui.Int
}

type Pool struct {
	Token0               string
	Token1               string
	Fee                  int
	SqrtRatioX96         *ui.Int
	Liquidity            *ui.Int
	FeeGrowthGlobal0X128 *ui.Int
	FeeGrowthGlobal1X128 *ui.Int
	TickSpacing          int
	TickCurrent          int
	TickData             *td.TickData
	Positions            map[string]*position.Info
}

func NewPool(token0, token1 string, fee int, sqrtRatioX96 *ui.Int) *Pool {
	invariant.Invariant(fee >= 0 && fee <= 1000000, "fee must be between 0 and 1000000")
	tickSpacing := cons.TickSpaces[fee]
	liquidity := ui.NewInt(0)
	tickCurrent := tickmath.TM.GetTickAtSqrtRatio(sqrtRatioX96)
	tickData := td.NewTickData(tickSpacing)

	positions := make(map[string]*position.Info)
	pool := &Pool{
		token0,
		token1,
		fee,
		sqrtRatioX96,
		liquidity,
		cons.Zero.Clone(),
		cons.Zero.Clone(),
		tickSpacing,
		tickCurrent,
		tickData,
		positions,
	}
	return pool
}

func (p *Pool) Clone() *Pool {
	positions := make(map[string]*position.Info)
	for k, v := range p.Positions {
		positions[k] = v.Clone()
	}
	return &Pool{
		Token0:               p.Token0,
		Token1:               p.Token1,
		Fee:                  p.Fee,
		SqrtRatioX96:         p.SqrtRatioX96.Clone(),
		Liquidity:            p.Liquidity.Clone(),
		FeeGrowthGlobal0X128: p.FeeGrowthGlobal0X128.Clone(),
		FeeGrowthGlobal1X128: p.FeeGrowthGlobal1X128.Clone(),
		TickSpacing:          p.TickSpacing,
		TickCurrent:          p.TickCurrent,
		TickData:             p.TickData.Clone(),
		Positions:            positions,
	}
}
func (p *Pool) updatePosition(tickLower, tickUpper int, amount *ui.Int) (pos *position.Info) {
	var clearLower, clearUpper bool

	if !amount.IsZero() {
		clearLower = p.TickData.UpdateTick(tickLower, p.TickCurrent, amount, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, false)
		clearUpper = p.TickData.UpdateTick(tickUpper, p.TickCurrent, amount, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, true)
	}

	feeGrowthInside0X128, feeGrowthInside1X128 := p.TickData.GetFeeGrowthInside(tickLower, tickUpper, p.TickCurrent, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128)
	searchString := string(tickLower) + "-" + string(tickUpper)
	pos = p.Positions[searchString]
	if pos == nil {
		pos = position.NewPosition()
		pos.Update(amount, feeGrowthInside0X128, feeGrowthInside1X128)
		p.Positions[searchString] = pos
	} else {
		pos.Update(amount, feeGrowthInside0X128, feeGrowthInside1X128)
		p.Positions[searchString] = pos
	}
	if amount.Sign() == -1 {
		if clearLower {
			p.TickData.ClearTick(tickLower)
		}
		if clearUpper {
			p.TickData.ClearTick(tickUpper)
		}
	}
	return
}

func (p *Pool) checkTicks(tickLower, tickUpper int) {
	invariant.Invariant(tickLower < tickUpper, "tickLower must be less than tickUpper")
	invariant.Invariant(tickLower >= tickmath.MinTick, "tickLower must be greater than or equal to MinTick")
	invariant.Invariant(tickUpper <= tickmath.MaxTick, "tickUpper must be less than or equal to MaxTick")
}

func (p *Pool) modifyPositionStrategy(tickLower int, tickUpper int, amount *ui.Int) (pos *position.Info, amount0, amount1 *ui.Int) {
	p.checkTicks(tickLower, tickUpper)
	pos = p.updatePosition(tickLower, tickUpper, amount)

	if amount.IsZero() {
		return pos, cons.Zero.Clone(), cons.Zero.Clone()
	}
	if p.TickCurrent < tickLower {
		amount0 = sqrtprice_math.GetAmount0DeltaRounded(tickmath.TM.GetSqrtRatioAtTick(tickLower), tickmath.TM.GetSqrtRatioAtTick(tickUpper), amount)
		amount1 = ui.NewInt(0)
	} else if p.TickCurrent < tickUpper {
		amount0 = sqrtprice_math.GetAmount0DeltaRounded(p.SqrtRatioX96, tickmath.TM.GetSqrtRatioAtTick(tickUpper), amount)
		amount1 = sqrtprice_math.GetAmount1DeltaRounded(p.SqrtRatioX96, tickmath.TM.GetSqrtRatioAtTick(tickLower), amount)
		p.Liquidity.Add(p.Liquidity, amount)
	} else {
		amount0 = ui.NewInt(0)
		amount1 = sqrtprice_math.GetAmount1DeltaRounded(tickmath.TM.GetSqrtRatioAtTick(tickLower), tickmath.TM.GetSqrtRatioAtTick(tickUpper), amount)
	}

	return
}

func (p *Pool) MintStrategy(tickLower int, tickUpper int, amount *ui.Int) (amount0, amount1 *ui.Int) {
	invariant.Invariant(amount.Sgt(cons.Zero), "amount must be greater than zero")
	_, amount0, amount1 = p.modifyPositionStrategy(tickLower, tickUpper, amount)
	return
}

// BurnStrategy doesn't actually pay out. It just updates the position.
func (p *Pool) BurnStrategy(tickLower int, tickUpper int, amount *ui.Int) (*ui.Int, *ui.Int) {

	amountMinus := new(ui.Int)
	amountMinus.Neg(amount)
	pos, amount0Int, amount1Int := p.modifyPositionStrategy(tickLower, tickUpper, amountMinus)
	amount0, amount1 := new(ui.Int).Neg(amount0Int), new(ui.Int).Neg(amount1Int)
	pos.TokensOwed0.Add(pos.TokensOwed0, amount0)
	pos.TokensOwed1.Add(pos.TokensOwed1, amount1)
	// return is kinda useless
	return amount0, amount1
}

// CollectStrategy Always Collect all
func (p *Pool) CollectStrategy(tickLower int, tickUpper int) (amount0, amount1 *ui.Int) {
	searchString := string(tickLower) + "-" + string(tickUpper)
	pos := p.Positions[searchString]

	amount0 = pos.TokensOwed0.Clone()
	amount1 = pos.TokensOwed1.Clone()

	pos.TokensOwed0.Set(cons.Zero)
	pos.TokensOwed1.Set(cons.Zero)

	return

}

func (p *Pool) Mint(tickLower int, tickUpper int, amount *ui.Int) {
	invariant.Invariant(amount.Sgt(cons.Zero), "amount must be greater than zero")
	p.modifyPosition(tickLower, tickUpper, amount)
}

func (p *Pool) Burn(tickLower int, tickUpper int, amount *ui.Int) {
	amountMinus := new(ui.Int)
	amountMinus.Neg(amount)
	p.modifyPosition(tickLower, tickUpper, amountMinus)
}

func (p *Pool) ExactInputSwap(inputAmount *ui.Int, token string, sqrtPriceLimitX96 *ui.Int) (*ui.Int, *ui.Int) {
	zeroForOne := token == p.Token0
	return p.swap(zeroForOne, inputAmount, sqrtPriceLimitX96)
}

func (p *Pool) ExactOutputSwap(outputAmount *ui.Int, token string, sqrtPriceLimitX96 *ui.Int) (*ui.Int, *ui.Int) {
	zeroForOne := token == p.Token1
	return p.swap(zeroForOne, outputAmount, sqrtPriceLimitX96)
}

func (p *Pool) modifyPosition(lower int, upper int, amount *ui.Int) {
	p.checkTicks(lower, upper)
	clearLower := p.TickData.UpdateTick(lower, p.TickCurrent, amount, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, false)
	clearUpper := p.TickData.UpdateTick(upper, p.TickCurrent, amount, p.FeeGrowthGlobal0X128, p.FeeGrowthGlobal1X128, true)

	if clearLower {
		p.TickData.ClearTick(lower)
	}
	if clearUpper {
		p.TickData.ClearTick(upper)
	}

	if p.TickCurrent >= lower && p.TickCurrent < upper {
		p.Liquidity.Add(p.Liquidity, amount)
	}
}

// Flash
// Use amounts instead of Paid
func (p *Pool) Flash(amount0 *ui.Int, amount1 *ui.Int) {
	invariant.Invariant(p.Liquidity.Sign() == 1, "Liquidity must be greater than zero")
	fee0 := fullmath.MulDivRoundingUp(amount0, ui.NewInt(uint64(p.Fee)), ui.NewInt(1_000_000))
	fee1 := fullmath.MulDivRoundingUp(amount1, ui.NewInt(uint64(p.Fee)), ui.NewInt(1_000_000))

	fee0Q128 := fullmath.MulDiv(fee0, cons.Q128, p.Liquidity)
	fee1Q128 := fullmath.MulDiv(fee1, cons.Q128, p.Liquidity)

	p.FeeGrowthGlobal0X128.Add(p.FeeGrowthGlobal0X128, fee0Q128)
	p.FeeGrowthGlobal1X128.Add(p.FeeGrowthGlobal1X128, fee1Q128)

}

// swap
// amountSpecified can be negative
func (p *Pool) swap(zeroForOne bool, amountSpecified *ui.Int, sqrtPriceLimitX96In *ui.Int) (*ui.Int, *ui.Int) {

	invariant.Invariant(!amountSpecified.IsZero(), "amount must be not zero")

	sqrtPriceLimitX96 := sqrtPriceLimitX96In.Clone()
	// Allow Zero as sqrtPriceLimitX96
	if sqrtPriceLimitX96.IsZero() {
		if zeroForOne {
			sqrtPriceLimitX96.Add(tickmath.MinSqrtRatio, cons.One)
		} else {
			sqrtPriceLimitX96.Sub(tickmath.MaxSqrtRatio, cons.One)
		}
	}

	var cond bool
	if zeroForOne {
		cond = sqrtPriceLimitX96.Cmp(p.SqrtRatioX96) == -1 && sqrtPriceLimitX96.Cmp(tickmath.MinSqrtRatio) == 1
	} else {
		cond = sqrtPriceLimitX96.Cmp(p.SqrtRatioX96) == 1 && sqrtPriceLimitX96.Cmp(tickmath.MaxSqrtRatio) == -1
	}
	invariant.Invariant(cond, "sqrtPriceLimitX96 must be between min and max")

	exactInput := amountSpecified.Sign() >= 0
	var feeGrowthGlobalX128 *ui.Int
	if zeroForOne {
		feeGrowthGlobalX128 = p.FeeGrowthGlobal0X128
	} else {
		feeGrowthGlobalX128 = p.FeeGrowthGlobal1X128
	}
	state := stateStruct{
		amountSpecified.Clone(),
		ui.NewInt(0),
		p.SqrtRatioX96.Clone(),
		p.TickCurrent,
		feeGrowthGlobalX128.Clone(),
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

		if exactInput {
			state.amountSpecifiedRemainingI.Sub(state.amountSpecifiedRemainingI, new(ui.Int).Add(step.amountIn, step.feeAmount))
			state.amountCalculatedI.Sub(state.amountCalculatedI, step.amountOut)

		} else { // exactOutput
			state.amountSpecifiedRemainingI = new(ui.Int).Add(state.amountSpecifiedRemainingI, step.amountOut)
			state.amountCalculatedI = new(ui.Int).Add(state.amountCalculatedI, new(ui.Int).Add(step.amountIn, step.feeAmount))

		}

		if state.liquidity.Sign() > 0 {
			fee := fullmath.MulDiv(step.feeAmount, cons.Q128, state.liquidity)
			state.feeGrowthGlobalX128.Add(state.feeGrowthGlobalX128, fee)
		}

		if state.sqrtPriceX96.Cmp(step.sqrtPriceNextX96) == 0 {
			if step.initialized {
				var feeGrowthGlobal0X128, feeGrowthGlobal1X128 *ui.Int
				if zeroForOne {
					feeGrowthGlobal0X128 = state.feeGrowthGlobalX128
					feeGrowthGlobal1X128 = p.FeeGrowthGlobal1X128
				} else {
					feeGrowthGlobal0X128 = p.FeeGrowthGlobal0X128
					feeGrowthGlobal1X128 = state.feeGrowthGlobalX128
				}
				liquidityNet := p.TickData.Cross(step.tickNext, feeGrowthGlobal0X128, feeGrowthGlobal1X128)

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

	if zeroForOne {
		p.FeeGrowthGlobal0X128 = state.feeGrowthGlobalX128
	} else {
		p.FeeGrowthGlobal1X128 = state.feeGrowthGlobalX128
	}

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
