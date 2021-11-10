package tickdata

import (
	ui "uniswap-simulator/uint256"
)

type Tick struct {
	index        int
	LiquidityNet *ui.Int
}

type TickData struct {
	ticks       []Tick
	tickSpacing int
}

func NewTickData(tickSpacing int) *TickData {
	return &TickData{
		nil,
		tickSpacing,
	}
}

func (t *TickData) isBelowSmallest(tick int) bool {
	return tick < t.ticks[0].index
}

func (t *TickData) isAtOrAboveLargest(tick int) bool {
	return tick >= t.ticks[len(t.ticks)-1].index
}

func (t *TickData) GetTick(index int) Tick {
	tick := t.ticks[t.binarySearch(index)]
	return tick
}

func (t *TickData) UpdateTick(index int, liquidityDelta *ui.Int, upper bool) {
	i := t.binarySearch2(index)
	var tick Tick
	if upper {
		var z = new(ui.Int)
		z.Neg(liquidityDelta)
		tick = Tick{index, z}
	} else {
		tick = Tick{index, liquidityDelta}
	}
	switch i {
	case -2:
		t.ticks = append(t.ticks, tick)
	case -1:
		t.ticks = append([]Tick{tick}, t.ticks...)
	default:
		if i < len(t.ticks) && t.ticks[i].index == index {
			tick = t.ticks[i]
			if upper {
				tick.LiquidityNet.Sub(tick.LiquidityNet, liquidityDelta)
			} else {
				tick.LiquidityNet.Add(tick.LiquidityNet, liquidityDelta)
			}
		} else {
			t.ticks = append(t.ticks[:i+1], t.ticks[i:]...)
			t.ticks[i] = tick
		}
	}

}

func (t *TickData) NextInitializedTickWithinOneWord(tick int, lte bool) (int, bool) {
	compressed := tick / t.tickSpacing
	if lte {
		wordPos := compressed >> 8
		minimum := (wordPos << 8) * t.tickSpacing
		if t.isBelowSmallest(tick) {
			return minimum, false
		}

		index := t.nextInitializedTick(tick, lte).index
		nextInitializedTick := max(minimum, index)
		return nextInitializedTick, nextInitializedTick == index
	} else {
		wordPos := (compressed + 1) >> 8
		maximum := ((wordPos+1)<<8)*t.tickSpacing - 1
		if t.isAtOrAboveLargest(tick) {
			return maximum, false
		}
		index := t.nextInitializedTick(tick, lte).index
		nextInitializedTick := min(maximum, index)
		return nextInitializedTick, nextInitializedTick == index
	}
}

func (t *TickData) binarySearch(tick int) int {
	l := 0
	r := len(t.ticks) - 1
	var i int
	for {
		i = l + ((r - l) / 2)
		if t.ticks[i].index <= tick && (i == len(t.ticks)-1 || t.ticks[i+1].index > tick) {
			return i
		}
		if t.ticks[i].index < tick {
			l = i + 1
		} else {
			r = i - 1
		}
	}
}

func (t *TickData) binarySearch2(tick int) int {
	l := 0
	N := len(t.ticks)
	r := N - 1

	var i int
	for l < r {
		i = l + ((r - l) / 2)
		if i < N && tick == t.ticks[i].index {
			return i
		}
		if t.ticks[i].index < tick {
			l = i + 1
		} else {
			r = i
		}
	}
	if l == 0 && N == 0 {
		return -2
	}
	if l < N && l >= 0 && t.ticks[l].index < tick {
		l = -2
	}
	if l == 0 && N > 0 && t.ticks[0].index > tick {
		l = -1
	}
	return l
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
