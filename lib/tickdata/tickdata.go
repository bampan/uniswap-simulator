package tickdata

import (
	"fmt"
	ui "uniswap-simulator/uint256"
)

type Tick struct {
	Index          int
	LiquidityNet   *ui.Int
	LiquidityGross *ui.Int
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

func (t *TickData) GetTick(index int) Tick {
	tick := t.ticks[t.binarySearch(index)]
	return tick
}

func (t *TickData) GetStrategyTick(index int) (Tick, bool) {
	if len(t.ticks) == 0 {
		return Tick{}, false
	}
	l := 0
	r := len(t.ticks) - 1
	i := index
	for l <= r {
		i = (l + r) / 2
		if t.ticks[i].Index == index {
			return t.ticks[i], true
		}
		if t.ticks[i].Index < index {
			l = i + 1
		} else {
			r = i - 1
		}
	}
	return Tick{}, false

}

func (t *TickData) UpdateTick(index int, liquidityDelta *ui.Int, upper bool) {
	i := t.binarySearch2(index)
	var tick Tick
	var z = new(ui.Int)
	if upper {
		z.Neg(liquidityDelta)
		tick = Tick{index, z, new(ui.Int).Set(liquidityDelta)}
	} else {
		z.Set(liquidityDelta)
		tick = Tick{index, z, new(ui.Int).Set(liquidityDelta)}
	}
	switch i {
	case -2:
		t.ticks = append(t.ticks, tick)
	case -1:
		t.ticks = append([]Tick{tick}, t.ticks...)
	default:
		if i < len(t.ticks) && t.ticks[i].Index == index {
			tick = t.ticks[i]
			if upper {
				tick.LiquidityNet.Sub(tick.LiquidityNet, liquidityDelta)
			} else {
				tick.LiquidityNet.Add(tick.LiquidityNet, liquidityDelta)
			}
			tick.LiquidityGross.Add(tick.LiquidityGross, liquidityDelta)
			if tick.LiquidityGross.IsZero() {
				t.ticks = append(t.ticks[:i], t.ticks[i+1:]...)
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

func (t *TickData) binarySearch2(tick int) int {
	l := 0
	N := len(t.ticks)
	r := N - 1

	var i int
	for l < r {
		i = l + ((r - l) / 2)
		if i < N && tick == t.ticks[i].Index {
			return i
		}
		if t.ticks[i].Index < tick {
			l = i + 1
		} else {
			r = i
		}
	}
	if l == 0 && N == 0 {
		return -2
	}
	if l < N && l >= 0 && t.ticks[l].Index < tick {
		l = -2
	}
	if l == 0 && N > 0 && t.ticks[0].Index > tick {
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
