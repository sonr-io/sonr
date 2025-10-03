package vrf

import (
	"bytes"
	"testing"
	// "fmt"
)

func TestHonestComplete(t *testing.T) {
	sk, err := GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pk, _ := sk.Public()
	alice := []byte("alice")
	aliceVRF := sk.Compute(alice)
	aliceVRFFromProof, aliceProof := sk.Prove(alice)

	// fmt.Printf("pk:           %X\n", pk)
	// fmt.Printf("sk:           %X\n", *sk)
	// fmt.Printf("alice(bytes): %X\n", alice)
	// fmt.Printf("aliceVRF:     %X\n", aliceVRF)
	// fmt.Printf("aliceProof:   %X\n", aliceProof)

	if !pk.Verify(alice, aliceVRF, aliceProof) {
		t.Error("Gen -> Compute -> Prove -> Verify -> FALSE")
	}
	if !bytes.Equal(aliceVRF, aliceVRFFromProof) {
		t.Error("Compute != Prove")
	}
}

func TestConvertPrivateKeyToPublicKey(t *testing.T) {
	sk, err := GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	pk, ok := sk.Public()
	if !ok {
		t.Fatal("Couldn't obtain public key.")
	}
	if !bytes.Equal(sk[32:], pk) {
		t.Fatal("Raw byte respresentation doesn't match public key.")
	}
}

func TestFlipBitForgery(t *testing.T) {
	sk, err := GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pk, _ := sk.Public()
	alice := []byte("alice")
	for i := 0; i < 32; i++ {
		for j := uint(0); j < 8; j++ {
			aliceVRF := sk.Compute(alice)
			aliceVRF[i] ^= 1 << j
			_, aliceProof := sk.Prove(alice)
			if pk.Verify(alice, aliceVRF, aliceProof) {
				t.Fatalf("forged by using aliceVRF[%d]^=%d:\n (sk=%x)", i, j, sk)
			}
		}
	}
}

func BenchmarkHashToGE(b *testing.B) {
	alice := []byte("alice")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		hashToCurve(alice)
	}
}

func BenchmarkCompute(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sk.Compute(alice)
	}
}

func BenchmarkProve(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sk.Prove(alice)
	}
}

func BenchmarkVerify(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	aliceVRF := sk.Compute(alice)
	_, aliceProof := sk.Prove(alice)
	pk, _ := sk.Public()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pk.Verify(alice, aliceVRF, aliceProof)
	}
}
