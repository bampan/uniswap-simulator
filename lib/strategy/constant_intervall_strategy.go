package strategy

import (
	cons "uniswap-simulator/lib/constants"
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// ConstantIntervalStrategy [p-a, p+a]
type ConstantIntervalStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	tickLower     int
	tickUpper     int
	Positions     []Position
}

func NewConstantIntervallStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *ConstantIntervalStrategy {
	return &ConstantIntervalStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
	}
}

func (s *ConstantIntervalStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *ConstantIntervalStrategy) GetAmounts() (*ui.Int, *ui.Int) {
	amount0, amount1 := new(ui.Int), new(ui.Int)
	for _, position := range s.Positions {
		// Collect Fees
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, cons.Zero)
		fee0, fee1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, fee0)
		s.Amount1.Add(s.Amount1, fee1)

		sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(position.tickLower)
		sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(position.tickUpper)
		liquidityAmount0, liquidityAmount1 := la.GetAmountsForLiquidity(position.amount, s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96)
		amount0.Add(amount0, liquidityAmount0)
		amount1.Add(amount0, liquidityAmount1)

	}
	amount0.Add(amount0, s.Amount0)
	amount1.Add(amount1, s.Amount1)
	return amount0, amount1
}

func (s *ConstantIntervalStrategy) BurnAll() (amount0, amount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	return
}

func (s *ConstantIntervalStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	// New Positions
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)
	s.tickLower, s.tickUpper = tickLower, tickUpper
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)
	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
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

func (s *ConstantIntervalStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {

	// Only runs ones
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)
	// New Positions
	tickLower := s.tickLower
	tickUpper := s.tickUpper
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)
	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
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
