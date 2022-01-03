package strategy

import (
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/prices"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// BollingerBandsStrategy [pa - c*o, pa + c* o]
// Where pa is the current price
// c is a constant
// o is the volatility
type BollingerBandsStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	Positions     []Position
	MultiplierX10 *ui.Int // Q.6.10
	PriceHistory  *prices.Prices
}

func NewBollingerBandsStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, amountAverageSnapshots, multiplier int) *BollingerBandsStrategy {
	priceHistory := prices.NewPrices(amountAverageSnapshots)
	multiplierX10 := ui.NewInt(uint64(multiplier))
	return &BollingerBandsStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		Positions:     make([]Position, 0),
		MultiplierX10: multiplierX10,
		PriceHistory:  priceHistory,
	}
}

func (s *BollingerBandsStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *BollingerBandsStrategy) MakeSnapshot() {
	sqrtPriceX96 := s.Pool.SqrtRatioX96
	s.PriceHistory.Add(sqrtPriceX96)

}

func (s *BollingerBandsStrategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *BollingerBandsStrategy) BurnAll() (retamount0, retamount1 *ui.Int) {
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

func (s *BollingerBandsStrategy) getTicks() (tickLower, tickUpper int) {
	volatilityX192 := s.PriceHistory.Volatility()
	volatilityScaledX200 := new(ui.Int).Mul(volatilityX192, s.MultiplierX10)
	volatilityScaledX192 := new(ui.Int).Rsh(volatilityScaledX200, 10)
	sqrtVolatilityX96 := new(ui.Int).Sqrt(volatilityScaledX192)

	priceX192 := s.PriceHistory.Average()
	sqrtPriceX96 := new(ui.Int).Sqrt(priceX192)

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

func (s *BollingerBandsStrategy) mintPosition(tickLower, tickUpper int) {
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

func (s *BollingerBandsStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}

func (s *BollingerBandsStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.BurnAll()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}
