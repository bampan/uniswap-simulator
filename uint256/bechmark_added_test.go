package uint256

import (
	"math/big"
	"testing"
)

func Benchmark_ChangeSign(bench *testing.B) {
	bench.Run("single/MultiplybyOne", benchmarkMultiplybyone)
	bench.Run("single/ChangeSign", benchmarkChangeSign)
}

func benchmarkChangeSign(bench *testing.B) {
	b1 := big.NewInt(0).SetBytes(hex2Bytes("0123456789abcdeffedcba9876543210f2f3f4f5f6f7f8f9fff3f4f5f6f7f8f9"))
	f, _ := FromBig(b1)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		f.ChangeSign(f)
	}
}

func benchmarkMultiplybyone(bench *testing.B) {
	b1 := big.NewInt(0).SetBytes(hex2Bytes("0123456789abcdeffedcba9876543210f2f3f4f5f6f7f8f9fff3f4f5f6f7f8f9"))
	b2 := big.NewInt(-1)
	f, _ := FromBig(b1)
	f2, _ := FromBig(b2)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		f.Mul(f, f2)
	}
}
