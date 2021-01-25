package jacobsafuse

import (
	"context"
	"path"

	Logger "github.com/blobfusego/global/logger"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

func (n *jacobNode) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {

	Logger.LogDebug("FD : OpenDir called for " + n.Path())
	return nil
}

func (n *jacobNode) ReleaseDirHandle(
	ctx context.Context,
	req *fuseops.ReleaseDirHandleOp) (err error) {
	return nil
}

func (n *jacobNode) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {

	Logger.LogDebug("FD : ReadDir called for %s offset %d", n.Path(), op.Offset)
	if op.Offset != 0 {
		return nil
	}

	n.childLck.RLock()
	defer n.childLck.RUnlock()

	blobs, err := instance.client.ReadDir(n.Path())
	if err != nil {
		Logger.LogErr("FD : Failed to read directory (%s)", err)
		return err
	}

	Logger.LogDebug("FD : ReadDir came back with %d elements", len(blobs))
	op.BytesRead = 0
	var offset fuseops.DirOffset = 1

	for _, attr := range blobs {
		var nod *jacobNode
		if cnode, found := n.nameChild[attr.Name]; found {
			nod = cnode
		} else {
			nod = NewJacobNode(attr)
			n.AddChild(attr.Name, nod)
		}
		nod.RefreshAttr(attr)

		tmp := fuseutil.WriteDirent(op.Dst[op.BytesRead:], fuseutil.Dirent{
			Inode:  nod.nodeID,
			Name:   path.Base(attr.Name),
			Type:   nod.direntType,
			Offset: offset,
		})
		if tmp == 0 {
			break
		}
		offset++
		op.BytesRead += tmp
	}

	return nil
}

func (n *jacobNode) MkDir(
	ctx context.Context,
	op *fuseops.MkDirOp) error {

	Logger.LogErr("FD : Mkdir called for " + op.Name)
	return nil
}

func (n *jacobNode) RmDir(
	ctx context.Context,
	op *fuseops.RmDirOp) error {
	Logger.LogErr("FD : Rmdir called for " + op.Name)
	return nil
}
