package bazilfuse

import (
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type File struct {
	path   string
	size   uint64
	nodeid uint64
	fs     *FS
}

func (f *File) Attr(ctx context.Context, o *fuse.Attr) error {
	Logger.LogDebug("FD : File Attr called for %s", f.path)
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	Logger.LogDebug("FD : File Setattr called for %s", f.path)
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (h fs.Handle, err error) {
	Logger.LogDebug("FD : File Open called for %s", f.path)
	return h, err
}

func (f *File) ReadAll(ctx context.Context) (data []byte, err error) {
	Logger.LogDebug("FD : File ReadAll called for %s", f.path)
	return data, err
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	Logger.LogDebug("FD : File Write called for %s", f.path)
	return nil
}

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	Logger.LogDebug("FD : File Flush called for %s", f.path)
	return nil
}
