package bazilfuse

import (
	"os"
	"sync"
	"sync/atomic"
	"syscall"

	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var conn *Conn
var fileSys *FS

// FS is the File System created to serve the calls at user space
type FS struct {
	rootPath    string					// Path to the root of this FS

	nodeLock	sync.RWMutex			// Lock for the nodeMap
	nodeMap		map[string][]*fs.Node		// List of nodes within this FS
	
	lastNodeId	uint64
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
    return atomic.AddUint64(&fsys.lastNodeId, 1)
}


// NewFS : Create the root directory holder for the mounted FS
func NewFS() *FS {
	Logger.LogDebug("FD : Creating the root structure for FS")

	fsys := &FS{
		rootPath: 	"/",
		nodeMap:    make(map[string][]*fs.Node),
	}

	return fsys
}

// NewNode : Create a new node for the FS
func (fsys *FS) NewNode(n *fs.Node) {
	path := n.getRealPath()
	Logger.LogDebug("FD : NewNode called for " + path)

	fsys.nodeLock.Lock()
	defer fsys.nodeLock.Unlock()
	fsys.nodeMap[path] = append(fsys.nodeMap[path], n)
	fsys.nextID()
}


// RenameNode : Rename existing node
func (fsys *FS) RenameNode(oldPath string, newPath string) {
	Logger.LogDebug("FD : RenameNode called for %s to %s", oldPath, newPath)

	fsys.nodeLock.Lock()
	defer fsys.nodeLock.Unlock()

	fsys.nodeMap[newPath] = append(fsys.nodeMap[newPath], fsys.nodeMap[oldPath]...)
	delete(fsys.nodeMap, oldPath)
	for _, n := range fsys.nodeMap[newPath] {
		n.updateRealPath(newPath)
	}
}

// ReleaseNode : Release / Delete the existsing node
func (fsys *FS) ReleaseNode(n *fs.Node) {
	Logger.LogDebug("FD : ReleaseNode called for " + n.realPath)

	fsys.nodeLock.Lock()
	defer fsys.nodeLock.Unlock()

	nodes, ok := fsys.nodeMap[n.realPath]
	if !ok {
		return
	}

	found := -1
	for i, node := range nodes {
		if node == n {
			found = i
			break
		}
	}

	if found > -1 {
		nodes = append(nodes[:found], nodes[found+1:]...)
	}

	if len(nodes) == 0 {
		delete(fsys.nodeMap, n.realPath)
	} else {
		fsys.nodeMap[n.realPath] = nodes
	}
}

///   STANDARD INTERFACE IMPLEMENTATATIONS

// Root : Create the root node for the FS
func (fsys *FS) Root() (n fs.Node, err error) {
	Logger.LogDebug("FD : Root called for " + fsys.rootPath)

	nn := &fs.Node{realPath: fsys.rootPath, isDir: true, fs: fsys}
	fsys.NewNode(nn)
	return nn, nil
}


// Statfs implements fsys.FSStatfser interface for *FS
func (fsys *FS) Statfs(ctx context.Context,
					req *fuse.StatfsRequest, 
					resp *fuse.StatfsResponse) (err error) {

	Logger.LogDebug("FD : Statfs called for " + fsys.rootPath)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(fsys.rootPath, &stat); err != nil {
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