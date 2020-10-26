package gofuse

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	Logger "github.com/blobfusego/global/logger"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var _ = (fs.NodeOpendirer)((*gofuseNode)(nil))
var _ = (fs.NodeReaddirer)((*gofuseNode)(nil))
var _ = (fs.NodeMkdirer)((*gofuseNode)(nil))
var _ = (fs.NodeMknoder)((*gofuseNode)(nil))
var _ = (fs.NodeRmdirer)((*gofuseNode)(nil))
var _ = (fs.NodeRenamer)((*gofuseNode)(nil))

// ignoreList : List holding objects to be ignored in getattr and return ENOENT
var ignoreList = map[string]struct{}{
	".Trash":           {},
	".Trash-1000":      {},
	".xdg-volume-info": {},
	"autorun.inf":      {},
}

// Lookup : Check whether object exists in the given path or not
func (n *gofuseNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	Logger.LogDebug("FD : Lookup for %s/%s", n.path(), name)

	p := filepath.Join(n.path(), name)

	if _, ignore := ignoreList[p]; ignore {
		Logger.LogDebug("FD : Ignoring %s", p)
		return nil, syscall.ENOENT
	}

	if nod, found := gofuseNodeMap[p]; found {
		out.FromStat(nod.stat)
		return n.NewInode(ctx, nod, fs.StableAttr{
			Mode: uint32(nod.stat.Mode),
			Gen:  1,
			Ino:  nod.nodeID,
		}), 0
	}

	return nil, syscall.ENOENT
}

// Mknod : Create new node for the given object
func (n *gofuseNode) Mknod(ctx context.Context, name string, mode, rdev uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	Logger.LogDebug("FD : Mknod for %s/%s", n.path(), name)

	p := filepath.Join(n.path(), name)

	attr, err := instance.client.GetAttr(p)
	if err != nil {
		Logger.LogErr("FD : Failed to get attribute %s (%s)", p, err.Error())
		return nil, fs.ToErrno(err)
	}

	_ = NewGofuseNode(attr.Name, attr)
	return n.Lookup(ctx, name, out)
}

// Opendir : Open a directory for reading, do nothing for now here
func (n *gofuseNode) Opendir(ctx context.Context) syscall.Errno {
	Logger.LogDebug("FD : Opendir for %s", n.path())
	return fs.OK
}

// Readdir : Return list of files/dir in the given path
// Get list from the client and return
func (n *gofuseNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	Logger.LogDebug("FD : Readdir for %s", n.path())

	blobs, err := instance.client.ReadDir(n.path())
	if err != nil {
		Logger.LogErr("FD : Failed to read directory (%s)", err)
		return nil, fs.ToErrno(err)
	}
	Logger.LogDebug("FD : ReadDir came back with %d elements", len(blobs))

	lst := []fuse.DirEntry{}
	for _, attr := range blobs {
		var nod *gofuseNode
		np := filepath.Join(n.path(), attr.Name)
		if n, found := gofuseNodeMap[np]; found {
			nod = n
			nod.refreshAttr(attr)
		} else {
			nod = NewGofuseNode(np, attr)
		}

		lst = append(lst, fuse.DirEntry{
			Mode: nod.stat.Mode,
			Name: attr.Name,
			Ino:  nod.nodeID,
		})
	}

	return fs.NewListDirStream(lst), 0

}

///// TODO : To be implemented yet

func (n *gofuseNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p := filepath.Join(n.path(), name)
	err := os.Mkdir(p, os.FileMode(mode))
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	//n.preserveOwner(ctx, p)
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		syscall.Rmdir(p)
		return nil, fs.ToErrno(err)
	}

	out.Attr.FromStat(&st)

	node := &gofuseNode{}
	ch := n.NewInode(ctx, node, fs.StableAttr{})

	return ch, 0
}

func (n *gofuseNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	p := filepath.Join(n.path(), name)
	err := syscall.Rmdir(p)
	return fs.ToErrno(err)
}
