package fsinterface

import (
	"os"
	"time"
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

	OpenFile(string) error
	CloseFile(string) error

	ReadFile(string, int64, int64) ([]byte, error)
	WriteFile(string, int64, int64, []byte) (int, error)
	TruncateFile(string, int64) error

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

//////// Properties related interface and metadata

// BitMap : Generic BitMap to maintain flags
type BitMap uint16

// IsSet : Check whether the given bit is set or not
func (bm BitMap) IsSet(bit uint16) bool { return (bm & (1 << bit)) != 0 }

// Set : Set the given bit in bitmap
func (bm *BitMap) Set(bit uint16) { *bm |= (1 << bit) }

// Clear : Clear the given bit from bitmap
func (bm BitMap) Clear(bit uint16) { bm &= ^(1 << bit) }

// Flags represented in BitMap for various properties of the object
const (
	PropFlagUnknown uint16 = iota
	PropFlagNotExists
	PropFlagIsDir
	PropFlagEmptyDir
	PropFlagSymlink
	PropFlagMax
)

// BlobAttr : Attributes of any file/directory
type BlobAttr struct {
	Name    string      // name of the blob
	Size    uint64      // size of the object
	Mode    os.FileMode // permissions in 0xxx format
	Modtime time.Time   // last modified time
	Flags   BitMap      // Flags of the object
	NodeID  uint64      // Node Id of this element
}

// IsDir : Test blob is a directory or not
func (attr *BlobAttr) IsDir() bool {
	return attr.Flags.IsSet(PropFlagIsDir)
}

// IsSymlink : Test blob is a symlink or not
func (attr *BlobAttr) IsSymlink() bool {
	return attr.Flags.IsSet(PropFlagSymlink)
}
