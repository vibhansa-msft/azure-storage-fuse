package bazilfuse

import (
	"os"
	"sync/atomic"
	"syscall"
	"time"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var bazilConn *fuse.Conn

// BazilFS : Pointer to Bazil FS structure
var BazilFS *FS

// Compile-time interface checks.
var _ fs.FS = (*FS)(nil)
var _ fs.FSStatfser = (*FS)(nil)

var nodeMap = make(map[string]fs.Node)

// FS : Base structure for this file system
type FS struct {
	rootPath string
	root     *Dir
	NodeID   uint64
	size     int64

	client FSIntf.FileSystem
}

// ErrToFuseErr : Convert OS error to Fuse err codes
func ErrToFuseErr(err error) error {
	switch {
	case os.IsNotExist(err):
		return fuse.ENOENT
	case os.IsExist(err):
		return fuse.EEXIST
	case os.IsPermission(err):
		return fuse.EPERM
	default:
		return fuse.EIO
	}
}

func (fsys *FS) nextID() uint64 {
	if fsys == nil {
		return 1
	}

	return atomic.AddUint64(&fsys.NodeID, 1)
}

// NewBazilFS : Create the root directory holder for the mounted FS
func NewBazilFS() *FS {
	Logger.LogDebug("FD : Creating the root structure for FS")

	fsys := &FS{
		rootPath: *Config.BlobfuseConfig.MountPath,
		root:     nil,
		NodeID:   0,
		size:     0,
		client:   nil,
	}

	fsys.root = fsys.newDirNode("", &FSIntf.BlobAttr{
		Name:    "/",
		Size:    4096,
		Mode:    os.ModeDir | Config.BlobfuseConfig.DefaultPerm,
		Modtime: time.Now(),
	})

	if fsys.root.attr.Inode != 1 {
		Logger.LogDebug("FD : Root shall have Inode of 1")
		return nil
	}
	return fsys
}

///   STANDARD INTERFACE IMPLEMENTATATIONS

// Root : Create the root node for the FS
func (fsys *FS) Root() (n fs.Node, err error) {
	Logger.LogDebug("FD : Root called for " + fsys.rootPath)
	return fsys.root, nil
}

// Statfs implements fsys.FSStatfser interface for *FS
func (fsys *FS) Statfs(ctx context.Context,
	req *fuse.StatfsRequest,
	resp *fuse.StatfsResponse) (err error) {

	Logger.LogDebug("FD : Statfs called for " + fsys.rootPath)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(*Config.BlobfuseConfig.TmpPath, &stat); err != nil {
		Logger.LogErr("FD : Failed to do stat on root")
		return ErrToFuseErr(err)
	}

	resp.Blocks = stat.Blocks
	resp.Bfree = stat.Bfree
	resp.Bavail = stat.Bavail
	resp.Files = fsys.NodeID
	resp.Ffree = stat.Ffree
	resp.Bsize = uint32(stat.Bsize)

	return nil
}

func (fsys *FS) newFileNode(path string, attr *FSIntf.BlobAttr) *File {
	f := &File{
		attr: fuse.Attr{
			Valid:  time.Duration(*Config.BlobfuseConfig.AttrTimeOut),
			Inode:  fsys.nextID(),
			Atime:  attr.Modtime,
			Mtime:  attr.Modtime,
			Ctime:  attr.Modtime,
			Crtime: attr.Modtime,
			Mode:   attr.Mode,
		},
		path:  path,
		valid: true,
	}
	if attr.IsSymlink() {
		f.attr.Mode |= os.ModeSymlink
	}
	nodeMap[path] = f
	return f
}

func (fsys *FS) newDirNode(path string, attr *FSIntf.BlobAttr) *Dir {
	d := &Dir{
		attr: fuse.Attr{
			Valid:  time.Duration(*Config.BlobfuseConfig.AttrTimeOut),
			Inode:  fsys.nextID(),
			Atime:  attr.Modtime,
			Mtime:  attr.Modtime,
			Ctime:  attr.Modtime,
			Crtime: attr.Modtime,
			Mode:   os.ModeDir | attr.Mode,
		},
		path:  path,
		valid: true,
	}
	nodeMap[path] = d

	return d
}

// SetDirAttr : Refresh the directory attributes
func (d *Dir) SetDirAttr(attr *FSIntf.BlobAttr) {
	d.attr = fuse.Attr{
		Valid:  time.Duration(*Config.BlobfuseConfig.AttrTimeOut),
		Inode:  d.attr.Inode,
		Atime:  attr.Modtime,
		Mtime:  attr.Modtime,
		Ctime:  attr.Modtime,
		Crtime: attr.Modtime,
		Mode:   os.ModeDir | attr.Mode,
	}
}

// SetFileAttr : Refresh the File attributes
func (f *File) SetFileAttr(attr *FSIntf.BlobAttr) {
	f.attr = fuse.Attr{
		Valid:  time.Duration(*Config.BlobfuseConfig.AttrTimeOut),
		Inode:  f.attr.Inode,
		Atime:  attr.Modtime,
		Mtime:  attr.Modtime,
		Ctime:  attr.Modtime,
		Crtime: attr.Modtime,
		Mode:   attr.Mode,
	}
	if attr.IsSymlink() {
		f.attr.Mode |= os.ModeSymlink
	}
}
