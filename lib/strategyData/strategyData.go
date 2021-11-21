package strategyData

import (
	td "uniswap-simulator/lib/tickdata"
	ui "uniswap-simulator/uint256"
)

type StrategyData struct {
	FeeAmount0 *ui.Int
	FeeAmount1 *ui.Int
	TickData   *td.TickData
	Liquidity  *ui.Int
}

// Clone
func (sd *StrategyData) Clone() *StrategyData {
	return &StrategyData{
		FeeAmount0: sd.FeeAmount0.Clone(),
		FeeAmount1: sd.FeeAmount1.Clone(),
		TickData:   sd.TickData.Clone(),
		Liquidity:  sd.Liquidity.Clone(),
	}
}
