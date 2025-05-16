package embed

import (
	"encoding/json"

	"github.com/ipfs/boxo/files"
	motr "github.com/onsonr/motr/pkg/config"
)

const SchemaVersion = 1
const (
	AppManifestFileName   = "app.webmanifest"
	DWNConfigFileName     = "dwn.json"
	IndexHTMLFileName     = "index.html"
	MainJSFileName        = "main.js"
	ServiceWorkerFileName = "sw.js"
)

// NewVaultFS creates a new directory with the default files
func NewVaultFS(cfg *motr.Config) (files.Directory, error) {
	manifestBz, err := NewWebManifest()
	if err != nil {
		return nil, err
	}
	cnfBz, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return files.NewMapDirectory(map[string]files.Node{
		AppManifestFileName:   files.NewBytesFile(manifestBz),
		DWNConfigFileName:     files.NewBytesFile(cnfBz),
		IndexHTMLFileName:     files.NewBytesFile(IndexHTML),
		MainJSFileName:        files.NewBytesFile(MainJS),
		ServiceWorkerFileName: files.NewBytesFile(WorkerJS),
	}), nil
}

// NewVaultConfig returns the default vault config
func NewVaultConfig(addr string, ucanCID string) *motr.Config {
	return &motr.Config{
		MotrToken:      ucanCID,
		MotrAddress:    addr,
		IpfsGatewayUrl: "http://localhost:80",
		SonrApiUrl:     "http://localhost:1317",
		SonrRpcUrl:     "http://localhost:26657",
		SonrChainId:    "sonr-testnet-1",
	}
}
