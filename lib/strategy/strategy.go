package strategy

import (
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

type Position struct {
	amount    *ui.Int
	tickLower int
	tickUpper int
}

// Strategy [p-a, p+a]
type Strategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	Positions     []Position
}

func NewStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *Strategy {
	return &Strategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
	}
}

func (s *Strategy) BurnAll() {
	for _, position := range s.Positions {
		amount0, amount1 := s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
}

func (s *Strategy) Rebalance() {

	// We are not interested in GasFee So just burn every time.
	for _, position := range s.Positions {
		amount0, amount1 := s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	s.Positions = make([]Position, 0)

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
	//fmt.Printf("%d %d %d \n", tickLower, s.Pool.TickCurrent, tickUpper)
	amount0, amount1 := s.Pool.MintStrategy(tickLower, tickUpper, amount)
	s.Amount0.Sub(s.Amount0, amount0)
	s.Amount1.Sub(s.Amount1, amount1)

}
