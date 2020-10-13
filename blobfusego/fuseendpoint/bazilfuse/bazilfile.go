package bazilfuse

import (
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type File struct {
	path   string
	nodeid uint64
}

// Attr : Get the properties of a file
func (f File) Attr(ctx context.Context, o *fuse.Attr) error {
	Logger.LogDebug("FD : File Attr called for %s", f.path)

	var attr FSIntf.BlobAttr
	if err := instance.consumer.GetAttr(f.path, &attr); err != nil {
		return err
	}

	BlobAttrToFuseAttr(&attr, o)
	return nil
}

// Setattr : Set the attributes of a given file
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	Logger.LogDebug("FD : File Setattr called for %s", f.path)
	return nil
}

// Open : Open a file for read/write
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (h fs.Handle, err error) {
	Logger.LogDebug("FD : File Open called for %s", f.path)
	return h, err
}

// ReadAll : Read all the contenets of given file
func (f *File) ReadAll(ctx context.Context) (data []byte, err error) {
	Logger.LogDebug("FD : File ReadAll called for %s", f.path)
	return data, err
}

// Write : Write data to a file
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	Logger.LogDebug("FD : File Write called for %s", f.path)
	return nil
}

// Flush : Flush data to file on close
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	Logger.LogDebug("FD : File Flush called for %s", f.path)
	return nil
}
