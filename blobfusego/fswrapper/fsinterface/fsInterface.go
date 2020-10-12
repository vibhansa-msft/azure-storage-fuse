
package fsinterface

import (
	"time"
	"os"
)

// FileSystem : Master interface for the file system
// All implementations shall register to the factory and use it get the respective object
type FileSystem interface {

	// Init/DeInit the filesystem
	InitFS() int
	DeInitFs() int

	// Set the next component in pipeline for this system
	SetConsumer(cons FileSystem) int
	
	// Get the file system name
	GetName() string

	// Get the reference count
	GetCount() int

	// Print the pipeline
	PrintPipeline() string
	
	// Get the file system stats
	StatFS() int

	// Directory level operations
	CreateDir	(path string) int
	DeleteDir	(path string)

	OpenDir		(path string) int
	CloseDir	(path string)

	ReadDir		(path string) int
	RenameDir	(path string, name string) int


	// File level operations
	CreateFile	(path string, mode int) int
	DeleteFile	(path string) int

	OpenFile	(path string, mode int) int
	CloseFile	(path string)

	ReadFile	(path string, offset int, length int) int
	WriteFile	(path string, offset int, length int) int

	FlushFile	(path string) int
	ReleaseFile	(path string) int
	UnlinkFile	(path string) int

	// Symlink operations
	CreateLink	(path string, dst string) int
	ReadLink	(path string, link string) int

	// Filesystem level operations
	GetAttr		(path string, attr *BlobAttr) int
	SetAttr		(path string) int

	Chmod		(path string, mod int) int
	Chown		(path string, owner string) int
}


//////// Properties related interface and metadata

// BitMap : Generic BitMap to maintain flags
type BitMap uint16

// IsSet : Check whether the given bit is set or not
func (bm BitMap) IsSet(bit uint16) bool		{ return (bm & (1 << bit)) != 0}
// Set : Set the given bit in bitmap
func (bm BitMap) Set(bit uint16) 			{ bm |= (1 << bit)}
// Clear : Clear the given bit from bitmap
func (bm BitMap) Clear(bit uint16)			{ bm &= ^(1 << bit)}

// Flags represented in BitMap for various properties of the object
const (
    PropFlagUnknown       uint16 = iota
    PorpFlagNotExists
    PorpFlagIsDir
    PorpFlagEmptyDir
	PropFlagSymlink
    PropFlagMax   
)


// BlobAttr : Attributes of any file/directory
type BlobAttr struct {
	Size		uint64			// size of the object
	Mode		os.FileMode		// permissions in 0xxx format
	Modtime		time.Time		// last modified time
	Flags		BitMap			// Flags of the object
}

