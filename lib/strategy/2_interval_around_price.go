package strategy

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

/* 2 Interval Around Price
One Symmetric Interval around the price. [p-a,p+a]
And the remaining liquidity in a limit order [p, p + b] or [p - b, p]
*/

type TwoIntervalAroundPriceStrategy struct {
	Amount0   *ui.Int
	Amount1   *ui.Int
	Pool      *pool.Pool
	a         int // a in ticks
	b         int // b in ticks
	Positions []Position
}

func NewTwoIntervalAroundPriceStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, a, b int) *TwoIntervalAroundPriceStrategy {
	return &TwoIntervalAroundPriceStrategy{
		Amount0:   amount0.Clone(),
		Amount1:   amount1.Clone(),
		Pool:      pool.Clone(),
		a:         a,
		b:         b,
		Positions: make([]Position, 0),
	}
}

func (s *TwoIntervalAroundPriceStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *TwoIntervalAroundPriceStrategy) BurnAll() (amount0, amount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	return
}

func (s *TwoIntervalAroundPriceStrategy) mintPosition(tickLower, tickUpper int) {
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)

	amount := liquidity_amounts.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	s.Positions = append(s.Positions, Position{
		amount:    amount,
		tickLower: tickLower,
		tickUpper: tickUpper,
	})

	amount0, amount1 := s.Pool.MintStrategy(tickLower, tickUpper, amount)
	s.Amount0.Sub(s.Amount0, amount0)
	s.Amount1.Sub(s.Amount1, amount1)
}

func (s *TwoIntervalAroundPriceStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	// MainPosition
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.a, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.a, tickSpacing)

	s.mintPosition(tickLower, tickUpper)

	// SecondaryPosition
	if !s.Amount0.IsZero() {
		tickLower = tickmath.Ceil(s.Pool.TickCurrent, tickSpacing)
		tickUpper = tickLower + s.b
		s.mintPosition(tickLower, tickUpper)
	}

	if !s.Amount1.IsZero() {
		tickUpper = tickmath.Floor(s.Pool.TickCurrent, tickSpacing)
		tickLower = tickUpper - s.b
		s.mintPosition(tickLower, tickUpper)
	}

	return
}

func (s *TwoIntervalAroundPriceStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {

	// We are not interested in GasFee So just burn every time.
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}

	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)

	// MainPosition
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.a, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.a, tickSpacing)
	s.mintPosition(tickLower, tickUpper)

	// SecondaryPosition
	if !s.Amount0.IsZero() {
		tickLower = tickmath.Ceil(s.Pool.TickCurrent, tickSpacing)
		tickUpper = tickLower + s.b
		s.mintPosition(tickLower, tickUpper)
	}

	if !s.Amount1.IsZero() {
		tickUpper = tickmath.Floor(s.Pool.TickCurrent, tickSpacing)
		tickLower = tickUpper - s.b
		s.mintPosition(tickLower, tickUpper)
	}

	return
}