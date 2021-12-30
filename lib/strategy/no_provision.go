package strategy

import (
	"uniswap-simulator/lib/pool"
	ui "uniswap-simulator/uint256"
)

// V2Strategy [mintick, maxtick]
type NoProvisionStrategy struct {
	Amount0 *ui.Int
	Amount1 *ui.Int
	Pool    *pool.Pool
}

func (s *NoProvisionStrategy) Rebalance() (*ui.Int, *ui.Int) {
	//TODO implement me
	return nil, nil
}

func (s *NoProvisionStrategy) MakeSnapshot() {
	//TODO implement me
}

func NewNoProvisionStrategy(amount0, amount1 *ui.Int, pool *pool.Pool) *NoProvisionStrategy {
	return &NoProvisionStrategy{
		Amount0: amount0.Clone(),
		Amount1: amount1.Clone(),
		Pool:    pool.Clone(),
	}
}

func (s *NoProvisionStrategy) GetPool() *pool.Pool {
	return s.Pool
}

func (s *NoProvisionStrategy) GetAmounts() (*ui.Int, *ui.Int) {
	amount0, amount1 := s.Amount0.Clone(), s.Amount1.Clone()
	return amount0, amount1
}

func (s *NoProvisionStrategy) BurnAll() (amount0, amount1 *ui.Int) {
	amount0, amount1 = s.Amount0.Clone(), s.Amount1.Clone()
	return
}

func (s *NoProvisionStrategy) Init() (currAmount0, currAmount1 *ui.Int) {
	currAmount0, currAmount1 = s.Amount0.Clone(), s.Amount1.Clone()

	return
}
