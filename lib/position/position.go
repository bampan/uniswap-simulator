package position

import (
	cons "github.com/ftchann/uniswap-simulator/lib/constants"
	"github.com/ftchann/uniswap-simulator/lib/fullmath"

	ui "github.com/holiman/uint256"
)

func (i *Info) Clone() *Info {
	return &Info{
		Liquidity:                i.Liquidity.Clone(),
		FeeGrowthInside0LastX128: i.FeeGrowthInside0LastX128.Clone(),
		FeeGrowthInside1LastX128: i.FeeGrowthInside1LastX128.Clone(),
		TokensOwed0:              i.TokensOwed1.Clone(),
		TokensOwed1:              i.TokensOwed1.Clone(),
	}
}

type Info struct {
	Liquidity                *ui.Int
	FeeGrowthInside0LastX128 *ui.Int
	FeeGrowthInside1LastX128 *ui.Int
	TokensOwed0              *ui.Int
	TokensOwed1              *ui.Int
}

func NewPosition() *Info {
	return &Info{
		Liquidity:                ui.NewInt(0),
		FeeGrowthInside0LastX128: ui.NewInt(0),
		FeeGrowthInside1LastX128: ui.NewInt(0),
		TokensOwed0:              ui.NewInt(0),
		TokensOwed1:              ui.NewInt(0),
	}
}

func (i *Info) Update(LiquidityDelta, FeeGrowthInside0X128, FeeGrowthInside1X128 *ui.Int) {
	LiquidityNext := new(ui.Int).Add(i.Liquidity, LiquidityDelta)
	temp0 := new(ui.Int).Sub(FeeGrowthInside0X128, i.FeeGrowthInside0LastX128)
	temp1 := new(ui.Int).Sub(FeeGrowthInside1X128, i.FeeGrowthInside1LastX128)
	TokensOwed0 := fullmath.MulDiv(temp0, i.Liquidity, cons.Q128)
	TokensOwed1 := fullmath.MulDiv(temp1, i.Liquidity, cons.Q128)
	i.Liquidity = LiquidityNext
	i.FeeGrowthInside0LastX128 = FeeGrowthInside0X128
	i.FeeGrowthInside1LastX128 = FeeGrowthInside1X128
	i.TokensOwed0.Add(i.TokensOwed0, TokensOwed0)
	i.TokensOwed1.Add(i.TokensOwed1, TokensOwed1)
}
