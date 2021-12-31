package prices

import (
	"math/big"
	"uniswap-simulator/lib/invariant"
	ui "uniswap-simulator/uint256"
)

type Prices struct {
	prices []*ui.Int
	index  int
	length int
}

func NewPrices(len int) *Prices {
	prices := make([]*ui.Int, len)
	for i := 0; i < len; i++ {
		prices[i] = ui.NewInt(0)
	}
	return &Prices{prices, 0, len}
}

func (p *Prices) Add(price *ui.Int) {
	//Square the price
	// For USDC - ETH Pool this never overflows. Absolutely not guaranteed otherwise.
	// Todo Fix this so it never overflows
	priceSquareX192 := new(ui.Int).Mul(price, price)

	shifted := new(ui.Int).Rsh(priceSquareX192, 230)
	invariant.Invariant(shifted.IsZero(), "Price square should not overflow")

	p.prices[p.index] = priceSquareX192
	p.index = (p.index + 1) % p.length
}

func (p *Prices) Average() *ui.Int {
	sum := new(ui.Int)
	for _, price := range p.prices {
		sum.Add(sum, price)
	}
	length := ui.NewInt(uint64(p.length))
	return new(ui.Int).Div(sum, length)
}

// Volatility o sqrt(variance)
func (p *Prices) Volatility() *ui.Int {
	avg := p.Average()
	sum := big.NewInt(0)
	for _, price := range p.prices {
		diff := new(ui.Int).Sub(price, avg)
		// IDK why its needed going to Square it later anyway
		diff.Abs(diff)
		diffBig := diff.ToBig()
		invariant.Invariant(diffBig.Sign() >= 0, "Volatility: diffBig is not negative")
		diff2 := new(big.Int).Mul(diffBig, diffBig)
		sum.Add(sum, diff2)
	}
	nMinus1 := big.NewInt(int64(p.length - 1))
	variance := new(big.Int).Div(sum, nMinus1)
	invariant.Invariant(variance.Sign() >= 0, "Volatility: variance is not negative")
	// X192 number
	volatility := new(big.Int).Sqrt(variance)
	invariant.Invariant(volatility.Sign() >= 0, "Volatility: volatility is not negative")
	ret, _ := ui.FromBig(volatility)
	return ret

}
