package dleq

import (
	"testing"

	"github.com/pokt-network/go-dleq/secp256k1"
)

// BenchmarkComparison_ScalarBaseMul compares backend performance for scalar base multiplication
func BenchmarkComparison_ScalarBaseMul(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = curve.ScalarBaseMul(scalar)
	}
}

// BenchmarkComparison_ScalarMul compares backend performance for scalar multiplication
func BenchmarkComparison_ScalarMul(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()
	point := curve.ScalarBaseMul(curve.ScalarFromInt(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = curve.ScalarMul(scalar, point)
	}
}

// BenchmarkComparison_Sign compares backend performance for signing
func BenchmarkComparison_Sign(b *testing.B) {
	curve := secp256k1.NewCurve()
	privKey := curve.NewRandomScalar()
	msgPoint := curve.BasePoint()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := curve.Sign(privKey, msgPoint)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComparison_Verify compares backend performance for verification
func BenchmarkComparison_Verify(b *testing.B) {
	curve := secp256k1.NewCurve()
	privKey := curve.NewRandomScalar()
	pubKey := curve.ScalarBaseMul(privKey)
	msgPoint := curve.BasePoint()

	sig, err := curve.Sign(privKey, msgPoint)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !curve.Verify(pubKey, msgPoint, sig) {
			b.Fatal("verification failed")
		}
	}
}

// BenchmarkComparison_DLEQProofGeneration compares backend performance for DLEQ proof generation
func BenchmarkComparison_DLEQProofGeneration(b *testing.B) {
	curveA := secp256k1.NewCurve()
	curveB := secp256k1.NewCurve() // Using same curve for comparison

	x, err := GenerateSecretForCurves(curveA, curveB)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewProof(curveA, curveB, x)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComparison_DLEQProofVerification compares backend performance for DLEQ proof verification
func BenchmarkComparison_DLEQProofVerification(b *testing.B) {
	curveA := secp256k1.NewCurve()
	curveB := secp256k1.NewCurve() // Using same curve for comparison

	x, err := GenerateSecretForCurves(curveA, curveB)
	if err != nil {
		b.Fatal(err)
	}

	proof, err := NewProof(curveA, curveB, x)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := proof.Verify(curveA, curveB)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComparison_ParallelScalarMul tests parallel performance
func BenchmarkComparison_ParallelScalarMul(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		curve := secp256k1.NewCurve()
		scalar := curve.NewRandomScalar()
		point := curve.ScalarBaseMul(curve.ScalarFromInt(2))

		for pb.Next() {
			_ = curve.ScalarMul(scalar, point)
		}
	})
}

// BenchmarkComparison_Memory measures memory usage patterns
func BenchmarkComparison_Memory(b *testing.B) {
	b.ReportAllocs()
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()
	point := curve.BasePoint()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := curve.ScalarMul(scalar, point)
		_ = result.Encode() // Force full evaluation
	}
}
