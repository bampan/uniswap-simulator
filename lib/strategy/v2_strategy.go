package strategy

import (
	la "uniswap-simulator/lib/liquidity_amounts"
	"uniswap-simulator/lib/pool"
	"uniswap-simulator/lib/tickmath"
	ui "uniswap-simulator/uint256"
)

// V2Strategy [mintick, maxtick]
type V2Strategy struct {
	Amount0   *ui.Int
	Amount1   *ui.Int
	Pool      *pool.Pool
	Positions []Position
}

func (s *V2Strategy) Rebalance() (*ui.Int, *ui.Int) {
	//TODO implement me
	return nil, nil
}

func (s *V2Strategy) MakeSnapshot() {
	//TODO implement me
}

func NewV2Strategy(amount0, amount1 *ui.Int, pool *pool.Pool) *V2Strategy {
	return &V2Strategy{
		Amount0:   amount0.Clone(),
		Amount1:   amount1.Clone(),
		Pool:      pool.Clone(),
		Positions: make([]Position, 0),
	}
}

func (s *V2Strategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *V2Strategy) GetAmounts() (*ui.Int, *ui.Int) {
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

func (s *V2Strategy) BurnAll() (amount0, amount1 *ui.Int) {
	for _, position := range s.Positions {
		s.Pool.BurnStrategy(position.tickLower, position.tickUpper, position.amount)
		amount0, amount1 := s.Pool.CollectStrategy(position.tickLower, position.tickUpper)
		s.Amount0.Add(s.Amount0, amount0)
		s.Amount1.Add(s.Amount1, amount1)
	}
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	return
}

func (s *V2Strategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	// New Positions
	tickLower := -887270
	tickUpper := -tickLower
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
