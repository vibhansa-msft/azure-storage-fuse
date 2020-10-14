package bazilfuse

import (
	"io"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	Logger "github.com/blobfusego/global/logger"
	"golang.org/x/net/context"
)

var _ fs.Node = (*File)(nil)
var _ fs.NodeOpener = (*File)(nil)
var _ fs.NodeAccesser = (*File)(nil)

// File : Structure representing a file in system
type File struct {
	sync.RWMutex
	attr    fuse.Attr
	path    string
	created bool
}

func (f *File) getAttr() error {
	Logger.LogDebug("FD : File getAttr %s", f.path)

	attr, err := BazilFS.client.GetAttr(f.path)
	if err != nil {
		Logger.LogErr("FD : Failed to get attribute %s (%s)", f.path, err)
		return err
	}

	f.attr.Size = attr.Size
	f.attr.Mtime = attr.Modtime
	f.attr.Mode = attr.Mode

	return nil
}

// Access : Request to access the file (just ignore)
func (f *File) Access(ctx context.Context, req *fuse.AccessRequest) error {
	Logger.LogDebug("FD : File Access %s", f.path)
	return nil
}

// Attr : Get the attribute of the file...
func (f *File) Attr(ctx context.Context, o *fuse.Attr) error {
	Logger.LogDebug("FD : File Attr %s", f.path)

	f.RLock()
	err := f.getAttr()
	if err != nil {
		Logger.LogErr("FD : Failed to get file attributes %s (%s)", f.path, err)
		return err
	}

	*o = f.attr
	f.RUnlock()
	return nil
}

// Open : Open the file to read/write...
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	Logger.LogDebug("FD : File Open (%s, %d, %d)\n", f.path, int(req.Flags), f.attr.Mode)

	if err := BazilFS.client.OpenFile(f.path); err != nil {
		Logger.LogErr("FD : Failed to open file %s (%s)", f.path, err)
		return nil, err
	}
	return f, nil
}

// Release : Close the file
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	Logger.LogDebug("FD : Release %s)", f.path)

	if err := BazilFS.client.CloseFile(f.path); err != nil {
		Logger.LogErr("FD : Failed to close file %s (%s)", f.path, err)
		return err
	}

	return nil
}

// Read : Read the file
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	Logger.LogDebug("FD : Read %s", f.path)

	f.RLock()
	defer f.RUnlock()

	resp.Data = resp.Data[:req.Size]
	n, err := BazilFS.client.ReadFile(f.path, req.Offset, int64(req.Size))

	if err != nil && err != io.EOF {
		Logger.LogErr("FD : Failed to read the file %s (%s)", f.path, err)
		return err
	}

	resp.Data = n
	return nil
}

// Write : Write data to file
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	Logger.LogDebug("FD : Write %s", f.path)

	f.Lock()
	defer f.Unlock()

	n, err := BazilFS.client.WriteFile(f.path, req.Offset, int64(len(req.Data)), req.Data)
	if err != nil {
		Logger.LogErr("FD : Failed to write to file %s (%s)", f.path, err)
		return err
	}
	resp.Size = n
	return nil
}

var _ fs.NodeSetattrer = (*File)(nil)

// Setattr : Set attribute of a file
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	Logger.LogDebug("FD : Setattr %s", f.path)

	f.Lock()
	defer f.Unlock()

	valid := req.Valid
	if valid.Size() {
		err := BazilFS.client.TruncateFile(f.path, int64(req.Size))
		if err != nil {
			Logger.LogErr("FD : Failed to truncate file %s (%s)", f.path, err)
			return err
		}
		valid &^= fuse.SetattrSize
	}

	if valid.Mode() {
		err := BazilFS.client.Chmod(f.path, req.Mode)
		if err != nil {
			Logger.LogErr("FD : Failed to set mode of file %s (%s)", f.path, err)
			return err
		}
		valid &^= fuse.SetattrMode
	}

	valid &^= fuse.SetattrLockOwner | fuse.SetattrHandle
	if valid != 0 {
		// don't let an unhandled operation slip by without error
		Logger.LogErr("FD : Unsupported attribute for %s", f.path)
		return fuse.ENOSYS
	}
	return nil
}
