package fsinterface

import (
	"os"
)

// FileSystem : Master interface for the file system
// All implementations shall register to the factory and use it get the respective object
type FileSystem interface {

	// Init/DeInit the filesystem
	InitFS() int
	DeInitFs() int

	// Set the next component in pipeline for this system
	SetClient(cons FileSystem) int

	// Get the file system name
	GetName() string

	// Get the reference count
	GetCount() int

	// Print the pipeline
	PrintPipeline() string

	// Get the file system stats
	StatFS() error

	// Directory level operations
	CreateDir(string, os.FileMode) error
	DeleteDir(string) error

	OpenDir(string) error
	CloseDir(string) error

	ReadDir(string) ([]BlobAttr, error)
	RenameDir(string, string) error

	// File level operations
	CreateFile(string, os.FileMode) error
	DeleteFile(string) error

	OpenFile(string, int, os.FileMode) error
	CloseFile(string) error

	ReadFile(string, int64, int64) ([]byte, error)
	WriteFile(string, int64, int64, []byte) (int, error)
	TruncateFile(string, int64) error

	CopyToFile(string, *os.File) error
	CopyFromFile(string, *os.File) error

	FlushFile(string) error
	ReleaseFile(string) error
	UnlinkFile(string) error

	// Symlink operations
	CreateLink(string, string) error
	ReadLink(string) (string, error)

	// Filesystem level operations
	GetAttr(string) (BlobAttr, error)
	SetAttr(string, BlobAttr) error

	Chmod(string, os.FileMode) error
	Chown(string, string) error
}
