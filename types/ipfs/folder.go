package ipfs

import (
	"context"

	"github.com/ipfs/boxo/files"
)

// Folder represents a directory structure that can be stored in IPFS.
// It's aliased to files.Directory from the boxo library for compatibility
// with IPFS directory operations.
type Folder = files.Directory

// NewFolder creates a new directory structure from a collection of files.
// It uses the NewFileMap function to organize the files into a map structure
// suitable for IPFS directory representation.
func NewFolder(fs ...File) Folder {
	return files.NewMapDirectory(NewFileMap(fs))
}

// AddFolder stores a complete directory structure in IPFS and returns the root CID.
// The folder and all its contained files are added to IPFS as a unified directory node.
// This method is implemented on the client to maintain consistency with the Client interface.
func (c *ipfsClient) AddFolder(folder Folder) (string, error) {
	cidFile, err := c.api.Unixfs().Add(context.Background(), folder)
	if err != nil {
		return "", err
	}
	return cidFile.String(), nil
}
