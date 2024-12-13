package mpc

import (
	"fmt"
	"strings"

	"github.com/onsonr/sonr/crypto/core/curves"
	"github.com/onsonr/sonr/crypto/core/protocol"
	"github.com/onsonr/sonr/crypto/tecdsa/dklsv1/dkg"
)

// ╭───────────────────────────────────────────────────────────╮
// │                    Exported Generics                      │
// ╰───────────────────────────────────────────────────────────╯

type (
	AliceOut    *dkg.AliceOutput
	BobOut      *dkg.BobOutput
	Point       curves.Point
	Role        string                         // Role is the type for the role
	Message     *protocol.Message              // Message is the protocol.Message that is used for MPC
	Signature   *curves.EcdsaSignature         // Signature is the type for the signature
	RefreshFunc interface{ protocol.Iterator } // RefreshFunc is the type for the refresh function
	SignFunc    interface{ protocol.Iterator } // SignFunc is the type for the sign function
)

// ╭───────────────────────────────────────────────────────────╮
// │                      Keyshare Encoding                    │
// ╰───────────────────────────────────────────────────────────╯

type KeyShare string

func (k KeyShare) AliceOut() (AliceOut, error) {
	if k.Role() != RoleValidator {
		return nil, fmt.Errorf("invalid share role")
	}
	msg, err := k.Message()
	if err != nil {
		return nil, err
	}
	return getAliceOut(msg)
}

func (k KeyShare) BobOut() (BobOut, error) {
	if k.Role() != RoleUser {
		return nil, fmt.Errorf("invalid share role")
	}
	msg, err := k.Message()
	if err != nil {
		return nil, err
	}
	bobOut, err := getBobOut(msg)
	if err != nil {
		return nil, err
	}
	return bobOut, nil
}

func (k KeyShare) Message() (*protocol.Message, error) {
	r := k.Role()
	return r.Parse(k)
}

func (k KeyShare) Role() Role {
	parts := strings.Split(k.String(), ".")
	return determineRole(parts[0])
}

func (k KeyShare) String() string {
	return string(k)
}

func EncodeKeyshare(m Message, role Role) (KeyShare, error) {
	enc, err := protocol.EncodeMessage(m)
	if err != nil {
		return "", err
	}
	return role.Format(enc), nil
}

func DecodeKeyshare(s string) (KeyShare, error) {
	role := determineRole(s)
	if role == RoleUnknown {
		return "", fmt.Errorf("invalid share role")
	}
	return role.Format(s), nil
}

// ╭───────────────────────────────────────────────────────────╮
// │                  MPC Share Roles (Alice/Bob)              │
// ╰───────────────────────────────────────────────────────────╯

const (
	RoleUser      Role = "user."
	RoleValidator Role = "validator."
	RoleUnknown   Role = ""
)

func determineRole(s string) Role {
	if strings.Contains(s, "user") {
		return RoleUser
	}
	if strings.Contains(s, "validator") {
		return RoleValidator
	}
	return RoleUnknown
}

func (r Role) String() string {
	return string(r)
}

func (r Role) Format(encodedKeyShare string) KeyShare {
	return KeyShare(fmt.Sprintf("%s%s", r.String(), encodedKeyShare))
}

func (r Role) Parse(share KeyShare) (Message, error) {
	if r == RoleUnknown {
		return nil, fmt.Errorf("invalid share role")
	}
	parts := strings.Split(share.String(), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid share format")
	}
	if parts[0] != r.String() {
		return nil, fmt.Errorf("invalid share role")
	}
	return protocol.DecodeMessage(parts[1])
}

func (r Role) IsUser() bool      { return r == RoleUser }
func (r Role) IsValidator() bool { return r == RoleValidator }