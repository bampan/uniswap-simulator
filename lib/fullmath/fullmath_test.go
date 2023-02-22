package fullmath

import (
	"fmt"
	"testing"

	ui "github.com/holiman/uint256"
)

func TestMulDivRoundingUp(t *testing.T) {
	//TODO test for real big uint256
	tests := [][]uint64{
		{0, 500, 1000000, 0},
		{1, 500, 1000000, 1},
		{1000000, 1, 1000000, 1},
		{1000001, 1, 1000000, 2},
	}
	for _, arg := range tests {
		t.Run(fmt.Sprint(arg), func(t *testing.T) {
			result := MulDivRoundingUp(ui.NewInt(arg[0]), ui.NewInt(arg[1]), ui.NewInt(arg[2]))
			if ui.NewInt(arg[3]).Cmp(result) != 0 {
				t.Fatalf("want=%v result=%v", arg[3], result)
			}
		})
	}
}
