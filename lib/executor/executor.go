package executor

import (
	"math"
	cons "uniswap-simulator/lib/constants"
	"uniswap-simulator/lib/prices"
	sqrtmath "uniswap-simulator/lib/sqrtprice_math"
	strat "uniswap-simulator/lib/strategy"
	ent "uniswap-simulator/lib/transaction"
	ui "uniswap-simulator/uint256"
)

type Execution struct {
	Strategy               strat.Strategy
	StartTime              int
	UpdateInterval         int
	SnapShotInterval       int
	MovingAverageWindow    int
	AmountAverageSnapshots int
	AmountUSDSnapshots     []*ui.Int
	Transactions           []ent.Transaction
	PricesSnapshots        *prices.Prices
}

func CreateExecution(strategy strat.Strategy, starTime, updateInterval, snapShotInterval, movingAverageWindow, amountAverageSnapshots int, transactions []ent.Transaction) *Execution {

	maxTime := transactions[len(transactions)-1].Timestamp
	length := (maxTime - starTime) / updateInterval
	snapshots := make([]*ui.Int, 0, length)
	pricesSnapshots := prices.NewPrices(amountAverageSnapshots)

	return &Execution{
		Strategy:               strategy,
		StartTime:              starTime,
		UpdateInterval:         updateInterval,
		SnapShotInterval:       snapShotInterval,
		MovingAverageWindow:    movingAverageWindow,
		AmountAverageSnapshots: amountAverageSnapshots,
		Transactions:           transactions,
		AmountUSDSnapshots:     snapshots,
		PricesSnapshots:        pricesSnapshots,
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
	nextPriceSnapshot := math.MaxInt64

	priceSnapshotInterval := e.MovingAverageWindow / e.AmountAverageSnapshots

	for _, trans := range transactions {

		// Start Strategy
		if !started && trans.Timestamp >= e.StartTime {
			amount0, amount1 := strategy.Init()
			// Not Precise
			x96 := strategy.GetPool().SqrtRatioX96
			price := sqrtmath.GetPrice(x96)
			startAmount1to0 := new(ui.Int).Div(amount1, price)
			amountUSD := new(ui.Int).Add(startAmount1to0, amount0)
			e.AmountUSDSnapshots = append(e.AmountUSDSnapshots, amountUSD)

			nextUpdate = trans.Timestamp + e.UpdateInterval
			nextSnapshot = trans.Timestamp + e.SnapShotInterval
			started = true
		}

		// Price Snapshot
		if trans.Timestamp > nextPriceSnapshot {
			e.PricesSnapshots.Add(e.Strategy.GetPool().SqrtRatioX96)
			nextPriceSnapshot += priceSnapshotInterval
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
			//priceSquareX192 := e.PricesSnapshots.Average()
			//price := new(ui.Int).Sqrt(priceSquareX192)
			//sqrtPriceX96 := new(ui.Int).Lsh(price, 96)
			strategy.Rebalance()
			nextUpdate += e.UpdateInterval
		}
		switch trans.Type {
		case "Mint":
			pool.Mint(trans.TickLower, trans.TickUpper, trans.Amount)
		case "Burn":
			pool.Burn(trans.TickLower, trans.TickUpper, trans.Amount)
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
