package tickdata

import (
	"fmt"
	cons "uniswap-simulator/lib/constants"
	ui "uniswap-simulator/uint256"
)

type Tick struct {
	Index                 int
	LiquidityNet          *ui.Int
	LiquidityGross        *ui.Int
	FeeGrowthOutside0X128 *ui.Int
	FeeGrowthOutside1X128 *ui.Int
}

type TickData struct {
	ticks       []Tick
	tickSpacing int
}

func (t *TickData) Clone() *TickData {
	newTickData := &TickData{
		ticks:       make([]Tick, len(t.ticks)),
		tickSpacing: t.tickSpacing,
	}
	copy(newTickData.ticks, t.ticks)
	return newTickData
}

func (t *TickData) Print() {
	for _, c := range t.ticks {
		fmt.Printf("%d ", c.Index)
	}
	fmt.Println()
}

func NewTickData(tickSpacing int) *TickData {
	return &TickData{
		nil,
		tickSpacing,
	}
}

func (t *TickData) isBelowSmallest(tick int) bool {
	return tick < t.ticks[0].Index
}

func (t *TickData) isAtOrAboveLargest(tick int) bool {
	return tick >= t.ticks[len(t.ticks)-1].Index
}
func (t *TickData) isAboveLargest(tick int) bool {
	return tick > t.ticks[len(t.ticks)-1].Index
}

func (t *TickData) GetTick(index int) Tick {
	tick := t.ticks[t.binarySearch(index)]
	return tick
}

func (t *TickData) Cross(tick int, feeGrowthGlobal0X128, feeGrowthGlobal1X128 *ui.Int) (liquidityNet *ui.Int) {

	info := t.GetTick(tick)
	info.FeeGrowthOutside0X128.Sub(feeGrowthGlobal0X128, info.FeeGrowthOutside0X128)
	info.FeeGrowthOutside1X128.Sub(feeGrowthGlobal1X128, info.FeeGrowthOutside1X128)
	liquidityNet = info.LiquidityNet
	return
}

func makeTick(tick, tickCurrent int, liquidityNet, liquidityGross, feeGrowthOutside0X128, feeGrowthOutside1X128 *ui.Int) Tick {
	if tick <= tickCurrent {
		return Tick{
			Index:                 tick,
			LiquidityNet:          liquidityNet.Clone(),
			LiquidityGross:        liquidityGross.Clone(),
			FeeGrowthOutside0X128: feeGrowthOutside0X128.Clone(),
			FeeGrowthOutside1X128: feeGrowthOutside1X128.Clone(),
		}
	} else {
		return Tick{
			Index:                 tick,
			LiquidityNet:          liquidityNet.Clone(),
			LiquidityGross:        liquidityGross.Clone(),
			FeeGrowthOutside0X128: cons.Zero.Clone(),
			FeeGrowthOutside1X128: cons.Zero.Clone(),
		}
	}

}

func (t *TickData) ClearTick(index int) {
	i, found := t.binarySearch2(index)
	if found {
		t.ticks = append(t.ticks[:i], t.ticks[i+1:]...)
	} else {
		fmt.Println("ClearTick: tick not found")
	}
}

func (t *TickData) UpdateTick(index, tickCurrent int, liquidityDelta, feeGrowthGlobal0X128, feeGrowthGlobal1X128 *ui.Int, upper bool) bool {
	i, found := t.binarySearch2(index)
	var z = new(ui.Int)
	if upper {
		z.Neg(liquidityDelta)
	} else {
		z.Set(liquidityDelta)
	}
	if found {
		tick := t.ticks[i]
		if upper {
			tick.LiquidityNet.Sub(tick.LiquidityNet, liquidityDelta)
		} else {
			tick.LiquidityNet.Add(tick.LiquidityNet, liquidityDelta)
		}
		tick.LiquidityGross.Add(tick.LiquidityGross, liquidityDelta)
		//delete cause its zero
		if tick.LiquidityGross.IsZero() {
			return true
		}
	} else {
		tick := makeTick(index, tickCurrent, z, liquidityDelta, feeGrowthGlobal0X128, feeGrowthGlobal1X128)
		switch i {
		case -2:
			t.ticks = append(t.ticks, tick)
		case -1:
			t.ticks = append([]Tick{tick}, t.ticks...)
		default:
			t.ticks = append(t.ticks[:i+1], t.ticks[i:]...)
			t.ticks[i] = tick
		}
	}
	return false

}

func (t *TickData) GetFeeGrowthInside(tickLower, tickUpper, tickCurrent int, feeGrowthGlobal0X128, feeGrowthGlobal1X128 *ui.Int) (feeGrowthInside0X128, feeGrowthInside1X128 *ui.Int) {
	lower := t.GetTick(tickLower)
	upper := t.GetTick(tickUpper)
	var feeGrowthBelow0X128, feeGrowthBelow1X128 *ui.Int
	if tickCurrent >= tickLower {
		feeGrowthBelow0X128 = new(ui.Int).Set(lower.FeeGrowthOutside0X128)
		feeGrowthBelow1X128 = new(ui.Int).Set(lower.FeeGrowthOutside1X128)
	} else {
		feeGrowthBelow0X128 = new(ui.Int).Sub(feeGrowthGlobal0X128, lower.FeeGrowthOutside0X128)
		feeGrowthBelow1X128 = new(ui.Int).Sub(feeGrowthGlobal1X128, lower.FeeGrowthOutside1X128)
	}
	var feeGrowthAbove0X128, feeGrowthAbove1X128 *ui.Int
	if tickCurrent < tickUpper {
		feeGrowthAbove0X128 = new(ui.Int).Set(upper.FeeGrowthOutside0X128)
		feeGrowthAbove1X128 = new(ui.Int).Set(upper.FeeGrowthOutside1X128)
	} else {
		feeGrowthAbove0X128 = new(ui.Int).Sub(feeGrowthGlobal0X128, upper.FeeGrowthOutside0X128)
		feeGrowthAbove1X128 = new(ui.Int).Sub(feeGrowthGlobal1X128, upper.FeeGrowthOutside1X128)
	}
	feeGrowthInside0X128 = new(ui.Int).Sub(feeGrowthGlobal0X128, feeGrowthBelow0X128)
	feeGrowthInside0X128.Sub(feeGrowthInside0X128, feeGrowthAbove0X128)
	feeGrowthInside1X128 = new(ui.Int).Sub(feeGrowthGlobal1X128, feeGrowthBelow1X128)
	feeGrowthInside1X128.Sub(feeGrowthInside1X128, feeGrowthAbove1X128)
	return
}

func (t *TickData) NextInitializedTickWithinOneWord(tick int, lte bool) (int, bool) {
	compressed := tick / t.tickSpacing
	if lte {
		wordPos := compressed >> 8
		minimum := (wordPos << 8) * t.tickSpacing
		if t.isBelowSmallest(tick) {
			return minimum, false
		}

		index := t.nextInitializedTick(tick, lte).Index
		nextInitializedTick := max(minimum, index)
		return nextInitializedTick, nextInitializedTick == index
	} else {
		wordPos := (compressed + 1) >> 8
		maximum := (((wordPos + 1) << 8) - 1) * t.tickSpacing
		if t.isAtOrAboveLargest(tick) {
			return maximum, false
		}
		index := t.nextInitializedTick(tick, lte).Index
		nextInitializedTick := min(maximum, index)
		return nextInitializedTick, nextInitializedTick == index
	}
}

func (t *TickData) binarySearch(tick int) int {
	l := 0
	r := len(t.ticks) - 1
	var i int
	for {
		i = (l + r) / 2
		if t.ticks[i].Index <= tick && (i == len(t.ticks)-1 || t.ticks[i+1].Index > tick) {
			return i
		}
		if t.ticks[i].Index < tick {
			l = i + 1
		} else {
			r = i - 1
		}
	}
}

func (t *TickData) binarySearch2(tick int) (int, bool) {

	l := 0
	N := len(t.ticks)
	r := N - 1
	// Empty List whatever
	if len(t.ticks) == 0 {
		return -2, false
	}
	if t.isBelowSmallest(tick) {
		return -1, false
	}
	if t.isAboveLargest(tick) {
		return -2, false
	}

	var i int
	for l < r {
		i = l + ((r - l) / 2)
		if tick == t.ticks[i].Index {
			return i, true
		}
		if t.ticks[i].Index < tick {
			l = i + 1
		} else {
			r = i
		}
	}
	if tick == t.ticks[l].Index {
		return l, true
	}

	return l, false

}

func (t *TickData) nextInitializedTick(tick int, lte bool) Tick {
	if lte {
		if t.isAtOrAboveLargest(tick) {
			return t.ticks[len(t.ticks)-1]
		}
		index := t.binarySearch(tick)
		return t.ticks[index]
	} else {
		if t.isBelowSmallest(tick) {
			return t.ticks[0]
		}
		index := t.binarySearch(tick)
		return t.ticks[index+1]
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
