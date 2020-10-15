package fsinterface

import (
	"os"
	"time"
)

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
