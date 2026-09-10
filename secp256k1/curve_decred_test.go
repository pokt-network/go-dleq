//go:build !ethereum_secp256k1

package secp256k1

import (
	"bytes"
	"sync"
	"testing"

	dsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestPointReadersDoNotMutate proves Encode, Equals and IsZero are read-only,
// so a point shared across goroutines (e.g. a ring member's public key) can be
// read concurrently. Regression: these methods called ToAffine on the shared
// point in place, a data race with every concurrent reader. Run with -race.
func TestPointReadersDoNotMutate(t *testing.T) {
	c := NewCurve()
	s, err := c.HashToScalar([]byte("shared point"))
	if err != nil {
		t.Fatal(err)
	}
	p := c.ScalarBaseMul(s)
	q := p.Copy()
	want := p.Encode()

	const workers, iterations = 8, 500
	var wg sync.WaitGroup
	for g := 0; g < workers; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if !bytes.Equal(p.Encode(), want) {
					t.Error("Encode changed under concurrent reads")
					return
				}
				if !p.Equals(q) || p.IsZero() {
					t.Error("Equals/IsZero wrong under concurrent reads")
					return
				}
				// Readers must not race with a ScalarMul that reads p.
				_ = p.ScalarMul(s)
			}
		}()
	}
	wg.Wait()
}

// TestNewPointFromCoordinatesNormalizes proves a point built from
// un-normalized field values still encodes and compares correctly, since the
// read-only accessors no longer normalize lazily.
func TestNewPointFromCoordinatesNormalizes(t *testing.T) {
	c := NewCurve()
	base := c.BasePoint().(*PointImpl)

	// Build un-normalized copies of the affine coordinates: X + X + (-X) is
	// still X, but with magnitude 3 and no normalization.
	var x, y dsecp256k1.FieldVal
	x.Set(&base.inner.X).Add(&base.inner.X).Add(new(dsecp256k1.FieldVal).NegateVal(&base.inner.X, 1))
	y.Set(&base.inner.Y).Add(&base.inner.Y).Add(new(dsecp256k1.FieldVal).NegateVal(&base.inner.Y, 1))

	p := NewPointFromCoordinates(x, y)
	if !bytes.Equal(p.Encode(), base.Encode()) {
		t.Fatal("Encode differs for un-normalized coordinates")
	}
	if !p.Equals(base) {
		t.Fatal("Equals false for un-normalized coordinates")
	}
	rt, err := c.DecodeToPoint(p.Encode())
	if err != nil {
		t.Fatal(err)
	}
	if !rt.Equals(base) {
		t.Fatal("round trip mismatch")
	}
}
