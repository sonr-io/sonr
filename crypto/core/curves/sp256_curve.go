package curves

import (
	"crypto/elliptic"

	"github.com/dustinxie/ecc"
)

func SP256() elliptic.Curve {
	return ecc.P256k1()
}
