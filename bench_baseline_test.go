package dleq

import (
	"testing"

	"github.com/pokt-network/go-dleq/secp256k1"
)

// BenchmarkScalarBaseMul benchmarks the critical ScalarBaseMul operation
func BenchmarkScalarBaseMul(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = curve.ScalarBaseMul(scalar)
	}
}

// BenchmarkScalarMul benchmarks the critical ScalarMul operation
func BenchmarkScalarMul(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()
	point := curve.ScalarBaseMul(curve.ScalarFromInt(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = curve.ScalarMul(scalar, point)
	}
}

// BenchmarkPointScalarMul benchmarks point scalar multiplication
func BenchmarkPointScalarMul(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()
	point := curve.BasePoint()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = point.ScalarMul(scalar)
	}
}

// BenchmarkSign benchmarks signing operation
func BenchmarkSign(b *testing.B) {
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

// BenchmarkVerify benchmarks verification operation
func BenchmarkVerify(b *testing.B) {
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

// BenchmarkDLEQProofGeneration benchmarks full DLEQ proof generation
func BenchmarkDLEQProofGeneration(b *testing.B) {
	curveA := secp256k1.NewCurve()
	curveB := secp256k1.NewCurve() // Using same curve for simplicity

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

// BenchmarkDLEQProofVerification benchmarks full DLEQ proof verification
func BenchmarkDLEQProofVerification(b *testing.B) {
	curveA := secp256k1.NewCurve()
	curveB := secp256k1.NewCurve() // Using same curve for simplicity

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

// BenchmarkScalarDecoding benchmarks scalar decoding from bytes
func BenchmarkScalarDecoding(b *testing.B) {
	curve := secp256k1.NewCurve()
	scalar := curve.NewRandomScalar()
	encoded := scalar.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := curve.DecodeToScalar(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPointDecoding benchmarks point decoding from bytes
func BenchmarkPointDecoding(b *testing.B) {
	curve := secp256k1.NewCurve()
	point := curve.BasePoint()
	encoded := point.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := curve.DecodeToPoint(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}
