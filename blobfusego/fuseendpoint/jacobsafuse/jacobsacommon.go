package jacobsafuse

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

type jacobNode struct {
	fuseutil.NotImplementedFileSystem

	inodes     []*inode          // GUARDED_BY(mu)
	freeInodes []fuseops.InodeID // GUARDED_BY(mu)
}

func NewJacobRoot() *jacobNode {
	// Set up the basic struct.
	fs := &jacobNode{
		inodes: make([]*inode, fuseops.RootInodeID+1),
	}

	// Set up the root inode.
	rootAttrs := fuseops.InodeAttributes{
		Mode: 0777 | os.ModeDir,
		Uid:  0,
		Gid:  0,
	}

	fs.inodes[fuseops.RootInodeID] = newInode(rootAttrs)
	return fs
}

func (fs *jacobNode) getInodeOrDie(id fuseops.InodeID) *inode {
	inode := fs.inodes[id]
	if inode == nil {
		panic(fmt.Sprintf("Unknown inode: %v", id))
	}

	return inode
}

func (fs *jacobNode) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (fs *jacobNode) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	//fs.mu.Lock()
	//defer fs.mu.Unlock()

	// Grab the parent directory.
	inode := fs.getInodeOrDie(op.Parent)

	// Does the directory have an entry with the given name?
	childID, _, ok := inode.LookUpChild(op.Name)
	if !ok {
		return fuse.ENOENT
	}

	// Grab the child.
	child := fs.getInodeOrDie(childID)

	// Fill in the response.
	op.Entry.Child = childID
	op.Entry.Attributes = child.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.Entry.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)
	op.Entry.EntryExpiration = op.Entry.AttributesExpiration

	return nil
}

func (fs *jacobNode) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	//fs.mu.Lock()
	//defer fs.mu.Unlock()

	// Grab the inode.
	inode := fs.getInodeOrDie(op.Inode)

	// Fill in the response.
	op.Attributes = inode.attrs

	// We don't spontaneously mutate, so the kernel can cache as long as it wants
	// (since it also handles invalidation).
	op.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)

	return nil
}

func (fs *jacobNode) GetXattr(ctx context.Context,
	op *fuseops.GetXattrOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	//fs.mu.Lock()
	//defer fs.mu.Unlock()

	inode := fs.getInodeOrDie(op.Inode)
	if value, ok := inode.xattrs[op.Name]; ok {
		op.BytesRead = len(value)
		if len(op.Dst) >= len(value) {
			copy(op.Dst, value)
		} else if len(op.Dst) != 0 {
			return syscall.ERANGE
		}
	} else {
		return fuse.ENOATTR
	}

	return nil
}
