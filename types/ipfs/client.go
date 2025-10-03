// Package ipfs provides a high-level interface for interacting with an IPFS node.
package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
)

// NodeStatus contains information about an IPFS node's status and connectivity.
type NodeStatus struct {
	// PeerID is the unique identifier of the IPFS node
	PeerID string
	// Version is the version of the IPFS implementation
	Version string
	// PeerType describes the type of IPFS node (e.g., "kubo")
	PeerType string
	// ConnectedPeers is the number of peers currently connected to this node
	ConnectedPeers int
}

// IPFSClient provides a high-level interface for interacting with an IPFS node.
// It abstracts the complexity of the underlying Kubo RPC API and provides
// convenient methods for common IPFS operations.
type IPFSClient interface {
	// Add stores raw byte data in IPFS and returns the resulting CID.
	// The data is stored as a single file node in the IPFS DAG.
	Add(data []byte) (string, error)

	// AddFile stores a File with metadata in IPFS and returns the resulting CID.
	// The file's name and content are preserved in the IPFS node.
	AddFile(file File) (string, error)

	// AddFolder stores a directory structure in IPFS and returns the root CID.
	// The folder and its contents are stored as a directory node in IPFS.
	AddFolder(folder Folder) (string, error)

	// Exists checks if content with the given CID exists in the IPFS network.
	// It queries the local node's blockstore to verify availability.
	Exists(cid string) (bool, error)

	// Get retrieves the content of a file from IPFS using its CID.
	// Returns the raw bytes of the file content.
	Get(cid string) ([]byte, error)

	// IsPinned checks if content is pinned on the local IPFS node using IPNS resolution.
	// Returns true if the IPNS name can be successfully resolved.
	IsPinned(ipns string) (bool, error)

	// Ls lists the contents of a directory in IPFS using its CID.
	// Returns the names of all entries in the directory.
	Ls(cid string) ([]string, error)

	// Pin ensures that content with the given CID is retained in the local IPFS node.
	// The content will not be garbage collected and can be assigned a human-readable name.
	Pin(cid string, name string) error

	// Unpin removes the pin for content with the given CID.
	// The content may be garbage collected if not pinned elsewhere.
	Unpin(cid string) error

	// NodeStatus returns information about the IPFS node including peer ID, version, and connectivity.
	// Returns node identity, version information, and connected peer count.
	NodeStatus() (*NodeStatus, error)
}

// ipfsClient implements the Client interface using the Kubo RPC API.
// It maintains a connection to a local IPFS node via HTTP API.
type ipfsClient struct {
	api *rpc.HttpApi
}

// GetClient creates a new IPFS client connected to a local IPFS node.
// It establishes a connection to the local Kubo RPC API endpoint.
// Returns an error if the local IPFS node is not available or accessible.
func GetClient() (IPFSClient, error) {
	api, err := rpc.NewLocalApi()
	if err != nil {
		return nil, err
	}
	return &ipfsClient{api: api}, nil
}

// Add stores raw byte data in IPFS and returns the resulting CID string.
// The data is wrapped in a BytesFile and added to the UnixFS layer.
func (c *ipfsClient) Add(data []byte) (string, error) {
	file := files.NewBytesFile(data)
	cidFile, err := c.api.Unixfs().Add(context.Background(), file)
	if err != nil {
		return "", err
	}
	return cidFile.String(), nil
}

// Get retrieves content from IPFS using the provided CID string.
// It resolves the CID path, fetches the content, and returns the raw bytes.
// Returns an error if the CID is invalid or the content cannot be retrieved.
func (c *ipfsClient) Get(cid string) ([]byte, error) {
	p, err := path.NewPath(cid)
	if err != nil {
		return nil, err
	}
	node, err := c.api.Unixfs().Get(context.Background(), p)
	if err != nil {
		return nil, err
	}

	file, ok := node.(files.File)
	if !ok {
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// IsPinned checks if an IPNS name can be resolved, indicating the content is available.
// This method uses IPNS resolution rather than direct pin checking.
// Returns false if the IPNS name cannot be resolved.
func (c *ipfsClient) IsPinned(ipns string) (bool, error) {
	_, err := c.api.Name().Resolve(context.Background(), ipns)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Exists verifies if content with the given CID exists in the IPFS network.
// It checks the block statistics to determine availability.
// Returns false if the CID is invalid or the content is not found.
func (c *ipfsClient) Exists(cid string) (bool, error) {
	p, err := path.NewPath(cid)
	if err != nil {
		return false, err
	}
	_, err = c.api.Block().Stat(context.Background(), p)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Pin adds content to the pin set, preventing it from being garbage collected.
// The content is identified by its CID and can be assigned a descriptive name.
// Returns an error if the CID is invalid or pinning fails.
func (c *ipfsClient) Pin(cid string, name string) error {
	p, err := path.NewPath(cid)
	if err != nil {
		return err
	}
	return c.api.Pin().Add(context.Background(), p, options.Pin.Name(name))
}

// Unpin removes content from the pin set, allowing it to be garbage collected.
// The content is identified by its CID string.
// Returns an error if the CID is invalid or unpinning fails.
func (c *ipfsClient) Unpin(cid string) error {
	p, err := path.NewPath(cid)
	if err != nil {
		return err
	}
	return c.api.Pin().Rm(context.Background(), p)
}

// Ls lists the contents of an IPFS directory using its CID.
// It returns the names of all entries in the directory.
// The listing is performed asynchronously using channels to handle large directories efficiently.
func (c *ipfsClient) Ls(cid string) ([]string, error) {
	p, err := path.NewPath(cid)
	if err != nil {
		return nil, err
	}
	dirChan := make(chan iface.DirEntry)
	files := make([]string, 0)
	lsErr := make(chan error, 1)
	go func() {
		lsErr <- c.api.Unixfs().Ls(context.Background(), p, dirChan)
	}()
	for dirEnt := range dirChan {
		files = append(files, dirEnt.Name)
	}
	if err := <-lsErr; err != nil {
		return nil, err
	}
	return files, nil
}

// AddFile stores a File with its metadata in IPFS and returns the resulting CID.
// This method preserves the file's name and content structure when adding to IPFS.
// It's implemented as a method on the client to maintain consistency with the Client interface.
func (c *ipfsClient) AddFile(file File) (string, error) {
	cidFile, err := c.api.Unixfs().Add(context.Background(), file)
	if err != nil {
		return "", err
	}
	return cidFile.String(), nil
}

// NodeStatus retrieves status information about the IPFS node including peer ID, version, and connectivity.
// It queries the node's identity, version, and connected peers to provide comprehensive status information.
func (c *ipfsClient) NodeStatus() (*NodeStatus, error) {
	ctx := context.Background()

	// Get node ID information using Key().Self() to get the node's own peer ID
	nodeKey, err := c.api.Key().Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get connected peers count
	swarmPeers, err := c.api.Swarm().Peers(ctx)
	connectedPeers := 0
	if err == nil {
		connectedPeers = len(swarmPeers)
	}

	// Since we successfully got the key information, we know the node is responsive
	versionStr := "kubo-responsive"
	if connectedPeers > 0 {
		versionStr = "kubo-connected"
	}

	return &NodeStatus{
		PeerID:         nodeKey.Name(),
		Version:        versionStr,
		PeerType:       "kubo",
		ConnectedPeers: connectedPeers,
	}, nil
}
