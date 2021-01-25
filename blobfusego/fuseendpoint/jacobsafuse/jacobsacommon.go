package jacobsafuse

import (
	"context"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

// ignoreList : List holding objects to be ignored in getattr and return ENOENT
var ignoreList = map[string]struct{}{
	".Trash":           {},
	".Trash-1000":      {},
	".xdg-volume-info": {},
	"autorun.inf":      {},
}

type jacobNode struct {
	fuseutil.NotImplementedFileSystem

	nodePath string          // Full path of the object
	nodeID   fuseops.InodeID // NodeId assigned to this object

	attrs fuseops.InodeAttributes

	// Maintain a map searchable by inode id and name for all its children
	child      map[fuseops.InodeID]*jacobNode
	nameChild  map[string]*jacobNode
	direntType fuseutil.DirentType

	childLck sync.RWMutex
}

func nextID() uint64 {
	if instance == nil {
		return 1
	}

	return atomic.AddUint64(&instance.nodeID, 1)
}

// Path : Get the path of this node from root of mounted directory
func (n *jacobNode) Path() string {
	return n.nodePath
}

func (n *jacobNode) isDir() bool {
	return n.attrs.Mode&os.ModeDir != 0
}

func (n *jacobNode) isSymlink() bool {
	return n.attrs.Mode&os.ModeSymlink != 0
}

func (n *jacobNode) isFile() bool {
	return !(n.isDir() || n.isSymlink())
}

func (n *jacobNode) setDir(bool) {
	n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeDir
}

func (n *jacobNode) setFile(bool) {
	n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeIrregular
}

func (n *jacobNode) setSymlink(bool) {
	n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeSymlink
}

func (n *jacobNode) GetChildByInode(id fuseops.InodeID) *jacobNode {
	Logger.LogDebug("FD : GetChildByInode called for %s (%d)", n.Path(), id)

	n.childLck.RLock()
	defer n.childLck.RUnlock()

	var child *jacobNode
	var found bool
	if child, found = n.child[id]; !found {
		Logger.LogErr("FD : LookUpInode failed for inode %d", id)
		return nil
	}

	Logger.LogDebug("FD : WriteFile called for path : %s", child.Path())
	return child
}

// NewJacobRoot : Create the root node for the mounted FS
func NewJacobRoot() *jacobNode {
	Logger.LogDebug("FD : NewJacobRoot called")

	fs := &jacobNode{
		nodePath: "",
		nodeID:   fuseops.RootInodeID,

		attrs: fuseops.InodeAttributes{
			Mode:   Config.BlobfuseConfig.DefaultPerm | os.ModeDir,
			Uid:    0,
			Gid:    0,
			Nlink:  1,
			Size:   0,
			Mtime:  time.Now(),
			Atime:  time.Now(),
			Ctime:  time.Now(),
			Crtime: time.Now(),
		},
		direntType: fuseutil.DT_Directory,
		child:      make(map[fuseops.InodeID]*jacobNode),
		nameChild:  make(map[string]*jacobNode),
		childLck:   sync.RWMutex{},
	}

	instance.nodeID = fuseops.RootInodeID
	return fs
}

// NewJacobNode : Create the node for given object
func NewJacobNode(attr FSIntf.BlobAttr) *jacobNode {
	Logger.LogDebug("FD : NewJacobNode called")

	fs := &jacobNode{
		nodePath: attr.Name,
		nodeID:   fuseops.InodeID(nextID()),

		attrs: fuseops.InodeAttributes{
			Mode:   Config.BlobfuseConfig.DefaultPerm,
			Uid:    0,
			Gid:    0,
			Nlink:  1,
			Size:   0,
			Mtime:  time.Now(),
			Atime:  time.Now(),
			Ctime:  time.Now(),
			Crtime: time.Now(),
		},
		childLck: sync.RWMutex{},
	}

	if attr.IsDir() {
		fs.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeDir
		fs.direntType = fuseutil.DT_Directory
		fs.child = make(map[fuseops.InodeID]*jacobNode)
		fs.nameChild = make(map[string]*jacobNode)
	} else if attr.IsSymlink() {
		fs.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeSymlink
		fs.direntType = fuseutil.DT_Link
	} else {
		fs.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeIrregular
		fs.direntType = fuseutil.DT_File
	}
	return fs
}

func (n *jacobNode) AddChild(name string, node *jacobNode) error {
	n.child[node.nodeID] = node
	n.nameChild[name] = node

	node.nodePath = path.Join(n.nodePath, name)
	return nil
}

func (n *jacobNode) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {

	Logger.LogDebug("FD : Statfs called for " + n.Path())

	var stat syscall.Statfs_t
	if err := syscall.Statfs(*Config.BlobfuseConfig.TmpPath, &stat); err != nil {
		Logger.LogErr("FD : Failed to do stat on root")
		return err
	}

	op.BlockSize = uint32(stat.Bsize)
	op.Blocks = stat.Blocks
	op.BlocksAvailable = stat.Bavail
	op.BlocksFree = stat.Bfree
	return nil
}

func (n *jacobNode) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {

	Logger.LogDebug("FD : GetInodeAttributes called for " + n.Path())

	if n.nodeID == fuseops.RootInodeID {
		op.Attributes = instance.rootFD.attrs
		op.AttributesExpiration = time.Now().Add(120 * time.Second)
		op.Inode = fuseops.InodeID(fuseops.RootInodeID)
		return nil
	}

	n.childLck.RLock()
	defer n.childLck.RUnlock()

	var child *jacobNode
	var found bool
	if child, found = n.child[op.Inode]; !found {
		Logger.LogErr("FD : GetInodeAttributes failed for inode " + string(op.Inode))
		return fuse.ENOENT
	}

	op.Attributes = child.attrs
	op.AttributesExpiration = time.Now().Add(120 * time.Second)
	op.Inode = fuseops.InodeID(child.nodeID)
	return nil
}

func (n *jacobNode) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {

	Logger.LogDebug("FD : LookUpInode called for " + n.Path() + "/" + op.Name)
	if _, ignore := ignoreList[op.Name]; ignore {
		Logger.LogDebug("FD : Ignoring %s", op.Name)
		return syscall.ENOENT
	}

	n.childLck.RLock()
	defer n.childLck.RUnlock()

	var child *jacobNode
	var found bool
	if child, found = n.nameChild[op.Name]; !found {
		Logger.LogErr("FD : LookUpInode failed for inode " + op.Name)
		return fuse.ENOENT
	}

	op.Entry.Child = child.nodeID
	op.Entry.Attributes = child.attrs
	op.Entry.AttributesExpiration = time.Now().Add(time.Second * 120)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	return nil
}

func (n *jacobNode) GetXattr(ctx context.Context,
	op *fuseops.GetXattrOp) error {
	return nil
}

func (n *jacobNode) SetInodeAttributes(
	ctx context.Context,
	op *fuseops.SetInodeAttributesOp) error {
	return nil
}

func (n *jacobNode) RefreshAttr(attr FSIntf.BlobAttr) {
	n.attrs.Mtime = attr.Modtime
	n.attrs.Ctime = n.attrs.Mtime
	n.attrs.Crtime = n.attrs.Mtime
	n.attrs.Size = uint64(attr.Size)

	if attr.IsDir() {
		n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeDir
		n.direntType = fuseutil.DT_Directory
		n.attrs.Nlink = 2
	} else if attr.IsSymlink() {
		n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeSymlink
		n.direntType = fuseutil.DT_Link
	} else {
		n.attrs.Mode = Config.BlobfuseConfig.DefaultPerm | os.ModeIrregular
		n.direntType = fuseutil.DT_File
	}
}
