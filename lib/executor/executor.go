package executor

import (
	"math"
	cons "uniswap-simulator/lib/constants"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

type Execution struct {
	Strategy       strat.Strategy
	SqrtPricesX96  []*ui.Int
	Amounts0       []*ui.Int
	Amounts1       []*ui.Int
	Timestamps     []int
	StartTime      int
	UpdateInterval int
	Transactions   []ent.Transaction
}

func CreateExecution(strategy strat.Strategy, starTime, updateInterval int, transactions []ent.Transaction) *Execution {

	maxtime := transactions[len(transactions)-1].Timestamp

	length := (maxtime - starTime) / updateInterval

	prices := make([]*ui.Int, 0, length)
	amounts0 := make([]*ui.Int, 0, length)
	amounts1 := make([]*ui.Int, 0, length)

	return &Execution{
		Strategy:       strategy,
		StartTime:      starTime,
		UpdateInterval: updateInterval,
		Transactions:   transactions,
		SqrtPricesX96:  prices,
		Amounts0:       amounts0,
		Amounts1:       amounts1,
	}

}

func (e *Execution) Run() {
	//
	strategy := e.Strategy
	transactions := e.Transactions
	pool := strategy.GetPool()

	started := false
	nextUpdate := math.MaxInt64

	for _, trans := range transactions {

		// Start Strategy
		if !started && trans.Timestamp > e.StartTime {
			amount0, amount1 := strategy.Init()
			e.Amounts0 = append(e.Amounts0, amount0)
			e.Amounts1 = append(e.Amounts1, amount1)
			e.SqrtPricesX96 = append(e.SqrtPricesX96, pool.SqrtRatioX96.Clone())
			e.Timestamps = append(e.Timestamps, trans.Timestamp)

			nextUpdate = trans.Timestamp + e.UpdateInterval
			started = true
		}

		if trans.Timestamp > nextUpdate {
			strategy.Rebalance()
			//e.Amounts0 = append(e.Amounts0, amount0)
			//e.Amounts1 = append(e.Amounts1, amount1)
			//e.SqrtPricesX96 = append(e.SqrtPricesX96, pool.SqrtRatioX96.Clone())
			//e.Timestamps = append(e.Timestamps, trans.Timestamp)
			nextUpdate = nextUpdate + e.UpdateInterval
		}

		switch trans.Type {
		case "Mint":
			if !trans.Amount.IsZero() {
				pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
			}

		case "Burn":
			if !trans.Amount.IsZero() {
				pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
			}

		case "Swap":

			if trans.Amount0.Sign() > 0 {
				if trans.UseX96 {
					pool.ExactInputSwap(trans.Amount0, pool.Token0, trans.SqrtPriceX96)
				} else {
					pool.ExactInputSwap(trans.Amount0, pool.Token0, cons.Zero)
				}
			} else if trans.Amount1.Sign() > 0 {
				if trans.UseX96 {
					pool.ExactInputSwap(trans.Amount1, pool.Token1, trans.SqrtPriceX96)
				} else {
					pool.ExactInputSwap(trans.Amount1, pool.Token1, cons.Zero)
				}
			}
		case "Flash":
			pool.Flash(trans.Amount0, trans.Amount1)
		}
	}
	finalTimestamp := transactions[len(transactions)-1].Timestamp
	amount0, amount1 := strategy.BurnAll()
	e.Amounts0 = append(e.Amounts0, amount0)
	e.Amounts1 = append(e.Amounts1, amount1)
	finalPrice := pool.SqrtRatioX96.Clone()
	e.SqrtPricesX96 = append(e.SqrtPricesX96, finalPrice)
	e.Timestamps = append(e.Timestamps, finalTimestamp)

}
