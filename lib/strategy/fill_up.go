package strategy

import (
	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	la "github.com/ftchann/uniswap-simulator/lib/liquidity_amounts"
	"github.com/ftchann/uniswap-simulator/lib/pool"
	"github.com/ftchann/uniswap-simulator/lib/tickmath"

	ui "github.com/holiman/uint256"
)

type FillUpStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	Positions     []Position
}

func (s *FillUpStrategy) MakeSnapshot() {

}
func (s *FillUpStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func NewFillUpStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *FillUpStrategy {
	return &FillUpStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
	}
}

func (s *FillUpStrategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *FillUpStrategy) BurnAll() (retAmount0, retAmount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	retAmount0, retAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)
	return
}

func (s *FillUpStrategy) mintPosition(tickLower, tickUpper int) {
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)

	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	if amount.IsZero() {
		return
	}
	s.Positions = append(s.Positions, Position{
		amount:    amount,
		tickLower: tickLower,
		tickUpper: tickUpper,
	})

	amount0, amount1 := s.Pool.MintStrategy(tickLower, tickUpper, amount)
	s.Amount0.Sub(s.Amount0, amount0)
	s.Amount1.Sub(s.Amount1, amount1)
}

func (s *FillUpStrategy) setPositions() {
	// New Positions
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)

	s.mintPosition(tickLower, tickUpper)
	// SecondaryPosition
	if !s.Amount0.IsZero() {
		tickLower = tickmath.Ceil(s.Pool.TickCurrent, tickSpacing)
		if tickLower < tickUpper {
			s.mintPosition(tickLower, tickUpper)
		}
	} else if !s.Amount1.IsZero() {
		tickUpper = tickmath.Floor(s.Pool.TickCurrent, tickSpacing)
		if tickLower < tickUpper {
			s.mintPosition(tickLower, tickUpper)
		}
	}

}

func (s *FillUpStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	s.setPositions()
	return
}

func (s *FillUpStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.BurnAll()
	s.setPositions()
	return
}
