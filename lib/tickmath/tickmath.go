package tickmath

import (
	"math"
	"math/big"

	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	"github.com/ftchann/uniswap-simulator/lib/invariant"

	ui "github.com/holiman/uint256"
)

const (
	MinTick    int = -887272  // The minimum tick that can be used on any pool.
	MaxTick    int = -MinTick // The maximum tick that can be used on any pool.
	TotalTicks int = MaxTick - MinTick + 1
)

var (
	Q32             = ui.NewInt(1 << 32)
	MinSqrtRatio    = ui.NewInt(4295128739) // The sqrt ratio corresponding to the minimum tick that could be used on any pool.
	maxbigratio, _  = new(big.Int).SetString("1461446703485210103287273052203988822378723970342", 10)
	MaxSqrtRatio, _ = ui.FromBig(maxbigratio) // The sqrt ratio corresponding to the maximum tick that could be used on any pool.
)

type TickMath struct {
	ticks [TotalTicks]*ui.Int
}

var TM = initTickMath()

func initTickMath() *TickMath {
	t := new(TickMath)

	for i := 0; i < TotalTicks; i++ {
		t.ticks[i] = getSqrtRatioAtTick(i + MinTick)
	}
	return t
}

func Round(ix, iunit int) int {
	x := float64(ix)
	unit := float64(iunit)
	return int(math.Round(x/unit) * unit)
}

func Ceil(ix, iunit int) int {
	x := float64(ix)
	unit := float64(iunit)
	return int(math.Ceil(x/unit) * unit)
}

func Floor(ix, iunit int) int {
	x := float64(ix)
	unit := float64(iunit)
	return int(math.Floor(x/unit) * unit)
}

func (t *TickMath) GetSqrtRatioAtTick(tick int) *ui.Int {
	return new(ui.Int).Set(t.ticks[tick+MaxTick])
}
func (t *TickMath) GetTickAtSqrtRatio(sqrtRatioX96 *ui.Int) int {
	invariant.Invariant(sqrtRatioX96.Cmp(MinSqrtRatio) >= 0 && sqrtRatioX96.Cmp(MaxSqrtRatio) < 0, "sqrtRatioX96 must be between MinSqrtRatio and MaxSqrtRatio")
	l := 0
	r := TotalTicks - 1
	var mid int
	for l < r {
		// Ticks never overflow, so we can use the mid as an index.
		mid = (l + r + 1) / 2
		if t := t.ticks[mid]; t.Cmp(sqrtRatioX96) > 0 {
			r = mid - 1
		} else {
			l = mid
		}
	}
	return l + MinTick
}

// GetSqrtRatioAtTick
// Returns the sqrt ratio as a Q64.96 for the given tick. The sqrt ratio is computed as sqrt(1.0001)^tick
// @param tick the tick for which to compute the sqrt ratio
func getSqrtRatioAtTick(tick int) *ui.Int {
	absTick := tick
	if tick < 0 {
		absTick = -tick
	}
	invariant.Invariant(absTick <= MaxTick, "tick out of range")
	var ratio *ui.Int
	if absTick&0x1 != 0 {
		ratio, _ = ui.FromHex("0xfffcb933bd6fad37aa2d162d1a594001")
	} else {
		ratio, _ = ui.FromHex("0x100000000000000000000000000000000")
	}
	if (absTick & 0x2) != 0 {
		ratio = mulShift(ratio, "0xfff97272373d413259a46990580e213a")
	}
	if (absTick & 0x4) != 0 {
		ratio = mulShift(ratio, "0xfff2e50f5f656932ef12357cf3c7fdcc")
	}
	if (absTick & 0x8) != 0 {
		ratio = mulShift(ratio, "0xffe5caca7e10e4e61c3624eaa0941cd0")
	}
	if (absTick & 0x10) != 0 {
		ratio = mulShift(ratio, "0xffcb9843d60f6159c9db58835c926644")
	}
	if (absTick & 0x20) != 0 {
		ratio = mulShift(ratio, "0xff973b41fa98c081472e6896dfb254c0")
	}
	if (absTick & 0x40) != 0 {
		ratio = mulShift(ratio, "0xff2ea16466c96a3843ec78b326b52861")
	}
	if (absTick & 0x80) != 0 {
		ratio = mulShift(ratio, "0xfe5dee046a99a2a811c461f1969c3053")
	}
	if (absTick & 0x100) != 0 {
		ratio = mulShift(ratio, "0xfcbe86c7900a88aedcffc83b479aa3a4")
	}
	if (absTick & 0x200) != 0 {
		ratio = mulShift(ratio, "0xf987a7253ac413176f2b074cf7815e54")
	}
	if (absTick & 0x400) != 0 {
		ratio = mulShift(ratio, "0xf3392b0822b70005940c7a398e4b70f3")
	}
	if (absTick & 0x800) != 0 {
		ratio = mulShift(ratio, "0xe7159475a2c29b7443b29c7fa6e889d9")
	}
	if (absTick & 0x1000) != 0 {
		ratio = mulShift(ratio, "0xd097f3bdfd2022b8845ad8f792aa5825")
	}
	if (absTick & 0x2000) != 0 {
		ratio = mulShift(ratio, "0xa9f746462d870fdf8a65dc1f90e061e5")
	}
	if (absTick & 0x4000) != 0 {
		ratio = mulShift(ratio, "0x70d869a156d2a1b890bb3df62baf32f7")
	}
	if (absTick & 0x8000) != 0 {
		ratio = mulShift(ratio, "0x31be135f97d08fd981231505542fcfa6")
	}
	if (absTick & 0x10000) != 0 {
		ratio = mulShift(ratio, "0x9aa508b5b7a84e1c677de54f3e99bc9")
	}
	if (absTick & 0x20000) != 0 {
		ratio = mulShift(ratio, "0x5d6af8dedb81196699c329225ee604")
	}
	if (absTick & 0x40000) != 0 {
		ratio = mulShift(ratio, "0x2216e584f5fa1ea926041bedfe98")
	}
	if (absTick & 0x80000) != 0 {
		ratio = mulShift(ratio, "0x48a170391f7dc42444e8fa2")
	}
	if tick > 0 {
		ratio = new(ui.Int).Div(cons.MaxUint256, ratio)
	}

	// back to Q96
	if new(ui.Int).SMod(ratio, Q32).Sign() > 0 {
		return new(ui.Int).Add(new(ui.Int).Div(ratio, Q32), cons.One)
	} else {
		return new(ui.Int).Div(ratio, Q32)
	}
}

// GetTickAtSqrtRatio /**
func getTickAtSqrtRatio(sqrtRatioX96 *ui.Int) int {
	sqrtRatioX128 := new(ui.Int).Lsh(sqrtRatioX96, 32)
	msb := MostSignificantBit(sqrtRatioX128)
	var r *ui.Int
	if ui.NewInt(msb).Cmp(ui.NewInt(128)) >= 0 {
		r = new(ui.Int).Rsh(sqrtRatioX128, uint(msb-127))
	} else {
		r = new(ui.Int).Lsh(sqrtRatioX128, uint(127-msb))
	}

	log2 := new(ui.Int).Lsh(new(ui.Int).Sub(ui.NewInt(msb), ui.NewInt(128)), 64)

	for i := 0; i < 14; i++ {
		r = new(ui.Int).Rsh(new(ui.Int).Mul(r, r), 127)
		f := new(ui.Int).Rsh(r, 128)
		log2 = new(ui.Int).Or(log2, new(ui.Int).Lsh(f, uint(63-i)))
		r = new(ui.Int).Rsh(r, uint(f.Uint64()))
	}

	magicSqrt10001, _ := ui.FromHex("0x3627A301D71055774C85")
	logSqrt10001 := new(ui.Int).Mul(log2, magicSqrt10001)

	magicTickLow, _ := ui.FromHex("0x28F6481AB7F045A5AF012A19D003AAA")
	tickLow := int(new(ui.Int).Rsh(new(ui.Int).Sub(logSqrt10001, magicTickLow), 128).Uint64())
	magicTickHigh, _ := ui.FromHex("0xDB2DF09E81959A81455E260799A0632F")
	tickHigh := int(new(ui.Int).Rsh(new(ui.Int).Add(logSqrt10001, magicTickHigh), 128).Uint64())

	if tickLow == tickHigh {
		return tickLow
	}

	sqrtRatio := getSqrtRatioAtTick(tickHigh)
	if sqrtRatio.Cmp(sqrtRatioX96) <= 0 {
		return tickHigh
	} else {
		return tickLow
	}
}

func MostSignificantBit(x *ui.Int) uint64 {
	var msb uint64
	for _, power := range []int64{128, 64, 32, 16, 8, 4, 2, 1} {
		min := new(ui.Int).Exp(ui.NewInt(2), ui.NewInt(uint64(power)))
		if x.Cmp(min) >= 0 {
			x = new(ui.Int).Rsh(x, uint(power))
			msb += uint64(power)
		}
	}
	return msb
}

func mulShift(val *ui.Int, mulBy string) *ui.Int {
	mulByBig, _ := ui.FromHex(mulBy)

	return new(ui.Int).Rsh(new(ui.Int).Mul(val, mulByBig), 128)
}
