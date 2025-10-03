package types

import (
	"encoding/hex"
	"slices"
	"strings"

	"lukechampine.com/blake3"
)

var SupportedDIDAssertionMethods = []string{
	"sonr",
	"btcr",
	"ethr",
	"ssh",
	"tel",
	"email",
	"github",
	"google",
}

func IsSupportedDIDAssertionMethod(method string) bool {
	return slices.Contains(SupportedDIDAssertionMethods, method)
}

// HashAssertionValue hashes an assertion value using blake3
func HashAssertionValue(value string) string {
	hash := blake3.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

type DIDAssertionMethod string

func (m DIDAssertionMethod) Parse() error {
	return nil
}

func (m DIDAssertionMethod) String() string {
	return string(m)
}

func TrimDIDMethodPrefix(did string) string {
	if after, ok := strings.CutPrefix(did, "did:"); ok {
		return after
	}
	return did
}
