package strategy

import (
	cons "uniswap-simulator/lib/constants"
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// IntervalAroundPriceStrategy [pc-a, pc+a]
// Where pc is the current price

type IntervalAroundPriceAndSwapStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	Positions     []Position
}

func (s *IntervalAroundPriceAndSwapStrategy) MakeSnapshot() {
	//TODO implement me
}

func (s *IntervalAroundPriceAndSwapStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *IntervalAroundPriceAndSwapStrategy) GetAmounts() (*ui.Int, *ui.Int) {
	amount0, amount1 := new(ui.Int), new(ui.Int)
	for _, position := range s.Positions {
		sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(position.tickLower)
		sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(position.tickUpper)
		liquidityAmount0, liquidityAmount1 := la.GetAmountsForLiquidity(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, position.amount)
		amount0.Add(amount0, liquidityAmount0)
		amount1.Add(amount1, liquidityAmount1)
	}
	amount0.Add(amount0, s.Amount0)
	amount1.Add(amount1, s.Amount1)
	return amount0, amount1
}

func NewIntervalAroundPriceAndSwapStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *IntervalAroundPriceAndSwapStrategy {
	return &IntervalAroundPriceAndSwapStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
	}
}

func (s *IntervalAroundPriceAndSwapStrategy) BurnAll() (retamount0, retamount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	retamount0, retamount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)
	return
}

func (s *IntervalAroundPriceAndSwapStrategy) mintPosition(tickLower, tickUpper int) {
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)

	amountNoSwap := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	amount0NoSwap, amount1NoSwap := la.GetAmountsForLiquidity(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, amountNoSwap)

	amount0Diff := new(ui.Int).Sub(s.Amount0, amount0NoSwap)

	// 10^9 so that it is about equal to amount1
	amount0Cmp := new(ui.Int).Lsh(amount0Diff, 9)
	amount1Diff := new(ui.Int).Sub(s.Amount1, amount1NoSwap)

	if amount0Cmp.Cmp(amount1Diff) > 0 {
		// amount0 is more than amount1
		amount0DiffHalf := new(ui.Int).Div(amount0Diff, ui.NewInt(2))
		if amount0DiffHalf.Sign() == 1 {
			amount0Swap, amount1Swap := s.Pool.ExactInputSwap(amount0DiffHalf, s.Pool.Token0, cons.Zero)
			s.Amount0.Sub(s.Amount0, amount0Swap)
			s.Amount1.Sub(s.Amount1, amount1Swap)
		}
	} else {
		// amount1 is more than amount0
		amount1DiffHalf := new(ui.Int).Div(amount1Diff, ui.NewInt(2))
		if amount1DiffHalf.Sign() == 1 {
			amount0Swap, amount1Swap := s.Pool.ExactInputSwap(amount1DiffHalf, s.Pool.Token1, cons.Zero)
			s.Amount0.Sub(s.Amount0, amount0Swap)
			s.Amount1.Sub(s.Amount1, amount1Swap)
		}
	}
	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	s.Positions = append(s.Positions, Position{
		amount:    amount,
		tickLower: tickLower,
		tickUpper: tickUpper,
	})

	amount0, amount1 := s.Pool.MintStrategy(tickLower, tickUpper, amount)
	s.Amount0.Sub(s.Amount0, amount0)
	s.Amount1.Sub(s.Amount1, amount1)

}

func (s *IntervalAroundPriceAndSwapStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	// New Positions
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)
	s.mintPosition(tickLower, tickUpper)
	return
}

func (s *IntervalAroundPriceAndSwapStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {

	// We are not interested in GasFee So just burn every time.
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}

	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)

	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)

	s.mintPosition(tickLower, tickUpper)
	return
}
