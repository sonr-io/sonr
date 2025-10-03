package ipfs

import "github.com/ipfs/boxo/files"

// File represents a file that can be stored in IPFS.
// It extends the boxo files.File interface with a Name method
// to provide file metadata along with content.
type File interface {
	files.File
	Name() string
}

// NewFileMap creates a map suitable for creating IPFS directory structures.
// It maps each file's name to its corresponding files.Node for use with
// the boxo files library.
func NewFileMap(vs []File) map[string]files.Node {
	m := make(map[string]files.Node)
	for _, f := range vs {
		m[f.Name()] = f
	}
	return m
}

// file implements the File interface, wrapping a boxo files.File
// with additional metadata like the file name. This allows for
// proper file representation when storing content in IPFS.
type file struct {
	files.File
	name string
}

// Name returns the filename associated with this file.
// This method satisfies the File interface requirement.
func (f *file) Name() string {
	return f.name
}

// NewFile creates a new File instance from a name and byte data.
// The returned File can be used with IPFS operations and preserves
// the original filename for directory structures.
func NewFile(name string, data []byte) File {
	return &file{File: files.NewBytesFile(data), name: name}
}
