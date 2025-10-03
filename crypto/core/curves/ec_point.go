//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package curves

import (
	"crypto/elliptic"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/sonr-io/sonr/crypto/core"

	"github.com/dustinxie/ecc"
	"github.com/sonr-io/sonr/crypto/internal"
)

var curveNameToID = map[string]byte{
	"secp256k1": 0,
	"P-224":     1,
	"P-256":     2,
	"P-384":     3,
	"P-521":     4,
}

var curveIDToName = map[byte]func() elliptic.Curve{
	0: ecc.P256k1,
	1: elliptic.P224,
	2: elliptic.P256,
	3: elliptic.P384,
	4: elliptic.P521,
}

var curveMapper = map[string]func() elliptic.Curve{
	"secp256k1": ecc.P256k1,
	"P-224":     elliptic.P224,
	"P-256":     elliptic.P256,
	"P-384":     elliptic.P384,
	"P-521":     elliptic.P521,
}

// EcPoint represents an elliptic curve Point
type EcPoint struct {
	Curve elliptic.Curve
	X, Y  *big.Int
}

// EcPointJSON encapsulates the data that is serialized to JSON
// used internally and not for external use. Public so other pieces
// can use for serialization
type EcPointJSON struct {
	X         *big.Int `json:"x"`
	Y         *big.Int `json:"y"`
	CurveName string   `json:"curve_name"`
}

// MarshalJSON serializes EcPoint to JSON
func (a EcPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(EcPointJSON{
		CurveName: a.Curve.Params().Name,
		X:         a.X,
		Y:         a.Y,
	})
}

// UnmarshalJSON deserializes JSON to EcPoint
func (a *EcPoint) UnmarshalJSON(bytes []byte) error {
	data := new(EcPointJSON)
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return err
	}
	if mapper, ok := curveMapper[data.CurveName]; ok {
		a.Curve = mapper()
		a.X = data.X
		a.Y = data.Y
		return nil
	}
	return fmt.Errorf("unknown curve deserialized")
}

// MarshalBinary serializes EcPoint to binary
func (a *EcPoint) MarshalBinary() ([]byte, error) {
	result := [65]byte{}
	if code, ok := curveNameToID[a.Curve.Params().Name]; ok {
		result[0] = code
		a.X.FillBytes(result[1:33])
		a.Y.FillBytes(result[33:65])
		return result[:], nil
	}
	return nil, fmt.Errorf("unknown curve serialized")
}

// UnmarshalBinary deserializes binary to EcPoint
func (a *EcPoint) UnmarshalBinary(data []byte) error {
	if mapper, ok := curveIDToName[data[0]]; ok {
		a.Curve = mapper()
		a.X = new(big.Int).SetBytes(data[1:33])
		a.Y = new(big.Int).SetBytes(data[33:65])
		return nil
	}
	return fmt.Errorf("unknown curve deserialized")
}

// IsValid checks if the point is valid
func (a EcPoint) IsValid() bool {
	return a.IsOnCurve() || a.IsIdentity()
}

// IsOnCurve checks if the point is on the curve
func (a EcPoint) IsOnCurve() bool {
	return a.Curve.IsOnCurve(a.X, a.Y)
}

// IsIdentity returns true if this Point is the Point at infinity
func (a EcPoint) IsIdentity() bool {
	x := core.ConstantTimeEqByte(a.X, core.Zero)
	y := core.ConstantTimeEqByte(a.Y, core.Zero)
	return (x & y) == 1
}

// Equals return true if a and b have the same x,y coordinates
func (a EcPoint) Equals(b *EcPoint) bool {
	if !sameCurve(&a, b) {
		return false
	}
	x := core.ConstantTimeEqByte(a.X, b.X)
	y := core.ConstantTimeEqByte(a.Y, b.Y)
	return (x & y) == 1
}

// IsBasePoint returns true if this Point is curve's base Point
func (a EcPoint) IsBasePoint() bool {
	p := a.Curve.Params()
	x := core.ConstantTimeEqByte(a.X, p.Gx)
	y := core.ConstantTimeEqByte(a.Y, p.Gy)
	return (x & y) == 1
}

// reduceModN normalizes the Scalar to a positive element smaller than the base Point order.
func reduceModN(curve elliptic.Curve, k *big.Int) *big.Int {
	return new(big.Int).Mod(k, curve.Params().N)
}

// Add performs elliptic curve addition on two points
func (a *EcPoint) Add(b *EcPoint) (*EcPoint, error) {
	if a == nil || b == nil {
		return nil, internal.ErrNilArguments
	}
	if !sameCurve(a, b) {
		return nil, internal.ErrPointsDistinctCurves
	}
	p := &EcPoint{Curve: a.Curve}
	p.X, p.Y = a.Curve.Add(a.X, a.Y, b.X, b.Y)
	if !p.IsValid() {
		return nil, internal.ErrNotOnCurve
	}
	return p, nil
}

// Neg returns the negation of a Weierstrass Point.
func (a *EcPoint) Neg() (*EcPoint, error) {
	if a == nil {
		return nil, internal.ErrNilArguments
	}
	p := &EcPoint{Curve: a.Curve, X: a.X, Y: new(big.Int).Sub(a.Curve.Params().P, a.Y)}
	if !p.IsValid() {
		return nil, internal.ErrNotOnCurve
	}
	return p, nil
}

// ScalarMult multiplies this Point by a Scalar
func (a *EcPoint) ScalarMult(k *big.Int) (*EcPoint, error) {
	if a == nil || k == nil {
		return nil, fmt.Errorf("cannot multiply nil Point or element")
	}
	n := reduceModN(a.Curve, k)
	p := new(EcPoint)
	p.Curve = a.Curve
	p.X, p.Y = a.Curve.ScalarMult(a.X, a.Y, n.Bytes())
	if !p.IsValid() {
		return nil, fmt.Errorf("result not on the curve")
	}
	return p, nil
}

// NewScalarBaseMult creates a Point from the base Point multiplied by a field element
func NewScalarBaseMult(curve elliptic.Curve, k *big.Int) (*EcPoint, error) {
	if curve == nil || k == nil {
		return nil, fmt.Errorf("nil parameters are not supported")
	}
	n := reduceModN(curve, k)
	p := new(EcPoint)
	p.Curve = curve
	p.X, p.Y = curve.ScalarBaseMult(n.Bytes())
	if !p.IsValid() {
		return nil, fmt.Errorf("result not on the curve")
	}
	return p, nil
}

// Bytes returns the bytes represented by this Point with x || y
func (a EcPoint) Bytes() []byte {
	fieldSize := internal.CalcFieldSize(a.Curve)
	out := make([]byte, fieldSize*2)

	a.X.FillBytes(out[0:fieldSize])
	a.Y.FillBytes(out[fieldSize : fieldSize*2])
	return out
}

// PointFromBytesUncompressed outputs uncompressed X || Y similar to
// https://www.secg.org/sec1-v1.99.dif.pdf section 2.2 and 2.3
func PointFromBytesUncompressed(curve elliptic.Curve, b []byte) (*EcPoint, error) {
	fieldSize := internal.CalcFieldSize(curve)
	if len(b) != fieldSize*2 {
		return nil, fmt.Errorf("invalid number of bytes")
	}
	p := &EcPoint{
		Curve: curve,
		X:     new(big.Int).SetBytes(b[:fieldSize]),
		Y:     new(big.Int).SetBytes(b[fieldSize:]),
	}
	if !p.IsValid() {
		return nil, fmt.Errorf("invalid Point")
	}
	return p, nil
}

// sameCurve determines if points a,b appear to be from the same curve
func sameCurve(a, b *EcPoint) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aParams := a.Curve.Params()
	bParams := b.Curve.Params()
	return aParams.P.Cmp(bParams.P) == 0 &&
		aParams.N.Cmp(bParams.N) == 0 &&
		aParams.B.Cmp(bParams.B) == 0 &&
		aParams.BitSize == bParams.BitSize &&
		aParams.Gx.Cmp(bParams.Gx) == 0 &&
		aParams.Gy.Cmp(bParams.Gy) == 0 &&
		aParams.Name == bParams.Name
}
