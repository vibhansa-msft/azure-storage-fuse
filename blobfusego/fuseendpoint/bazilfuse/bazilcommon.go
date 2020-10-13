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

//var bazilCfg	*fs.Config
var bazilFS *FS

// FS is the File System created to serve the calls at user space
type FS struct {
	mountPath  string // Path to the root of this FS
	tempPath   string // Path to temp directory
	lastNodeId uint64 // Node id assigned to last new node
	rootDir    *Dir   // Pointer to Dir structure of the root
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

// nextID : Generate the next INode id for the element
func (fsys *FS) nextID() uint64 {
	if fsys != nil {
		Logger.LogDebug("FD : nextID returning : %d", (fsys.lastNodeId + 1))
		return atomic.AddUint64(&(fsys.lastNodeId), 1)
	}
	Logger.LogDebug("FD : nextID returning 1")
	return 1
}

// NewFS : Create the root directory holder for the mounted FS
func NewFS() *FS {
	Logger.LogDebug("FD : Creating the root structure for FS")

	fsys := &FS{
		mountPath:  *Config.BlobfuseConfig.MountPath,
		tempPath:   *Config.BlobfuseConfig.TmpPath,
		lastNodeId: 0,
		rootDir: &Dir{
			path:    "",
			nodelst: make(map[string]fs.Node),
		},
	}

	fsys.rootDir.nodeid = fsys.nextID()

	return fsys
}

///   STANDARD INTERFACE IMPLEMENTATATIONS

// Root : Create the root node for the FS
func (fsys *FS) Root() (n fs.Node, err error) {
	Logger.LogDebug("FD : Root called for " + fsys.mountPath)
	return fsys.rootDir, nil
}

// Statfs implements fsys.FSStatfser interface for *FS
func (fsys *FS) Statfs(ctx context.Context,
	req *fuse.StatfsRequest,
	resp *fuse.StatfsResponse) (err error) {

	Logger.LogDebug("FD : Statfs called for " + fsys.mountPath)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(fsys.tempPath, &stat); err != nil {
		Logger.LogErr("FD : Failed to do stat on root")
		return ErrToFuseErr(err)
	}

	resp.Blocks = stat.Blocks
	resp.Bfree = stat.Bfree
	resp.Bavail = stat.Bavail
	resp.Files = fsys.lastNodeId
	resp.Ffree = stat.Ffree
	resp.Bsize = uint32(stat.Bsize)

	return nil
}

// BlobAttrToFuseAttr : Convert Blob Attr to Fuse Attr
func BlobAttrToFuseAttr(fsAttr *FSIntf.BlobAttr, fuseAttr *fuse.Attr) {
	fuseAttr.Valid = time.Duration(*Config.BlobfuseConfig.AttrTimeOut)
	fuseAttr.Atime = fsAttr.Modtime
	fuseAttr.Mtime = fuseAttr.Atime
	fuseAttr.Ctime = fuseAttr.Atime
	fuseAttr.Crtime = fuseAttr.Atime

	if fsAttr.Flags.IsSet(FSIntf.PropFlagIsDir) {
		fuseAttr.Mode = os.ModeDir | Config.BlobfuseConfig.DefaultPerm
		fuseAttr.Size = 4096
	} else {
		fuseAttr.Mode = Config.BlobfuseConfig.DefaultPerm
		fuseAttr.Size = fsAttr.Size
	}

}

// Compile-time interface checks.
var _ fs.FS = (*FS)(nil)
var _ fs.FSStatfser = (*FS)(nil)

var _ fs.Node = (*Dir)(nil)
var _ fs.NodeCreater = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)
var _ fs.NodeRenamer = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)

var _ fs.HandleReadAller = (*File)(nil)
var _ fs.HandleWriter = (*File)(nil)
var _ fs.Node = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.NodeSetattrer = (*File)(nil)
var _ fs.HandleFlusher = (*File)(nil)
