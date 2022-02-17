package strategy

import (
	cons "uniswap-simulator/lib/constants"
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

type LimitOrderStrategy struct {
	Amount0          *ui.Int
	Amount1          *ui.Int
	Pool             *pool.Pool
	IntervalWidth    int // a in ticks
	Positions        []Position
	CurrentLimitTick int
	Direction        bool
}

func (s *LimitOrderStrategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *LimitOrderStrategy) GetDirections() bool {
	return s.Direction
}

func NewLimitOrderStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth int) *LimitOrderStrategy {
	return &LimitOrderStrategy{
		Amount0:       amount0,
		Amount1:       amount1,
		Pool:          pool,
		IntervalWidth: intervalWidth,
		Positions:     []Position{},
	}
}

func (s *LimitOrderStrategy) MakeSnapshot() {
	//TODO implement me
}

func (s *LimitOrderStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *LimitOrderStrategy) GetCurrentLimitTick() int {
	return s.CurrentLimitTick
}

func (s *LimitOrderStrategy) GetDirection() bool {
	return s.Direction
}

func (s *LimitOrderStrategy) mintPosition(tickLower, tickUpper int) {
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)

	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	if amount.IsZero() {
		return
	}

	s.mintAmount(tickLower, tickUpper, amount)

}

func (s *LimitOrderStrategy) mintAmount(tickLower, tickUpper int, amount *ui.Int) {
	s.Positions = append(s.Positions, Position{
		amount:    amount,
		tickLower: tickLower,
		tickUpper: tickUpper,
	})

	amount0, amount1 := s.Pool.MintStrategy(tickLower, tickUpper, amount)
	s.Amount0.Sub(s.Amount0, amount0)
	s.Amount1.Sub(s.Amount1, amount1)
}

func (s *LimitOrderStrategy) mintHalf(tickLower, tickUpper int) {
	sqrtRatioAX96 := tickmath.TM.GetSqrtRatioAtTick(tickLower)
	sqrtRatioBX96 := tickmath.TM.GetSqrtRatioAtTick(tickUpper)

	amount := la.GetLiquidityForAmount(s.Pool.SqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, s.Amount0, s.Amount1)
	amountHalf := new(ui.Int).Div(amount, ui.NewInt(2))
	if amountHalf.IsZero() {
		return
	}
	s.mintAmount(tickLower, tickUpper, amountHalf)
}

func (s *LimitOrderStrategy) BurnAll() (amount0, amount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.Positions = make([]Position, 0)
	return
}

func (s *LimitOrderStrategy) setPositions() {
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower := tickmath.Round(s.Pool.TickCurrent-s.IntervalWidth, tickSpacing)
	tickUpper := tickmath.Round(s.Pool.TickCurrent+s.IntervalWidth, tickSpacing)
	s.mintPosition(tickLower, tickUpper)

	// Secondary position
	if !s.Amount0.IsZero() {
		tickLower = tickmath.Ceil(s.Pool.TickCurrent, tickSpacing)
		tickUpper = tickLower + 10
		s.mintHalf(tickLower, tickUpper)
		s.CurrentLimitTick = tickUpper
		s.Direction = true
	}

	if !s.Amount1.IsZero() {
		tickUpper = tickmath.Floor(s.Pool.TickCurrent, tickSpacing)
		tickLower = tickUpper - 10
		s.mintHalf(tickLower, tickUpper)
		s.CurrentLimitTick = tickLower
		s.Direction = false
	}
}

func (s *LimitOrderStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()
	s.setPositions()
	return
}

func (s *LimitOrderStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.BurnAll()
	s.setPositions()
	return
}
