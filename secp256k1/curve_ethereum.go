//go:build cgo && ethereum_secp256k1
// +build cgo,ethereum_secp256k1

package secp256k1

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"

	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/sha3"

	"github.com/athanorlabs/go-dleq/types"
)

type Curve = types.Curve
type Point = types.Point
type Scalar = types.Scalar

var _ Curve = &CurveImpl{}
var _ Scalar = &ScalarImpl{}
var _ Point = &PointImpl{}

type CurveImpl struct {
	order        *big.Int
	basePoint    Point
	altBasePoint Point
}

func NewCurve() Curve {
	// TODO_READABILITY: Consider using a more readable format for the order constant
	orderBytes, err := hex.DecodeString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141")
	if err != nil {
		panic(err)
	}

	return &CurveImpl{
		order:        new(big.Int).SetBytes(orderBytes),
		basePoint:    basePoint(),
		altBasePoint: altBasePoint(),
	}
}

func basePoint() Point {
	// Generator point for secp256k1
	gx, _ := new(big.Int).SetString("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", 16)
	gy, _ := new(big.Int).SetString("483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8", 16)

	return &PointImpl{
		x: gx,
		y: gy,
	}
}

func altBasePoint() Point {
	const str = "0250929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0"
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}

	// Parse compressed point
	if len(b) != 33 || (b[0] != 0x02 && b[0] != 0x03) {
		panic("invalid compressed point")
	}

	x := new(big.Int).SetBytes(b[1:])
	y := decompressPoint(x, b[0] == 0x03)

	return &PointImpl{
		x: x,
		y: y,
	}
}

// decompressPoint recovers Y coordinate from X coordinate
func decompressPoint(x *big.Int, isOdd bool) *big.Int {
	// secp256k1: y² = x³ + 7
	curve := ethsecp256k1.S256()
	p := curve.Params().P

	// Compute x³ + 7
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Add(x3, big.NewInt(7))
	x3.Mod(x3, p)

	// Compute square root
	y := new(big.Int).ModSqrt(x3, p)
	if y == nil {
		panic("invalid point: no square root")
	}

	// Choose correct sign
	if (y.Bit(0) == 1) != isOdd {
		y.Sub(p, y)
	}

	return y
}

func (*CurveImpl) BitSize() uint64 {
	return 255
}

func (*CurveImpl) CompressedPointSize() int {
	return 33
}

func (*CurveImpl) DecodeToPoint(in []byte) (Point, error) {
	if len(in) != 33 {
		return nil, errors.New("invalid compressed point length")
	}

	cp := make([]byte, len(in))
	copy(cp, in)

	if cp[0] != 0x02 && cp[0] != 0x03 {
		return nil, errors.New("invalid compressed point format")
	}

	x := new(big.Int).SetBytes(cp[1:])
	y := decompressPoint(x, cp[0] == 0x03)

	return &PointImpl{
		x: x,
		y: y,
	}, nil
}

func (*CurveImpl) DecodeToScalar(in []byte) (Scalar, error) {
	if len(in) != 32 {
		return nil, errors.New("invalid scalar length")
	}

	cp := make([]byte, len(in))
	copy(cp, in)

	return &ScalarImpl{
		value: new(big.Int).SetBytes(cp),
	}, nil
}

func (c *CurveImpl) BasePoint() Point {
	return c.basePoint
}

func (c *CurveImpl) AltBasePoint() Point {
	return c.altBasePoint
}

func (*CurveImpl) NewRandomScalar() Scalar {
	var b [32]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}

	return &ScalarImpl{
		value: new(big.Int).SetBytes(b[:]),
	}
}

func reverse(in [32]byte) [32]byte {
	rs := [32]byte{}
	for i := 0; i < 32; i++ {
		rs[i] = in[32-i-1]
	}
	return rs
}

// ScalarFromBytes sets a Scalar from LE bytes.
func (*CurveImpl) ScalarFromBytes(b [32]byte) Scalar {
	// reverse bytes, since we're getting LE bytes but need BE
	in := reverse(b)
	return &ScalarImpl{
		value: new(big.Int).SetBytes(in[:]),
	}
}

func (*CurveImpl) ScalarFromInt(in uint32) Scalar {
	return &ScalarImpl{
		value: big.NewInt(int64(in)),
	}
}

func (c *CurveImpl) HashToScalar(in []byte) (Scalar, error) {
	h := sha3.Sum512(in)
	n := new(big.Int).SetBytes(h[:])
	n = new(big.Int).Mod(n, c.order)

	return &ScalarImpl{
		value: n,
	}, nil
}

// ScalarBaseMul uses go-ethereum's optimized scalar base multiplication
func (*CurveImpl) ScalarBaseMul(s Scalar) Point {
	ss, ok := s.(*ScalarImpl)
	if !ok {
		// TODO_IMPROVE: Consider returning error instead of panic for better error handling
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	// Convert scalar to 32-byte array using pooled buffer
	scalarBytes := getBytes32()
	defer putBytes32(scalarBytes)
	ss.value.FillBytes(scalarBytes)

	// Use Ethereum's optimized scalar base multiplication
	// TODO_OPTIMIZE: ScalarBaseMul is currently slower than Decred (43μs vs 36μs)
	// Consider using direct libsecp256k1 calls instead of go-ethereum wrapper
	x, y := ethsecp256k1.S256().ScalarBaseMult(scalarBytes)

	return &PointImpl{
		x: x,
		y: y,
	}
}

// ScalarMul uses go-ethereum's optimized scalar multiplication
// TODO_IMPROVE: Add nil checks for s and p parameters to prevent runtime panics
func (*CurveImpl) ScalarMul(s Scalar, p Point) Point {
	ss, ok := s.(*ScalarImpl)
	if !ok {
		// TODO_IMPROVE: Consider returning error instead of panic for better error handling
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	pp, ok := p.(*PointImpl)
	if !ok {
		// TODO_IMPROVE: Consider returning error instead of panic for better error handling
		panic("invalid point; type is not *secp256k1.PointImpl")
	}

	// Convert scalar to 32-byte array using pooled buffer
	scalarBytes := getBytes32()
	defer putBytes32(scalarBytes)
	ss.value.FillBytes(scalarBytes)

	// Use Ethereum's optimized scalar multiplication
	x, y := ethsecp256k1.S256().ScalarMult(pp.x, pp.y, scalarBytes)

	return &PointImpl{
		x: x,
		y: y,
	}
}

// Sign accepts a private key `s` and signs the encoded point `p`.
func (*CurveImpl) Sign(s Scalar, p Point) ([]byte, error) {
	ss, ok := s.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	// Convert scalar to 32-byte private key using pooled buffer
	privKeyBytes := getBytes32()
	defer putBytes32(privKeyBytes)
	ss.value.FillBytes(privKeyBytes)

	// Get message to sign
	msg := p.Encode()
	hash := sha256.Sum256(msg)

	// Use Ethereum's secp256k1 signing
	sig, err := ethsecp256k1.Sign(hash[:], privKeyBytes)
	if err != nil {
		return nil, err
	}

	// Convert to DER format for compatibility using pooled big.Int
	r := getBigInt()
	defer putBigInt(r)
	r.SetBytes(sig[:32])

	s2 := getBigInt()
	defer putBigInt(s2)
	s2.SetBytes(sig[32:64])

	return encodeDER(r, s2), nil
}

func (*CurveImpl) Verify(pubkey, msgPoint Point, sig []byte) bool {
	pp, ok := pubkey.(*PointImpl)
	if !ok {
		panic("invalid point; type is not *secp256k1.PointImpl")
	}

	// Decode DER signature
	r, s, err := decodeDER(sig)
	if err != nil {
		return false
	}

	// Convert to Ethereum format (64 bytes) using pooled buffer
	ethSig := getBytes64()
	defer putBytes64(ethSig)
	r.FillBytes(ethSig[:32])
	s.FillBytes(ethSig[32:64])

	// Encode public key using pooled buffer
	pubKeyBytes := getBytes65()
	defer putBytes65(pubKeyBytes)
	pubKeyBytes[0] = 0x04 // uncompressed
	pp.x.FillBytes(pubKeyBytes[1:33])
	pp.y.FillBytes(pubKeyBytes[33:65])

	msg := msgPoint.Encode()
	hash := sha256.Sum256(msg)

	// Use Ethereum's verification
	return ethsecp256k1.VerifySignature(pubKeyBytes, hash[:], ethSig)
}

// encodeDER encodes r,s signature components in DER format
func encodeDER(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Add leading zero if high bit is set
	if len(rBytes) > 0 && rBytes[0] >= 0x80 {
		rBytes = append([]byte{0}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0] >= 0x80 {
		sBytes = append([]byte{0}, sBytes...)
	}

	totalLen := 4 + len(rBytes) + len(sBytes) // 4 bytes for DER headers

	der := make([]byte, 0, totalLen+2)
	der = append(der, 0x30, byte(totalLen))    // SEQUENCE header
	der = append(der, 0x02, byte(len(rBytes))) // INTEGER header for r
	der = append(der, rBytes...)
	der = append(der, 0x02, byte(len(sBytes))) // INTEGER header for s
	der = append(der, sBytes...)

	return der
}

// decodeDER decodes DER signature to r,s components
func decodeDER(sig []byte) (*big.Int, *big.Int, error) {
	if len(sig) < 6 {
		return nil, nil, errors.New("signature too short")
	}

	if sig[0] != 0x30 {
		return nil, nil, errors.New("invalid DER signature")
	}

	offset := 2
	if offset >= len(sig) {
		return nil, nil, errors.New("invalid signature")
	}

	// Read r
	if sig[offset] != 0x02 {
		return nil, nil, errors.New("invalid r component")
	}
	offset++

	rLen := int(sig[offset])
	offset++

	if offset+rLen > len(sig) {
		return nil, nil, errors.New("invalid r length")
	}

	r := new(big.Int).SetBytes(sig[offset : offset+rLen])
	offset += rLen

	// Read s
	if offset >= len(sig) || sig[offset] != 0x02 {
		return nil, nil, errors.New("invalid s component")
	}
	offset++

	sLen := int(sig[offset])
	offset++

	if offset+sLen != len(sig) {
		return nil, nil, errors.New("invalid s length")
	}

	s := new(big.Int).SetBytes(sig[offset:])

	return r, s, nil
}

type ScalarImpl struct {
	value *big.Int
}

func (s *ScalarImpl) Add(b Scalar) Scalar {
	ss, ok := b.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	result := getBigInt()
	result.Add(s.value, ss.value)
	curve := ethsecp256k1.S256()
	result.Mod(result, curve.Params().N)

	return &ScalarImpl{
		value: result,
	}
}

func (s *ScalarImpl) Sub(b Scalar) Scalar {
	ss, ok := b.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	result := getBigInt()
	result.Sub(s.value, ss.value)
	curve := ethsecp256k1.S256()
	result.Mod(result, curve.Params().N)

	return &ScalarImpl{
		value: result,
	}
}

func (s *ScalarImpl) Negate() Scalar {
	curve := ethsecp256k1.S256()
	result := getBigInt()
	result.Sub(curve.Params().N, s.value)
	result.Mod(result, curve.Params().N)

	return &ScalarImpl{
		value: result,
	}
}

func (s *ScalarImpl) Mul(b Scalar) Scalar {
	ss, ok := b.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	result := getBigInt()
	result.Mul(s.value, ss.value)
	curve := ethsecp256k1.S256()
	result.Mod(result, curve.Params().N)

	return &ScalarImpl{
		value: result,
	}
}

func (s *ScalarImpl) Inverse() Scalar {
	curve := ethsecp256k1.S256()
	result := getBigInt()
	result.ModInverse(s.value, curve.Params().N)
	if result == nil {
		putBigInt(result) // Return to pool before panicking
		panic("scalar has no inverse")
	}

	return &ScalarImpl{
		value: result,
	}
}

func (s *ScalarImpl) Encode() []byte {
	b := make([]byte, 32)
	s.value.FillBytes(b)
	return b
}

func (s *ScalarImpl) Eq(other Scalar) bool {
	o, ok := other.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	return s.value.Cmp(o.value) == 0
}

func (s *ScalarImpl) IsZero() bool {
	return s.value.Sign() == 0
}

type PointImpl struct {
	x, y *big.Int
}

func NewPointFromCoordinates(x, y *big.Int) *PointImpl {
	return &PointImpl{
		x: new(big.Int).Set(x),
		y: new(big.Int).Set(y),
	}
}

func (p *PointImpl) Copy() Point {
	return &PointImpl{
		x: new(big.Int).Set(p.x),
		y: new(big.Int).Set(p.y),
	}
}

func (p *PointImpl) Add(b Point) Point {
	pp, ok := b.(*PointImpl)
	if !ok {
		panic("invalid point; type is not *secp256k1.PointImpl")
	}

	// Handle nil coordinates by initializing to zero
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}

	ppx, ppy := pp.x, pp.y
	if ppx == nil {
		ppx = big.NewInt(0)
	}
	if ppy == nil {
		ppy = big.NewInt(0)
	}

	curve := ethsecp256k1.S256()
	x, y := curve.Add(px, py, ppx, ppy)

	return &PointImpl{
		x: x,
		y: y,
	}
}

func (p *PointImpl) Sub(b Point) Point {
	pp, ok := b.(*PointImpl)
	if !ok {
		panic("invalid point; type is not *secp256k1.PointImpl")
	}

	// Handle nil coordinates
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}

	ppx, ppy := pp.x, pp.y
	if ppx == nil {
		ppx = big.NewInt(0)
	}
	if ppy == nil {
		ppy = big.NewInt(0)
	}

	// Negate the point and add
	curve := ethsecp256k1.S256()
	negY := new(big.Int).Sub(curve.Params().P, ppy)
	x, y := curve.Add(px, py, ppx, negY)

	return &PointImpl{
		x: x,
		y: y,
	}
}

func (p *PointImpl) ScalarMul(s Scalar) Point {
	ss, ok := s.(*ScalarImpl)
	if !ok {
		panic("invalid scalar; type is not *secp256k1.ScalarImpl")
	}

	// Handle nil coordinates
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}

	// Convert scalar to bytes
	scalarBytes := make([]byte, 32)
	ss.value.FillBytes(scalarBytes)

	// Use Ethereum's scalar multiplication
	x, y := ethsecp256k1.S256().ScalarMult(px, py, scalarBytes)

	return &PointImpl{
		x: x,
		y: y,
	}
}

func (p *PointImpl) Encode() []byte {
	// Handle nil coordinates
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}

	// Return compressed point encoding
	compressed := make([]byte, 33)
	if py.Bit(0) == 1 {
		compressed[0] = 0x03
	} else {
		compressed[0] = 0x02
	}
	px.FillBytes(compressed[1:])
	return compressed
}

func (p *PointImpl) IsZero() bool {
	// Handle nil coordinates
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}
	return px.Sign() == 0 && py.Sign() == 0
}

func (p *PointImpl) Equals(other Point) bool {
	pp, ok := other.(*PointImpl)
	if !ok {
		panic("invalid point; type is not *secp256k1.PointImpl")
	}

	// Handle nil coordinates
	px, py := p.x, p.y
	if px == nil {
		px = big.NewInt(0)
	}
	if py == nil {
		py = big.NewInt(0)
	}

	ppx, ppy := pp.x, pp.y
	if ppx == nil {
		ppx = big.NewInt(0)
	}
	if ppy == nil {
		ppy = big.NewInt(0)
	}

	return px.Cmp(ppx) == 0 && py.Cmp(ppy) == 0
}
