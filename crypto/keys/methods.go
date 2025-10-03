package keys

type DIDMethod string

const (
	DIDMethodKey      DIDMethod = "key"
	DIDMethodSonr     DIDMethod = "sonr"
	DIDMehthodBitcoin DIDMethod = "btcr"
	DIDMethodEthereum DIDMethod = "ethr"
	DIDMethodCbor     DIDMethod = "cbor"
	DIDMethodCID      DIDMethod = "cid"
	DIDMethodIPFS     DIDMethod = "ipfs"
)

func (d DIDMethod) String() string {
	return string(d)
}
