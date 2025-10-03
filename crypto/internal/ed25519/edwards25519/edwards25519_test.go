// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edwards25519

import (
	"crypto/rand"
	"testing"
)

func TestScNegMulAddIdentity(t *testing.T) {
	one := [32]byte{1}
	var seed [64]byte
	var a, minusA, x [32]byte

	rand.Reader.Read(seed[:])
	ScReduce(&a, &seed)
	copy(minusA[:], a[:])
	ScNeg(&minusA, &minusA)
	ScMulAdd(&x, &one, &a, &minusA)
	if x != [32]byte{} {
		t.Errorf("%x --neg--> %x --sum--> %x", a, minusA, x)
	}
}

func TestGeAddAgainstgeAdd(t *testing.T) {
	var x, y [32]byte
	rand.Reader.Read(x[:])
	rand.Reader.Read(y[:])
	var X, Y ExtendedGroupElement
	x[31] &= 127
	y[31] &= 127
	GeScalarMultBase(&X, &x)
	GeScalarMultBase(&Y, &y)

	var SBytesRef [32]byte
	var SRef ExtendedGroupElement
	var Ycached CachedGroupElement
	var SCompletedRef CompletedGroupElement
	Y.ToCached(&Ycached)
	geAdd(&SCompletedRef, &X, &Ycached)
	SCompletedRef.ToExtended(&SRef)
	SRef.ToBytes(&SBytesRef)

	var SBytes [32]byte
	var S ExtendedGroupElement
	GeAdd(&S, &X, &Y)
	S.ToBytes(&SBytes)

	if SBytes != SBytesRef {
		t.Errorf("GeAdd does not match geAdd: %x != %x", SBytes, SBytesRef)
	}
}

func TestGeScalarMultAgainstDoubleScalarMult(t *testing.T) {
	var zero [32]byte

	var x [32]byte
	rand.Reader.Read(x[:])
	var X ExtendedGroupElement
	x[31] &= 127
	GeScalarMultBase(&X, &x)

	var a [32]byte
	rand.Reader.Read(a[:])
	a[31] &= 127

	var ABytesRef [32]byte
	var Aref ProjectiveGroupElement
	GeDoubleScalarMultVartime(&Aref, &a, &X, &zero)
	Aref.ToBytes(&ABytesRef)

	var ABytes [32]byte
	var A ExtendedGroupElement
	GeScalarMult(&A, &a, &X)
	A.ToBytes(&ABytes)

	if ABytes != ABytesRef {
		t.Errorf(
			"GeScalarMult does not match GeDoubleScalarMultVartime: %x != %x",
			ABytes,
			ABytesRef,
		)
	}
}

func inc(b *[32]byte) {
	acc := uint(1)
	for i := 0; i < 32; i++ {
		acc += uint(b[i])
		acc, b[i] = uint(acc>>8), byte(acc)
	}
}

func TestGeAddAgainstScalarMult(t *testing.T) {
	var x, s [32]byte
	rand.Reader.Read(x[:])
	x[31] &= 127
	var X ExtendedGroupElement
	GeScalarMultBase(&X, &x)
	var A ExtendedGroupElement
	A.Zero()
	var ABytes, ABytesMult [32]byte
	A.ToBytes(&ABytes)
	var AMult ExtendedGroupElement
	GeScalarMult(&AMult, &s, &X)
	AMult.ToBytes(&ABytesMult)
	if ABytes != ABytesMult {
		t.Errorf("addition does not match multiplication (%x) -> %x != %x", s, ABytes, ABytesMult)
	}

	GeAdd(&A, &A, &X)
	inc(&s)
}

func TestGeAddAgainstDoubleScalarMult(t *testing.T) {
	var zero, x, s [32]byte
	rand.Reader.Read(x[:])
	x[31] &= 127
	var X ExtendedGroupElement
	GeScalarMultBase(&X, &x)
	var A ExtendedGroupElement
	A.Zero()
	var ABytes, ABytesMult [32]byte
	A.ToBytes(&ABytes)
	var AMult ProjectiveGroupElement
	GeDoubleScalarMultVartime(&AMult, &s, &X, &zero)
	AMult.ToBytes(&ABytesMult)
	if ABytes != ABytesMult {
		t.Errorf("addition does not match multiplication (%x)", s)
	}

	GeAdd(&A, &A, &X)
	inc(&s)
}

func TestGeScalarMultiBaseScalarMultDH(t *testing.T) {
	var a, b [32]byte
	rand.Reader.Read(a[:])
	rand.Reader.Read(b[:])

	var A, B ExtendedGroupElement
	a[31] &= 127
	b[31] &= 127
	GeScalarMultBase(&A, &a)
	GeScalarMultBase(&B, &b)

	var kA, kB ExtendedGroupElement
	GeScalarMult(&kA, &a, &B)
	GeScalarMult(&kB, &b, &A)

	var ka, kb [32]byte
	kA.ToBytes(&ka)
	kB.ToBytes(&kb)

	if ka != kb {
		t.Fatal("DH shared secrets do not match")
	}
}

func TestGeScalarMultDH(t *testing.T) {
	var one [32]byte
	one[0] = 1
	var base ExtendedGroupElement
	GeScalarMultBase(&base, &one)

	var a, b [32]byte
	rand.Reader.Read(a[:])
	rand.Reader.Read(b[:])

	var A, B ExtendedGroupElement
	GeScalarMult(&A, &a, &base)
	GeScalarMult(&B, &b, &base)

	var kA, kB ExtendedGroupElement
	GeScalarMult(&kA, &a, &B)
	GeScalarMult(&kB, &b, &A)

	var ka, kb [32]byte
	kA.ToBytes(&ka)
	kB.ToBytes(&kb)

	if ka != kb {
		t.Fatal("DH shared secrets do not match")
	}

	var bytesA, bytesB [32]byte
	A.ToBytes(&bytesA)
	B.ToBytes(&bytesB)

	if bytesA == bytesB {
		t.Fatalf("DH public key collision: g^%x = g^%x = %x", a, b, A)
	}
}

func BenchmarkGeScalarMultBase(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])
	var P ExtendedGroupElement

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeScalarMultBase(&P, &s)
	}
}

func BenchmarkGeScalarMult(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])

	var P ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeScalarMult(&P, &s, &P)
	}
}

func BenchmarkGeDoubleScalarMultVartime(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])

	var P, Pout ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)

	var out ProjectiveGroupElement

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeDoubleScalarMultVartime(&out, &s, &P, &[32]byte{})
		out.ToExtended(&Pout)
	}
}

func BenchmarkGeAdd(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])

	var R, P ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)
	R = P

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeAdd(&R, &R, &P)
	}
}

func BenchmarkGeDouble(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])

	var R, P ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)
	R = P

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeDouble(&R, &P)
	}
}
