package fullmath

import (
	"fmt"
	"math/big"
	"testing"
	ui "uniswap-simulator/uint256"
)

func Test(t *testing.T) {
	test1 := []string{"0", "500", "1000000"}
	var ui_arr [3]*ui.Int
	for i, str := range test1 {
		big_d, _ := new(big.Int).SetString(str, 10)
		ui_arr[i], _ = ui.FromBig(big_d)
	}
	res2 := MulDivRoundingUp(ui_arr[0], ui_arr[1], ui_arr[2])
	fmt.Printf("%d \n", res2)
}
