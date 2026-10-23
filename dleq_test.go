package dleq

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/go-dleq/ed25519"
	"github.com/pokt-network/go-dleq/secp256k1"
)

func TestWitnessSize(t *testing.T) {
	curveA := secp256k1.NewCurve()
	curveB := ed25519.NewCurve()
	x, err := GenerateSecretForCurves(curveA, curveB)
	require.NoError(t, err)
	err = checkWitnessSize(x, min(curveA.BitSize(), curveB.BitSize()))
	require.NoError(t, err)

	x = [32]byte{}
	x[31] = 0xff
	err = checkWitnessSize(x, 255)
	require.Error(t, err)

	x = [32]byte{}
	x[30] = 0xff
	err = checkWitnessSize(x, 247)
	require.Error(t, err)

	x = [32]byte{}
	x[30] = 0b00010101
	err = checkWitnessSize(x, 244)
	require.Error(t, err)
	err = checkWitnessSize(x, 245)
	require.NoError(t, err)
}

func TestGenerateCommitments(t *testing.T) {
	curve := secp256k1.NewCurve()
	x, err := generateRandomBits(curve.BitSize())
	require.NoError(t, err)
	commitments, err := generateCommitments(curve, x[:], curve.BitSize())
	require.NoError(t, err)
	require.Equal(t, int(curve.BitSize()), len(commitments))

	X := curve.ScalarBaseMul(curve.ScalarFromBytes(x))
	err = verifyCommitmentsSum(curve, commitments, X)
	require.NoError(t, err)
}

func TestGenerateRingSignature(t *testing.T) {
	curve := secp256k1.NewCurve()
	x, err := generateRandomBits(curve.BitSize())
	require.NoError(t, err)
	commitmentsA, err := generateCommitments(curve, x[:], curve.BitSize())
	require.NoError(t, err)
	require.Equal(t, int(curve.BitSize()), len(commitmentsA))
	commitmentsB, err := generateCommitments(curve, x[:], curve.BitSize())
	require.NoError(t, err)
	require.Equal(t, int(curve.BitSize()), len(commitmentsB))

	for i := 0; i < int(curve.BitSize()); i++ {
		bit := getBit(x[:], uint64(i))
		_, err := generateRingSignature(curve, curve, bit, commitmentsA[i], commitmentsB[i])
		require.NoError(t, err)
	}
}

func TestProveAndVerify(t *testing.T) {
	curveA := secp256k1.NewCurve()
	curveB := ed25519.NewCurve()
	x, err := GenerateSecretForCurves(curveA, curveB)
	require.NoError(t, err)
	proof, err := NewProof(curveA, curveB, x)
	require.NoError(t, err)
	err = proof.Verify(curveA, curveB)
	require.NoError(t, err)
}

func powerOfTwo(curve Curve, n int) Scalar {
	p := curve.ScalarFromInt(1)
	two := curve.ScalarFromInt(2)
	for i := 0; i < n; i++ {
		p = p.Mul(two)
	}
	return p
}

func TestVerifyRejectsExtraBitProof(t *testing.T) {
	curveA := secp256k1.NewCurve()
	curveB := ed25519.NewCurve()
	x, err := GenerateSecretForCurves(curveA, curveB)
	require.NoError(t, err)
	proof, err := NewProof(curveA, curveB, x)
	require.NoError(t, err)

	bits := int(min(curveA.BitSize(), curveB.BitSize()))
	dA := curveA.ScalarFromInt(1)
	dB := curveB.ScalarFromInt(7)
	proof.proofs = append(proof.proofs, bitProof{
		commitmentA: commitment{commitment: curveA.ScalarBaseMul(dA.Mul(powerOfTwo(curveA, bits).Inverse()))},
		commitmentB: commitment{commitment: curveB.ScalarBaseMul(dB.Mul(powerOfTwo(curveB, bits).Inverse()))},
		ringSig:     proof.proofs[0].ringSig,
	})
	xA := curveA.ScalarFromBytes(x).Add(dA)
	xB := curveB.ScalarFromBytes(x).Add(dB)
	proof.CommitmentA = curveA.ScalarBaseMul(xA)
	proof.CommitmentB = curveB.ScalarBaseMul(xB)
	proof.signatureA.inner, err = curveA.Sign(xA, proof.CommitmentA)
	require.NoError(t, err)
	proof.signatureB.inner, err = curveB.Sign(xB, proof.CommitmentB)
	require.NoError(t, err)

	var decoded Proof
	require.NoError(t, decoded.Deserialize(curveA, curveB, proof.Serialize()))
	require.Error(t, decoded.Verify(curveA, curveB))
}

func TestVerifyRejectsShortProof(t *testing.T) {
	curveA := secp256k1.NewCurve()
	curveB := ed25519.NewCurve()
	x, err := GenerateSecretForCurves(curveA, curveB)
	require.NoError(t, err)
	proof, err := NewProof(curveA, curveB, x)
	require.NoError(t, err)

	for _, n := range []int{0, 1, 10} {
		short := *proof
		short.proofs = proof.proofs[:n]
		require.NotPanics(t, func() {
			require.Error(t, short.Verify(curveA, curveB))
		})
	}
}
