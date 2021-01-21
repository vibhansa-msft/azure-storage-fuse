package jacobsafuse

import (
	"context"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
)

func (fs *jacobNode) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	//fs.mu.Lock()
	//defer fs.mu.Unlock()

	// We don't mutate spontaneosuly, so if the VFS layer has asked for an
	// inode that doesn't exist, something screwed up earlier (a lookup, a
	// cache invalidation, etc.).
	inode := fs.getInodeOrDie(op.Inode)

	if !inode.isDir() {
		panic("Found non-dir.")
	}

	return nil
}

func (fs *jacobNode) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	if op.OpContext.Pid == 0 {
		return fuse.EINVAL
	}

	//fs.mu.Lock()
	//defer fs.mu.Unlock()

	// Grab the directory.
	inode := fs.getInodeOrDie(op.Inode)

	// Serve the request.
	op.BytesRead = inode.ReadDir(op.Dst, int(op.Offset))

	return nil
}
