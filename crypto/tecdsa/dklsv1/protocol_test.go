//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package dklsv1

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/sonr-io/sonr/crypto/core/curves"
	"github.com/sonr-io/sonr/crypto/core/protocol"
	"github.com/sonr-io/sonr/crypto/ot/extension/kos"
	"github.com/sonr-io/sonr/crypto/tecdsa/dklsv1/dkg"
)

// For DKG bob starts first. For refresh and sign, Alice starts first.
func runIteratedProtocol(
	firstParty protocol.Iterator,
	secondParty protocol.Iterator,
) (error, error) {
	var (
		message *protocol.Message
		aErr    error
		bErr    error
	)

	for aErr != protocol.ErrProtocolFinished || bErr != protocol.ErrProtocolFinished {
		// Crank each protocol forward one iteration
		message, bErr = firstParty.Next(message)
		if bErr != nil && bErr != protocol.ErrProtocolFinished {
			return nil, bErr
		}

		message, aErr = secondParty.Next(message)
		if aErr != nil && aErr != protocol.ErrProtocolFinished {
			return aErr, nil
		}
	}
	return aErr, bErr
}

// Running steps in sequence ensures that no hidden read/write dependency exist in the read/write interfaces.
func TestDkgProto(t *testing.T) {
	curveInstances := []*curves.Curve{
		curves.K256(),
		curves.P256(),
	}
	for _, curve := range curveInstances {
		alice := NewAliceDkg(curve, protocol.Version1)
		bob := NewBobDkg(curve, protocol.Version1)
		aErr, bErr := runIteratedProtocol(bob, alice)

		t.Run("both alice/bob complete simultaneously", func(t *testing.T) {
			require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
			require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
		})

		for i := 0; i < kos.Kappa; i++ {
			if alice.Alice.Output().SeedOtResult.OneTimePadDecryptionKey[i] != bob.Bob.Output().SeedOtResult.OneTimePadEncryptionKeys[i][alice.Alice.Output().SeedOtResult.RandomChoiceBits[i]] {
				t.Errorf("oblivious transfer is incorrect at index i=%v", i)
			}
		}

		t.Run("Both parties produces identical composite pubkey", func(t *testing.T) {
			require.True(t, alice.Alice.Output().PublicKey.Equal(bob.Bob.Output().PublicKey))
		})

		var aliceResult *dkg.AliceOutput
		var bobResult *dkg.BobOutput
		t.Run("alice produces valid result", func(t *testing.T) {
			// Get the result
			r, err := alice.Result(protocol.Version1)

			// Test
			require.NoError(t, err)
			require.NotNil(t, r)
			aliceResult, err = DecodeAliceDkgResult(r)
			require.NoError(t, err)
		})
		t.Run("bob produces valid result", func(t *testing.T) {
			// Get the result
			r, err := bob.Result(protocol.Version1)

			// Test
			require.NoError(t, err)
			require.NotNil(t, r)
			bobResult, err = DecodeBobDkgResult(r)
			require.NoError(t, err)
		})

		t.Run("alice/bob agree on pubkey", func(t *testing.T) {
			require.Equal(t, aliceResult.PublicKey, bobResult.PublicKey)
		})
	}
}

//
// // DKG > Refresh > Sign
// func TestRefreshProto(t *testing.T) {
// 	t.Parallel()
// 	curveInstances := []*curves.Curve{
// 		curves.K256(),
// 		curves.P256(),
// 	}
// 	for _, curve := range curveInstances {
// 		boundCurve := curve
// 		t.Run(fmt.Sprintf("testing refresh for curve %s", boundCurve.Name), func(tt *testing.T) {
// 			tt.Parallel()
// 			// DKG
// 			aliceDkg := NewAliceDkg(boundCurve, protocol.Version1)
// 			bobDkg := NewBobDkg(boundCurve, protocol.Version1)
//
// 			aDkgErr, bDkgErr := runIteratedProtocol(bobDkg, aliceDkg)
// 			require.ErrorIs(tt, aDkgErr, protocol.ErrProtocolFinished)
// 			require.ErrorIs(tt, bDkgErr, protocol.ErrProtocolFinished)
//
// 			aliceDkgResultMessage, err := aliceDkg.Result(protocol.Version1)
// 			require.NoError(tt, err)
// 			bobDkgResultMessage, err := bobDkg.Result(protocol.Version1)
// 			require.NoError(tt, err)
//
// 			// Refresh
// 			aliceRefreshResultMessage, bobRefreshResultMessage := refreshV1(t, boundCurve, aliceDkgResultMessage, bobDkgResultMessage)
//
// 			// sign
// 			signV1(t, boundCurve, aliceRefreshResultMessage, bobRefreshResultMessage)
// 		})
// 	}
// }

// DKG > Output > NewDklsSign > Sign > Output
func TestDkgSignProto(t *testing.T) {
	// Setup
	curve := curves.K256()

	aliceDkg := NewAliceDkg(curve, protocol.Version1)
	bobDkg := NewBobDkg(curve, protocol.Version1)

	// DKG
	aErr, bErr := runIteratedProtocol(bobDkg, aliceDkg)
	require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
	require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)

	// Output
	aliceDkgResultMessage, err := aliceDkg.Result(protocol.Version1)
	require.NoError(t, err)

	bobDkgResultMessage, err := bobDkg.Result(protocol.Version1)
	require.NoError(t, err)

	// New DklsSign
	msg := []byte("As soon as you trust yourself, you will know how to live.")
	aliceSign, err := NewAliceSign(
		curve,
		sha3.New256(),
		msg,
		aliceDkgResultMessage,
		protocol.Version1,
	)
	require.NoError(t, err)
	bobSign, err := NewBobSign(curve, sha3.New256(), msg, bobDkgResultMessage, protocol.Version1)
	require.NoError(t, err)

	// Sign
	t.Run("sign", func(t *testing.T) {
		aErr, bErr = runIteratedProtocol(aliceSign, bobSign)
		require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
		require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
	})
	// Don't continue to verifying results if sign didn't run to completion.
	require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
	require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)

	// Output
	var result *curves.EcdsaSignature
	t.Run("bob produces result of correct type", func(t *testing.T) {
		resultMessage, err := bobSign.Result(protocol.Version1)
		require.NoError(t, err)
		result, err = DecodeSignature(resultMessage)
		require.NoError(t, err)
	})
	require.NotNil(t, result)

	t.Run("valid signature", func(t *testing.T) {
		hash := sha3.New256()
		_, err = hash.Write(msg)
		require.NoError(t, err)
		digest := hash.Sum(nil)
		unCompressedAffinePublicKey := aliceDkg.Output().PublicKey.ToAffineUncompressed()
		require.Equal(t, 65, len(unCompressedAffinePublicKey))
		x := new(big.Int).SetBytes(unCompressedAffinePublicKey[1:33])
		y := new(big.Int).SetBytes(unCompressedAffinePublicKey[33:])
		ecCurve, err := curve.ToEllipticCurve()
		require.NoError(t, err)
		publicKey := &curves.EcPoint{
			Curve: ecCurve,
			X:     x,
			Y:     y,
		}
		require.True(t,
			curves.VerifyEcdsa(publicKey,
				digest[:],
				result,
			),
			"signature failed verification",
		)
	})
}

//
// // Decode > NewDklsSign > Sign > Output
// // NOTE: this cold-start test ensures backwards compatibility with durable,
// // encoding DKG state that may exist within production systems like test
// // Breaking changes must consider downstream production impacts.
// func TestSignColdStart(t *testing.T) {
// 	// Decode alice/bob state from file
// 	aliceDkg, err := os.ReadFile("testdata/alice-dkls-v1-dkg.bin")
// 	require.NoError(t, err)
// 	bobDkg, err := os.ReadFile("testdata/bob-dkls-v1-dkg.bin")
// 	require.NoError(t, err)
//
// 	// The choice of json marshaling is arbitrary, the binary could have been marshaled in other forms as well
// 	// The purpose here is to obtain an instance of `protocol.Message`
//
// 	aliceDkgMessage := &protocol.Message{}
// 	err = json.Unmarshal(aliceDkg, aliceDkgMessage)
// 	require.NoError(t, err)
//
// 	bobDkgMessage := &protocol.Message{}
// 	err = json.Unmarshal(bobDkg, bobDkgMessage)
// 	require.NoError(t, err)
//
// 	signV1(t, curves.K256(), aliceDkgMessage, bobDkgMessage)
// }
//
// func TestEncodeDecode(t *testing.T) {
// 	curve := curves.K256()
//
// 	alice := NewAliceDkg(curve, protocol.Version1)
// 	bob := NewBobDkg(curve, protocol.Version1)
// 	_, _ = runIteratedProtocol(bob, alice)
//
// 	var aliceBytes []byte
// 	var bobBytes []byte
// 	t.Run("Encode Alice/Bob", func(t *testing.T) {
// 		aliceDkgMessage, err := EncodeAliceDkgOutput(alice.Alice.Output(), protocol.Version1)
// 		require.NoError(t, err)
// 		aliceBytes, err = json.Marshal(aliceDkgMessage)
// 		require.NoError(t, err)
//
// 		bobDkgMessage, err := EncodeBobDkgOutput(bob.Bob.Output(), protocol.Version1)
// 		require.NoError(t, err)
// 		bobBytes, err = json.Marshal(bobDkgMessage)
// 		require.NoError(t, err)
// 	})
// 	require.NotEmpty(t, aliceBytes)
// 	require.NotEmpty(t, bobBytes)
//
// 	t.Run("Decode Alice", func(t *testing.T) {
// 		decodedAliceMessage := &protocol.Message{}
// 		err := json.Unmarshal(aliceBytes, decodedAliceMessage)
// 		require.NoError(t, err)
// 		require.NotNil(t, decodedAliceMessage)
// 		decodedAlice, err := DecodeAliceDkgResult(decodedAliceMessage)
// 		require.NoError(t, err)
//
// 		require.True(t, alice.Output().PublicKey.Equal(decodedAlice.PublicKey))
// 		require.Equal(t, alice.Output().SecretKeyShare, decodedAlice.SecretKeyShare)
// 	})
//
// 	t.Run("Decode Bob", func(t *testing.T) {
// 		decodedBobMessage := &protocol.Message{}
// 		err := json.Unmarshal(bobBytes, decodedBobMessage)
// 		require.NoError(t, err)
// 		require.NotNil(t, decodedBobMessage)
// 		decodedBob, err := DecodeBobDkgResult(decodedBobMessage)
// 		require.NoError(t, err)
//
// 		require.True(t, bob.Output().PublicKey.Equal(decodedBob.PublicKey))
// 		require.Equal(t, bob.Output().SecretKeyShare, decodedBob.SecretKeyShare)
// 	})
// }
//
// func signV1(t *testing.T, curve *curves.Curve, aliceDkgResultMessage *protocol.Message, bobDkgResultMessage *protocol.Message) {
// 	t.Helper()
// 	// New DklsSign
// 	msg := []byte("As soon as you trust yourself, you will know how to live.")
// 	aliceSign, err := NewAliceSign(curve, sha3.New256(), msg, aliceDkgResultMessage, protocol.Version1)
// 	require.NoError(t, err)
// 	bobSign, err := NewBobSign(curve, sha3.New256(), msg, bobDkgResultMessage, protocol.Version1)
// 	require.NoError(t, err)
//
// 	// Sign
// 	var aErr error
// 	var bErr error
// 	t.Run("sign", func(t *testing.T) {
// 		aErr, bErr = runIteratedProtocol(aliceSign, bobSign)
// 		require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
// 		require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
// 	})
// 	// Don't continue to verifying results if sign didn't run to completion.
// 	require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
// 	require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
//
// 	// Output
// 	var result *curves.EcdsaSignature
// 	t.Run("bob produces result of correct type", func(t *testing.T) {
// 		resultMessage, err := bobSign.Result(protocol.Version1)
// 		require.NoError(t, err)
// 		result, err = DecodeSignature(resultMessage)
// 		require.NoError(t, err)
// 	})
// 	require.NotNil(t, result)
//
// 	aliceDkg, err := DecodeAliceDkgResult(aliceDkgResultMessage)
// 	require.NoError(t, err)
//
// 	t.Run("valid signature", func(t *testing.T) {
// 		hash := sha3.New256()
// 		_, err = hash.Write(msg)
// 		require.NoError(t, err)
// 		digest := hash.Sum(nil)
// 		unCompressedAffinePublicKey := aliceDkg.PublicKey.ToAffineUncompressed()
// 		require.Equal(t, 65, len(unCompressedAffinePublicKey))
// 		x := new(big.Int).SetBytes(unCompressedAffinePublicKey[1:33])
// 		y := new(big.Int).SetBytes(unCompressedAffinePublicKey[33:])
// 		ecCurve, err := curve.ToEllipticCurve()
// 		require.NoError(t, err)
// 		publicKey := &curves.EcPoint{
// 			Curve: ecCurve,
// 			X:     x,
// 			Y:     y,
// 		}
// 		require.True(t,
// 			curves.VerifyEcdsa(publicKey,
// 				digest[:],
// 				result,
// 			),
// 			"signature failed verification",
// 		)
// 	})
// }
//
// func refreshV1(t *testing.T, curve *curves.Curve, aliceDkgResultMessage, bobDkgResultMessage *protocol.Message) (aliceRefreshResultMessage, bobRefreshResultMessage *protocol.Message) {
// 	t.Helper()
// 	aliceRefresh, err := NewAliceRefresh(curve, aliceDkgResultMessage, protocol.Version1)
// 	require.NoError(t, err)
// 	bobRefresh, err := NewBobRefresh(curve, bobDkgResultMessage, protocol.Version1)
// 	require.NoError(t, err)
//
// 	aErr, bErr := runIteratedProtocol(aliceRefresh, bobRefresh)
// 	require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
// 	require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
//
// 	aliceRefreshResultMessage, err = aliceRefresh.Result(protocol.Version1)
// 	require.NoError(t, err)
// 	require.NotNil(t, aliceRefreshResultMessage)
// 	_, err = DecodeAliceRefreshResult(aliceRefreshResultMessage)
// 	require.NoError(t, err)
//
// 	bobRefreshResultMessage, err = bobRefresh.Result(protocol.Version1)
// 	require.NoError(t, err)
// 	require.NotNil(t, bobRefreshResultMessage)
// 	_, err = DecodeBobRefreshResult(bobRefreshResultMessage)
// 	require.NoError(t, err)
//
// 	return aliceRefreshResultMessage, bobRefreshResultMessage
// }
//
// func BenchmarkDKGProto(b *testing.B) {
// 	curveInstances := []*curves.Curve{
// 		curves.K256(),
// 		curves.P256(),
// 	}
// 	for _, curve := range curveInstances {
// 		alice := NewAliceDkg(curve, protocol.Version1)
// 		bob := NewBobDkg(curve, protocol.Version1)
// 		aErr, bErr := runIteratedProtocol(bob, alice)
//
// 		b.Run("both alice/bob complete simultaneously", func(t *testing.B) {
// 			require.ErrorIs(t, aErr, protocol.ErrProtocolFinished)
// 			require.ErrorIs(t, bErr, protocol.ErrProtocolFinished)
// 		})
//
// 		for i := 0; i < kos.Kappa; i++ {
// 			if alice.Alice.Output().SeedOtResult.OneTimePadDecryptionKey[i] != bob.Bob.Output().SeedOtResult.OneTimePadEncryptionKeys[i][alice.Alice.Output().SeedOtResult.RandomChoiceBits[i]] {
// 				b.Errorf("oblivious transfer is incorrect at index i=%v", i)
// 			}
// 		}
//
// 		b.Run("Both parties produces identical composite pubkey", func(t *testing.B) {
// 			require.True(t, alice.Alice.Output().PublicKey.Equal(bob.Bob.Output().PublicKey))
// 		})
//
// 		var aliceResult *dkg.AliceOutput
// 		var bobResult *dkg.BobOutput
// 		b.Run("alice produces valid result", func(t *testing.B) {
// 			// Get the result
// 			r, err := alice.Result(protocol.Version1)
//
// 			// Test
// 			require.NoError(t, err)
// 			require.NotNil(t, r)
// 			aliceResult, err = DecodeAliceDkgResult(r)
// 			require.NoError(t, err)
// 		})
// 		b.Run("bob produces valid result", func(t *testing.B) {
// 			// Get the result
// 			r, err := bob.Result(protocol.Version1)
//
// 			// Test
// 			require.NoError(t, err)
// 			require.NotNil(t, r)
// 			bobResult, err = DecodeBobDkgResult(r)
// 			require.NoError(t, err)
// 		})
//
// 		b.Run("alice/bob agree on pubkey", func(t *testing.B) {
// 			require.Equal(t, aliceResult.PublicKey, bobResult.PublicKey)
// 		})
// 	}
// }
