module vault

go 1.24.7

replace github.com/sonr-io/sonr/crypto => ../../crypto/

require (
	github.com/extism/go-pdk v1.1.3
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/sonr-io/sonr/crypto v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bits-and-blooms/bitset v1.24.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/bwesterb/go-ristretto v1.2.3 // indirect
	github.com/consensys/gnark-crypto v0.19.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/dustinxie/ecc v0.0.0-20210511000915-959544187564 // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20181016162300-f8f6d4d2b643 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)
