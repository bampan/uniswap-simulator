package strategy

import (
	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	la "github.com/ftchann/uniswap-simulator/lib/liquidity_amounts"
	"github.com/ftchann/uniswap-simulator/lib/pool"
	"github.com/ftchann/uniswap-simulator/lib/prices"
	"github.com/ftchann/uniswap-simulator/lib/tickmath"

	ui "github.com/holiman/uint256"
)

// IntervalAroundAverageStrategy [pa -a, pa + a]
// Where pa is the average of the last n values of the price
// And a is a parameter
type IntervalAroundAverageStrategy struct {
	Amount0       *ui.Int
	Amount1       *ui.Int
	Pool          *pool.Pool
	IntervalWidth int // a in ticks
	Positions     []Position
	PriceHistory  *prices.Prices
}

func NewIntervalAroundAverageStrategy(amount0, amount1 *ui.Int, pool *pool.Pool, intervalWidth, amountAverageSnapshots int) *IntervalAroundAverageStrategy {
	priceHistory := prices.NewPrices(amountAverageSnapshots)
	return &IntervalAroundAverageStrategy{
		Amount0:       amount0.Clone(),
		Amount1:       amount1.Clone(),
		Pool:          pool.Clone(),
		IntervalWidth: intervalWidth,
		Positions:     make([]Position, 0),
		PriceHistory:  priceHistory,
	}
}

func (s *IntervalAroundAverageStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *IntervalAroundAverageStrategy) MakeSnapshot() {
	sqrtPriceX96 := s.Pool.SqrtRatioX96
	s.PriceHistory.Add(sqrtPriceX96)

}

func (s *IntervalAroundAverageStrategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *IntervalAroundAverageStrategy) BurnAll() (retamount0, retamount1 *ui.Int) {
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

func (s *IntervalAroundAverageStrategy) getTicks() (tickLower, tickUpper int) {
	priceX192 := s.PriceHistory.Average()
	sqrtPriceX96 := new(ui.Int).Sqrt(priceX192)
	tick := tickmath.TM.GetTickAtSqrtRatio(sqrtPriceX96)
	tickSpacing := cons.TickSpaces[s.Pool.Fee]
	tickLower = tickmath.Round(tick-s.IntervalWidth, tickSpacing)
	tickUpper = tickmath.Round(tick+s.IntervalWidth, tickSpacing)
	return
}

func (s *IntervalAroundAverageStrategy) mintPosition(tickLower, tickUpper int) {
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

func (s *IntervalAroundAverageStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}

func (s *IntervalAroundAverageStrategy) Rebalance() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.BurnAll()

	tickLower, tickUpper := s.getTicks()
	s.mintPosition(tickLower, tickUpper)
	return
}
