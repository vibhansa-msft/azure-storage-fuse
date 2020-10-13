package bazilfuse

import (
	"os"
	"sync"
	"time"

	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// Dir : Structure holding the Dir info for fuse
type Dir struct {
	path   string
	dirlck sync.RWMutex
	nodeid uint64
	fs     *FS
}

// Attr : Get the attributes of the given directory
func (d *Dir) Attr(ctx context.Context, o *fuse.Attr) error {
	Logger.LogDebug("FD : Dir Attr called for %s", d.path)

	if d.path == "" {
		// Called for mount path
		Logger.LogDebug("FD : Dir Attr called for mount point")
		o.Inode = d.nodeid
		o.Valid = time.Duration(*Config.BlobfuseConfig.AttrTimeOut)
		o.Atime = Config.BlobfuseConfig.MountTime
		o.Mtime = o.Atime
		o.Ctime = o.Atime
		o.Crtime = o.Atime

		o.Mode = os.ModeDir | Config.BlobfuseConfig.DefaultPerm
		o.Size = 4096

		o.Uid = 0
		o.Gid = 0
	}

	return nil
}

// Lookup : Check whether given object exists in the directory structure or not
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	Logger.LogDebug("FD : Dir Lookup called for %s : %s", d.path, name)

	// Ignore certain linux standard things
	if name == ".Trash" ||
		name == ".Trash-1000" ||
		name == ".xdg-volume-info" ||
		name == "autorun.inf" {
		Logger.LogDebug("FD : Dir Lookup ignored for %s", name)
		return nil, fuse.ENOENT
	}
	return &Dir{}, nil
}

// ReadDirAll : Get the list of objects from a directory
func (d *Dir) ReadDirAll(ctx context.Context) (dirs []fuse.Dirent, err error) {
	Logger.LogDebug("FD : Dir ReadDirAll called for %s", d.path)
	return dirs, err
}

// Mkdir : Create a new directory
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (n fs.Node, err error) {
	Logger.LogDebug("FD : Dir Mkdir called for %s", d.path)
	return n, nil
}

// Create : Create a new node in directory
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (n fs.Node, h fs.Handle, err error) {
	Logger.LogDebug("FD : Dir Create called for %s", d.path)
	return n, h, nil
}

// Rename : Rename a directory
func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	Logger.LogDebug("FD : Dir Rename called for %s", d.path)
	return nil
}

// Remove : Delete a directory
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	Logger.LogDebug("FD : Dir Remove called for %s", d.path)
	return nil
}
