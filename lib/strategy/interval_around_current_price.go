package strategy

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// IntervalAroundPriceStrategy [pc-a, pc+a]
// Where pc is the current price

type IntervalAroundPriceStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	Positions     []Position
}

func (s *IntervalAroundPriceStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func NewIntervalAroundPriceStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *IntervalAroundPriceStrategy {
	return &IntervalAroundPriceStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
	}
}

func (s *IntervalAroundPriceStrategy) BurnAll() (amount0, amount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	return
}

func (s *IntervalAroundPriceStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	// New Positions
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)

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
	return
}

func (s *IntervalAroundPriceStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {

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

	// New Positions
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
	return
}
