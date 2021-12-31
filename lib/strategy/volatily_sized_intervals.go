package strategy

import (
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/prices"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// VolatilitySizedIntervalStrategy [pc - c*o, pc + c* o]
// Where pc is the current price
// c is a constant
// o is the volatility
type VolatilitySizedIntervalStrategy struct {
	Amount0      *ui.Int
	Amount1      *ui.Int
	Pool         *pool.Pool
	Positions    []Position
	MultiplierX8 *ui.Int // Q8.8
	PriceHistory *prices.Prices
}

func NewVolatilitySizedIntervalStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, amountAverageSnapshots, multiplier int) *VolatilitySizedIntervalStrategy {
	priceHistory := prices.NewPrices(amountAverageSnapshots)
	multiplierX8 := ui.NewInt(uint64(multiplier))
	return &VolatilitySizedIntervalStrategy{
		Amount0:      amount0.Clone(),
		Amount1:      amount1.Clone(),
		Pool:         pool.Clone(),
		Positions:    make([]Position, 0),
		MultiplierX8: multiplierX8,
		PriceHistory: priceHistory,
	}
}

func (s *VolatilitySizedIntervalStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *VolatilitySizedIntervalStrategy) MakeSnapshot() {
	sqrtPriceX96 := s.Pool.SqrtRatioX96
	s.PriceHistory.Add(sqrtPriceX96)

}

func (s *VolatilitySizedIntervalStrategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *VolatilitySizedIntervalStrategy) BurnAll() (retamount0, retamount1 *ui.Int) {
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

func (s *VolatilitySizedIntervalStrategy) getTicks() (tickLower, tickUpper int) {
	volatilityX192 := s.PriceHistory.Volatility()
	sqrtPriceX96 := s.Pool.SqrtRatioX96
	volatilityScaledX200 := new(ui.Int).Mul(volatilityX192, s.MultiplierX8)
	volatilityScaledX192 := new(ui.Int).Rsh(volatilityScaledX200, 8)
	//priceX192 := new(ui.Int).Mul(sqrtPriceX96, sqrtPriceX96)
	sqrtVolatilityX96 := new(ui.Int).Sqrt(volatilityScaledX192)

	sqrtRatioAX96, overflow0 := new(ui.Int).SubOverflow(sqrtPriceX96, sqrtVolatilityX96)
	sqrtRatioBX96, overflow1 := new(ui.Int).AddOverflow(sqrtPriceX96, sqrtVolatilityX96)

	if overflow0 || sqrtRatioAX96.Cmp(tickmath.MinSqrtRatio) == -1 {
		tickLower = tickmath.MinTick
	} else {
		tickLower = tickmath.TM.GetTickAtSqrtRatio(sqrtRatioAX96)
	}

	if overflow1 || sqrtRatioBX96.Cmp(tickmath.MaxSqrtRatio) == 1 {
		tickUpper = tickmath.MaxTick
	} else {
		tickUpper = tickmath.TM.GetTickAtSqrtRatio(sqrtRatioBX96)
	}

	return
}

func (s *VolatilitySizedIntervalStrategy) mintPosition(tickLower, tickUpper int) {
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

func (s *VolatilitySizedIntervalStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}

func (s *VolatilitySizedIntervalStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.BurnAll()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}
