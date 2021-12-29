package prices

import ui "uniswap-simulator/uint256"

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
	priceSquareX192 := new(ui.Int).Mul(price, price)
	p.prices[p.index] = priceSquareX192
	p.index = (p.index + 1) % p.length
}

func (p *Prices) Average() *ui.Int {
	sum := ui.NewInt(0)
	for _, price := range p.prices {
		sum.Add(sum, price)
	}
	len := ui.NewInt(uint64(int64(p.length)))
	return new(ui.Int).Div(sum, len)
}
