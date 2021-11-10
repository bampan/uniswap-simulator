package tickmath

import (
	"fmt"
	"testing"
)

// UpperBound ...
func UpperBound(array []int, target int) int {
	low, high, mid := 0, len(array)-1, 0

	for low < high {
		mid = (low + high + 1) / 2
		if array[mid] > target {
			high = mid - 1
		} else {
			low = mid
		}
	}
	return low
}

func TestUpperBound(t *testing.T) {
	type args struct {
		array  []int
		target int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{
			name: "test0",
			args: args{array: []int{1, 3, 5}, target: 2},
			want: 0,
		},
		{
			name: "test2",
			args: args{array: []int{1, 3, 5}, target: 4},
			want: 1,
		},
		{
			name: "test1",
			args: args{array: []int{1, 3, 5}, target: 3},
			want: 1,
		},
		{
			name: "test1",
			args: args{array: []int{1, 3, 5}, target: 5},
			want: 2,
		},
		{
			name: "test1",
			args: args{array: []int{1, 3, 5}, target: 6},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(TotalTicks)
			if got := UpperBound(tt.args.array, tt.args.target); got != tt.want {
				t.Errorf("UpperBound() = %v, want %v", got, tt.want)
			}
		})
	}
}
