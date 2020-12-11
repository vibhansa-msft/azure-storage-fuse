package gofuse

import (
	"context"
	"io"
	"path/filepath"
	"syscall"

	Logger "github.com/blobfusego/global/logger"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Refer for all operatoins and options : https://github.com/hanwen/go-fuse/blob/master/fuse/nodefs/api.go

var _ = (fs.NodeCreater)((*gofuseNode)(nil))
var _ = (fs.NodeOpener)((*gofuseNode)(nil))
var _ = (fs.NodeReader)((*gofuseNode)(nil))
var _ = (fs.NodeWriter)((*gofuseNode)(nil))
var _ = (fs.NodeFlusher)((*gofuseNode)(nil))

var _ = (fs.NodeReadlinker)((*gofuseNode)(nil))
var _ = (fs.NodeLinker)((*gofuseNode)(nil))
var _ = (fs.NodeSymlinker)((*gofuseNode)(nil))
var _ = (fs.NodeUnlinker)((*gofuseNode)(nil))

var _ = (fs.NodeSetattrer)((*gofuseNode)(nil))

var data = make([]byte, 10)

func (n *gofuseNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (inode *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	Logger.LogDebug("FD : File Create (%s/%s, %d)\n", n.path(), name, flags)

	p := filepath.Join(n.path(), name)

	if err := instance.client.CreateFile(p, 0); err != nil {
		Logger.LogErr("FD : Failed to create file %s (%s)", n.path(), err.Error())
		return nil, nil, 0, fs.ToErrno(err)
	}

	attr, err := instance.client.GetAttr(p)
	if err != nil {
		Logger.LogErr("FD : Failed to get attribute %s (%s)", p, err.Error())
		return nil, nil, 0, fs.ToErrno(err)
	}

	nod := NewGofuseNode(p, attr)
	return n.NewInode(ctx, nod, fs.StableAttr{
		Mode: uint32(nod.stat.Mode),
		Gen:  1,
		Ino:  nod.nodeID,
	}), nil, 0, 0
}

func (n *gofuseNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	Logger.LogDebug("FD : File Open (%s, %d)\n", n.path(), flags)

	if ((flags & syscall.O_APPEND) == syscall.O_APPEND) ||
		((flags & syscall.O_CREAT) == syscall.O_CREAT) ||
		((flags & syscall.O_WRONLY) == syscall.O_WRONLY) {
		if err := instance.client.CreateFile(n.path(), 0); err != nil {
			Logger.LogErr("FD : Failed to create file %s (%s)", n.path(), err.Error())
			return nil, 0, fs.ToErrno(err)
		}
	}

	if err := instance.client.OpenFile(n.path(), int(flags), 0); err != nil {
		Logger.LogErr("FD : Failed to open file %s (%s)", n.path(), err)
		return nil, 0, fs.ToErrno(err)
	}
	return fh, 0, 0
}

func (n *gofuseNode) Flush(ctx context.Context, fh fs.FileHandle) syscall.Errno {
	Logger.LogDebug("FD : File Flush %s\n", n.path())
	_ = instance.client.CloseFile(n.path())
	return 0
}

// Read simply returns the data that was already unpacked in the Open call
func (n *gofuseNode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	dest, err := instance.client.ReadFile(n.path(), off, int64(len(dest)))

	if err != nil && err != io.EOF {
		Logger.LogErr("FD : Failed to read the file %s (%s)", n.path(), err.Error())
		return fuse.ReadResultData(dest), fs.ToErrno(err)
	}

	return fuse.ReadResultData(dest), fs.OK
}

func (n *gofuseNode) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	bytes, err := instance.client.WriteFile(n.path(), off, int64(len(buf)), buf)

	if err != nil || bytes != len(buf) {
		Logger.LogErr("FD : Failed to read the file %s (%s)", n.path(), err.Error())
		return uint32(bytes), fs.ToErrno(err)
	}

	return uint32(bytes), fs.OK
}

func (n *gofuseNode) Unlink(ctx context.Context, name string) syscall.Errno {
	err := instance.client.DeleteFile(name)
	if err != nil {
		Logger.LogErr("Unable to delete file %s", name)
		return fs.ToErrno(err)
	}

	p := filepath.Join(n.path(), name)
	err = syscall.Unlink(p)
	return fs.ToErrno(err)
}

func (n *gofuseNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	/*newParentLoopback := toLoopbackNode(newParent)
	if flags&fs.RENAME_EXCHANGE != 0 {
		return n.renameExchange(name, newParentLoopback, newName)
	}

	p1 := filepath.Join(n.path(), name)
	p2 := filepath.Join(newParentLoopback.path(), newName)

	err := syscall.Rename(p1, p2)
	return fs.ToErrno(err)
	*/
	return fs.OK
}

func (n *gofuseNode) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p := filepath.Join(n.path(), name)
	err := syscall.Symlink(target, p)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	//n.preserveOwner(ctx, p)
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		syscall.Unlink(p)
		return nil, fs.ToErrno(err)
	}
	node := &gofuseNode{}
	ch := n.NewInode(ctx, node, fs.StableAttr{})

	out.Attr.FromStat(&st)
	return ch, 0
}

func (n *gofuseNode) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {

	p := filepath.Join(n.path(), name)
	targetNode := gofuseNode{} //toLoopbackNode(target)
	err := syscall.Link(targetNode.path(), p)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		syscall.Unlink(p)
		return nil, fs.ToErrno(err)
	}
	node := &gofuseNode{}
	ch := n.NewInode(ctx, node, fs.StableAttr{})

	out.Attr.FromStat(&st)
	return ch, 0
}

func (n *gofuseNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	p := n.path()

	for l := 256; ; l *= 2 {
		buf := make([]byte, l)
		sz, err := syscall.Readlink(p, buf)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		if sz < len(buf) {
			return buf[:sz], 0
		}
	}
}

func (n *gofuseNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	p := n.path()
	fsa, ok := f.(fs.FileSetattrer)
	if ok && fsa != nil {
		fsa.Setattr(ctx, in, out)
	} else {
		if m, ok := in.GetMode(); ok {
			if err := syscall.Chmod(p, m); err != nil {
				return fs.ToErrno(err)
			}
		}

		uid, uok := in.GetUID()
		gid, gok := in.GetGID()
		if uok || gok {
			suid := -1
			sgid := -1
			if uok {
				suid = int(uid)
			}
			if gok {
				sgid = int(gid)
			}
			if err := syscall.Chown(p, suid, sgid); err != nil {
				return fs.ToErrno(err)
			}
		}

		mtime, mok := in.GetMTime()
		atime, aok := in.GetATime()

		if mok || aok {

			ap := &atime
			mp := &mtime
			if !aok {
				ap = nil
			}
			if !mok {
				mp = nil
			}
			var ts [2]syscall.Timespec
			ts[0] = fuse.UtimeToTimespec(ap)
			ts[1] = fuse.UtimeToTimespec(mp)

			if err := syscall.UtimesNano(p, ts[:]); err != nil {
				return fs.ToErrno(err)
			}
		}

		if sz, ok := in.GetSize(); ok {
			if err := syscall.Truncate(p, int64(sz)); err != nil {
				return fs.ToErrno(err)
			}
		}
	}

	fga, ok := f.(fs.FileGetattrer)
	if ok && fga != nil {
		fga.Getattr(ctx, out)
	} else {
		st := syscall.Stat_t{}
		err := syscall.Lstat(p, &st)
		if err != nil {
			return fs.ToErrno(err)
		}
		out.FromStat(&st)
	}
	return fs.OK
}

/*
func (n *gofuseNode) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	sz, err := unix.Lgetxattr(n.path(), attr, dest)
	return uint32(sz), fs.ToErrno(err)
}

func (n *gofuseNode) Setxattr(ctx context.Context, attr string, data []byte, flags uint32) syscall.Errno {
	err := unix.Lsetxattr(n.path(), attr, data, int(flags))
	return fs.ToErrno(err)
}

func (n *gofuseNode) Removexattr(ctx context.Context, attr string) syscall.Errno {
	err := unix.Lremovexattr(n.path(), attr)
	return fs.ToErrno(err)
}

func (n *gofuseNode) Listxattr(ctx context.Context, dest []byte) (uint32, syscall.Errno) {
	sz, err := unix.Llistxattr(n.path(), dest)
	return uint32(sz), fs.ToErrno(err)
}
*/
