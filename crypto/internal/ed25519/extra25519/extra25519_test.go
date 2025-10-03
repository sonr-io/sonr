// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package extra25519

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"testing"

	"github.com/sonr-io/sonr/crypto/internal/ed25519/edwards25519"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
)

func TestCurve25519Conversion(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	var pubBytes [32]byte
	copy(pubBytes[:], public)
	var privBytes [64]byte
	copy(privBytes[:], private)

	var curve25519Public, curve25519Public2, curve25519Private [32]byte
	PrivateKeyToCurve25519(&curve25519Private, &privBytes)
	curve25519.ScalarBaseMult(&curve25519Public, &curve25519Private)

	if !PublicKeyToCurve25519(&curve25519Public2, &pubBytes) {
		t.Fatalf("PublicKeyToCurve25519 failed")
	}

	if !bytes.Equal(curve25519Public[:], curve25519Public2[:]) {
		t.Errorf(
			"Values didn't match: curve25519 produced %x, conversion produced %x",
			curve25519Public[:],
			curve25519Public2[:],
		)
	}
}

func TestHashNoCollisions(t *testing.T) {
	type intpair struct {
		i int
		j uint
	}
	rainbow := make(map[[32]byte]intpair)
	N := 25
	if testing.Short() {
		N = 3
	}
	var h [32]byte
	// NOTE: hash values 0b100000000000... and 0b00000000000... both map to
	// the identity. this is a core part of the elligator function and not a
	// collision we need to worry about because an attacker would need to find
	// the preimages of these hashes to exploit it.
	h[0] = 1
	for i := 0; i < N; i++ {
		for j := range uint(257) {
			if j < 256 {
				h[j>>3] ^= byte(1) << (j & 7)
			}

			var P edwards25519.ExtendedGroupElement
			HashToEdwards(&P, &h)
			var p [32]byte
			P.ToBytes(&p)
			if c, ok := rainbow[p]; ok {
				t.Fatalf("found collision: (%d, %d) and (%d, %d)", i, j, c.i, c.j)
			}
			rainbow[p] = intpair{i, j}

			if j < 256 {
				h[j>>3] ^= byte(1) << (j & 7)
			}
		}
		hh := sha512.Sum512(h[:]) // this package already imports sha512
		copy(h[:], hh[:])
	}
}

func TestElligator(t *testing.T) {
	var publicKey, publicKey2, publicKey3, representative, privateKey [32]byte

	for range 1000 {
		rand.Reader.Read(privateKey[:])

		if !ScalarBaseMult(&publicKey, &representative, &privateKey) {
			continue
		}
		RepresentativeToPublicKey(&publicKey2, &representative)
		if !bytes.Equal(publicKey[:], publicKey2[:]) {
			t.Fatal("The resulting public key doesn't match the initial one.")
		}

		curve25519.ScalarBaseMult(&publicKey3, &privateKey)
		if !bytes.Equal(publicKey[:], publicKey3[:]) {
			t.Fatal("The public key doesn't match the value that curve25519 produced.")
		}
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	var publicKey, representative, privateKey [32]byte

	// Find the private key that results in a point that's in the image of the map.
	for {
		rand.Reader.Read(privateKey[:])
		if ScalarBaseMult(&publicKey, &representative, &privateKey) {
			break
		}
	}

	for b.Loop() {
		ScalarBaseMult(&publicKey, &representative, &privateKey)
	}
}

func BenchmarkMap(b *testing.B) {
	var publicKey, representative [32]byte
	rand.Reader.Read(representative[:])

	for b.Loop() {
		RepresentativeToPublicKey(&publicKey, &representative)
	}
}
