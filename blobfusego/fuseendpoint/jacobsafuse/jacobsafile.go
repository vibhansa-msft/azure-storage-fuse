package jacobsafuse

import (
	"context"
	"io"
	"path/filepath"

	Logger "github.com/blobfusego/global/logger"
	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
)

func (n *jacobNode) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {

	Logger.LogDebug("FD : OpenFile called for %s (%d)", n.Path(), op.Inode)

	child := n.GetChildByInode(op.Inode)
	if child == nil {
		return fuse.ENOENT
	}

	Logger.LogDebug("FD : OpenFile called for path : %s", child.Path())
	if err := instance.client.OpenFile(child.Path(), 0, 0); err != nil {
		Logger.LogErr("FD : Failed to open file %s (%s)", child.Path(), err)
		return err
	}

	return nil
}

func (n *jacobNode) CreateFile(
	ctx context.Context,
	op *fuseops.CreateFileOp) (err error) {

	Logger.LogDebug("FD : CreateFile called for " + filepath.Join(n.Path(), op.Name))

	p := filepath.Join(n.Path(), op.Name)

	if err := instance.client.CreateFile(p, 0); err != nil {
		Logger.LogErr("FD : Failed to create file %s (%s)", n.Path(), err.Error())
		return err
	}
	return nil
}

func (n *jacobNode) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) (err error) {

	Logger.LogDebug("FD : ReadFile called for %s (%d)", n.Path(), op.Inode)

	child := n.GetChildByInode(op.Inode)
	if child == nil {
		return fuse.ENOENT
	}

	Logger.LogDebug("FD : ReadFile called for path : %s offset %d len %d", child.Path(), op.Offset, len(op.Dst))

	op.Dst, err = instance.client.ReadFile(child.Path(), op.Offset, int64(len(op.Dst)))
	if err != nil && err != io.EOF {
		Logger.LogErr("FD : Failed to read the file %s (%s)", n.Path(), err.Error())
		return err
	}

	op.BytesRead = len(op.Dst)
	return nil
}

func (n *jacobNode) WriteFile(
	ctx context.Context,
	op *fuseops.WriteFileOp) error {

	Logger.LogDebug("FD : WriteFile called for %s (%d)", n.Path(), op.Inode)

	child := n.GetChildByInode(op.Inode)
	if child == nil {
		return fuse.ENOENT
	}

	Logger.LogDebug("FD : WriteFile called for path : %s offset %d len %d", child.Path(), op.Offset, len(op.Data))

	bytes, err := instance.client.WriteFile(n.Path(), op.Offset, int64(len(op.Data)), op.Data)
	if err != nil || bytes != len(op.Data) {
		Logger.LogErr("FD : Failed to read the file %s (%s)", n.Path(), err.Error())
		return err
	}
	return nil
}

func (n *jacobNode) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) (err error) {

	Logger.LogDebug("FD : FlushFile called for %s (%d)", n.Path(), op.Inode)

	child := n.GetChildByInode(op.Inode)
	if child == nil {
		return fuse.ENOENT
	}

	Logger.LogDebug("FD : FlushFile called for path : %s", child.Path())

	_ = instance.client.CloseFile(child.Path())
	return nil
}

func (n *jacobNode) ReleaseFileHandle(ctx context.Context, req *fuseops.ReleaseFileHandleOp) (err error) {
	return
}
