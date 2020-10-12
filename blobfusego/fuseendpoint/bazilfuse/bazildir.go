package bazilfuse

import (
	Logger "github.com/blobfusego/global/logger"
	Conf   "github.com/blobfusego/global"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"

	"sync"
	"os"
	"bazil.org/fuse"
	//"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// Dir : Structure representing the directory
type Dir struct {
	// Lock to allow only one operation in a DIR at a time
	sync.RWMutex

	// Path to this dir
	path 	*string
}

// convertAttr : Convert the node attributes to fuse attributes
func BlobToBazilAttr(ba *FSIntf.BlobAttr, fa *fuse.Attr) {
	fa.Size = ba.Size
	fa.Mtime = ba.Modtime
	fa.Ctime = fa.Mtime
	fa.Crtime = fa.Mtime

	if ba.Flags.IsSet(FSIntf.PorpFlagIsDir) {
		// This is a directory
		fa.Mode = os.ModeDir | Conf.BlobfuseConfig.DefaultPerm
	} else if ba.Flags.IsSet(FSIntf.PropFlagSymlink) {
		// This is a symlink
		fa.Mode = os.ModeSymlink | ba.Mode
	} else {
		// This is a regular file
		fa.Mode = ba.Mode
	}
}

// Attr : Node interface for directories, to return attributes of the directory
func (d *Dir) Attr(ctx context.Context, o *fuse.Attr) error {
    Logger.LogDebug("FD : Dir Attr called for " + *d.path)
	d.RLock()
	defer d.RUnlock()

	if d.path == nil {
		// Called for the root directory
		o.Mode = os.ModeDir | Conf.BlobfuseConfig.DefaultPerm
		return nil
	}

	var attr FSIntf.BlobAttr
	err := instance.consumer.GetAttr(*d.path, &attr)
	if err != 0  {
		BlobToBazilAttr(&attr, o)
	}
	return nil
}

