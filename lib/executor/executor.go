package executor

import (
	"math"
	cons "uniswap-simulator/lib/constants"
	sqrtmath "uniswap-simulator/lib/sqrtprice_math"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

type Execution struct {
	Strategy           strat.Strategy
	StartTime          int
	UpdateInterval     int
	SnapShotInterval   int
	AmountUSDSnapshots []*ui.Int
	Transactions       []ent.Transaction
}

func CreateExecution(strategy strat.Strategy, starTime, updateInterval, SnapShotInterval int, transactions []ent.Transaction) *Execution {

	maxTime := transactions[len(transactions)-1].Timestamp
	length := (maxTime - starTime) / updateInterval
	snapshots := make([]*ui.Int, 0, length)

	return &Execution{
		Strategy:           strategy,
		StartTime:          starTime,
		UpdateInterval:     updateInterval,
		SnapShotInterval:   SnapShotInterval,
		Transactions:       transactions,
		AmountUSDSnapshots: snapshots,
	}

}

func (e *Execution) Run() {
	//
	strategy := e.Strategy
	transactions := e.Transactions
	pool := strategy.GetPool()

	started := false
	nextUpdate := math.MaxInt64
	nextSnapshot := math.MaxInt64

	for _, trans := range transactions {

		// Start Strategy
		if !started && trans.Timestamp >= e.StartTime {
			amount0, amount1 := strategy.Init()

			x96 := strategy.GetPool().SqrtRatioX96
			price := sqrtmath.GetPrice(x96)
			startAmount1to0 := new(ui.Int).Div(amount1, price)
			amountUSD := new(ui.Int).Add(startAmount1to0, amount0)
			e.AmountUSDSnapshots = append(e.AmountUSDSnapshots, amountUSD)

			nextUpdate = trans.Timestamp + e.UpdateInterval
			nextSnapshot = trans.Timestamp + e.SnapShotInterval
			started = true
		}

		// Snapshot
		if trans.Timestamp >= nextSnapshot {
			amount0, amount1 := strategy.GetAmounts()
			x96 := strategy.GetPool().SqrtRatioX96
			price := sqrtmath.GetPrice(x96)
			amount1to0 := new(ui.Int).Div(amount1, price)
			amountUSD := new(ui.Int).Add(amount1to0, amount0)
			e.AmountUSDSnapshots = append(e.AmountUSDSnapshots, amountUSD)
			nextSnapshot += e.SnapShotInterval
		}

		// Rebalance
		if trans.Timestamp >= nextUpdate {
			strategy.Rebalance()
			nextUpdate += e.UpdateInterval
		}
		switch trans.Type {
		case "Mint":
			if !trans.Amount.IsZero() {
				pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
				// add a line
			}

		case "Burn":
			if !trans.Amount.IsZero() {
				pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
			}

		case "Swap":

			if trans.Amount0.Sign() > 0 {
				pool.ExactInputSwap(trans.Amount0, pool.Token0, cons.Zero)
			} else if trans.Amount1.Sign() > 0 {
				pool.ExactInputSwap(trans.Amount1, pool.Token1, cons.Zero)
			}
		case "Flash":
			pool.Flash(trans.Amount0, trans.Amount1)
		}
	}
	amount0, amount1 := strategy.BurnAll()
	x96 := strategy.GetPool().SqrtRatioX96
	price := sqrtmath.GetPrice(x96)
	startAmount1to0 := new(ui.Int).Div(amount1, price)
	amountUSD := new(ui.Int).Add(startAmount1to0, amount0)
	e.AmountUSDSnapshots = append(e.AmountUSDSnapshots, amountUSD)

}
