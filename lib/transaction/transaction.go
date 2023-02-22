package transaction

import (
	"encoding/json"
	"strconv"

	ui "github.com/holiman/uint256"
)

type TransactionInput struct {
	Type         string `json:"type"`
	ID           string `json:"id"`
	Timestamp    int    `json:"timestamp"`
	Amount0      string `json:"amount0"`
	Amount1      string `json:"amount1"`
	Amount       string `json:"amount,omitempty"`
	SqrtPriceX96 string `json:"sqrtPriceX96,omitempty"`
	Tick         int    `json:"tick,omitempty"`
	TickLower    int    `json:"tickLower,omitempty"`
	TickUpper    int    `json:"tickUpper,omitempty"`
	UseX96       string `json:"useX96,omitempty"`
}

type Transaction struct {
	Type         string
	Amount       *ui.Int
	Amount0      *ui.Int
	Amount1      *ui.Int
	ID           string
	SqrtPriceX96 *ui.Int
	Tick         int
	TickLower    int
	TickUpper    int
	Timestamp    int
	UseX96       bool
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	switch t.Type {
	case "Swap":
		return json.Marshal(&TransactionInput{
			Type:         t.Type,
			Amount0:      t.Amount0.String(),
			Amount1:      t.Amount1.String(),
			ID:           t.ID,
			SqrtPriceX96: t.SqrtPriceX96.String(),
			Tick:         t.Tick,
			Timestamp:    t.Timestamp,
			UseX96:       strconv.FormatBool(t.UseX96),
		})
	case "Mint", "Burn":
		return json.Marshal(&TransactionInput{
			Type:      t.Type,
			Amount:    t.Amount.String(),
			Amount1:   t.Amount1.String(),
			Amount0:   t.Amount0.String(),
			TickLower: t.TickLower,
			TickUpper: t.TickUpper,
			ID:        t.ID,
			Timestamp: t.Timestamp,
		})
	case "Flash":
		return json.Marshal(&TransactionInput{
			Type:      t.Type,
			Amount0:   t.Amount0.String(),
			Amount1:   t.Amount1.String(),
			ID:        t.ID,
			Timestamp: t.Timestamp,
		})
	}
	panic("unreachable")

}
