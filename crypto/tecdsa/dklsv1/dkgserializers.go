package dklsv1

import (
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"

	"github.com/sonr-io/sonr/crypto/core/curves"
	"github.com/sonr-io/sonr/crypto/core/protocol"
	"github.com/sonr-io/sonr/crypto/ot/base/simplest"
	"github.com/sonr-io/sonr/crypto/tecdsa/dklsv1/dkg"
	"github.com/sonr-io/sonr/crypto/zkp/schnorr"
)

const payloadKey = "direct"

func newDkgProtocolMessage(payload []byte, round string, version uint) *protocol.Message {
	return &protocol.Message{
		Protocol: protocol.Dkls18Dkg,
		Version:  version,
		Payloads: map[string][]byte{payloadKey: payload},
		Metadata: map[string]string{"round": round},
	}
}

func registerTypes() {
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})
	gob.Register(&curves.ScalarP256{})
	gob.Register(&curves.PointP256{})
}

func encodeDkgRound1Output(commitment [32]byte, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&commitment); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "1", version), nil
}

func decodeDkgRound2Input(m *protocol.Message) ([32]byte, error) {
	if m.Version != protocol.Version1 {
		return [32]byte{}, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := [32]byte{}
	if err := dec.Decode(&decoded); err != nil {
		return [32]byte{}, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound2Output(output *dkg.Round2Output, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "2", version), nil
}

func decodeDkgRound3Input(m *protocol.Message) (*dkg.Round2Output, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.Round2Output)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound3Output(proof *schnorr.Proof, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "3", version), nil
}

func decodeDkgRound4Input(m *protocol.Message) (*schnorr.Proof, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound4Output(proof *schnorr.Proof, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "4", version), nil
}

func decodeDkgRound5Input(m *protocol.Message) (*schnorr.Proof, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound5Output(proof *schnorr.Proof, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "5", version), nil
}

func decodeDkgRound6Input(m *protocol.Message) (*schnorr.Proof, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound6Output(
	choices []simplest.ReceiversMaskedChoices,
	version uint,
) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(choices); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "6", version), nil
}

func decodeDkgRound7Input(m *protocol.Message) ([]simplest.ReceiversMaskedChoices, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ReceiversMaskedChoices{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound7Output(
	challenge []simplest.OtChallenge,
	version uint,
) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(challenge); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "7", version), nil
}

func decodeDkgRound8Input(m *protocol.Message) ([]simplest.OtChallenge, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallenge{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound8Output(
	responses []simplest.OtChallengeResponse,
	version uint,
) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(responses); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "8", version), nil
}

func decodeDkgRound9Input(m *protocol.Message) ([]simplest.OtChallengeResponse, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallengeResponse{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func encodeDkgRound9Output(
	opening []simplest.ChallengeOpening,
	version uint,
) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(opening); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "9", version), nil
}

func decodeDkgRound10Input(m *protocol.Message) ([]simplest.ChallengeOpening, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ChallengeOpening{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

// EncodeAliceDkgOutput serializes Alice DKG output based on the protocol version.
func EncodeAliceDkgOutput(result *dkg.AliceOutput, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "alice-output", version), nil
}

// DecodeAliceDkgResult deserializes Alice DKG output.
func DecodeAliceDkgResult(m *protocol.Message) (*dkg.AliceOutput, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	registerTypes()
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.AliceOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

// EncodeBobDkgOutput serializes Bob DKG output based on the protocol version.
func EncodeBobDkgOutput(result *dkg.BobOutput, version uint) (*protocol.Message, error) {
	if version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgProtocolMessage(buf.Bytes(), "bob-output", version), nil
}

// DecodeBobDkgResult deserializes Bob DKG output.
func DecodeBobDkgResult(m *protocol.Message) (*dkg.BobOutput, error) {
	if m.Version != protocol.Version1 {
		return nil, errors.New("only version 1 is supported")
	}
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.BobOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}
