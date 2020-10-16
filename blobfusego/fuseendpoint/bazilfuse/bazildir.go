package bazilfuse

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var _ fs.Node = (*Dir)(nil)
var _ fs.NodeCreater = (*Dir)(nil)
var _ fs.NodeMkdirer = (*Dir)(nil)
var _ fs.NodeRemover = (*Dir)(nil)
var _ fs.NodeRenamer = (*Dir)(nil)
var _ fs.NodeLinker = (*Dir)(nil)
var _ fs.NodeSymlinker = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)

// Dir : Structure representing a directory in system
type Dir struct {
	dirlck sync.RWMutex
	path   string
	attr   fuse.Attr
	valid  bool
}

// FileExists : Check file exists in the directory or not
func (d *Dir) FileExists(name string) bool {
	Logger.LogDebug("FD : Dir FileExists called for %s", d.path)

	path := filepath.Join(d.path, name)

	_, err := BazilFS.client.GetAttr(path)
	if err != nil {
		return false
	}

	return true
}

var ignoreList = map[string]struct{}{
	".Trash":           {},
	".Trash-1000":      {},
	".xdg-volume-info": {},
	"autorun.inf":      {},
}

// Attr : Get the attributes of the given directory
func (d Dir) Attr(ctx context.Context, o *fuse.Attr) error {
	Logger.LogDebug("FD : Dir Attr called for %s", d.path)

	if d.path == "/" {
		// Attr for root called
		*o = BazilFS.root.attr
		return nil
	}

	d.dirlck.Lock()
	defer d.dirlck.Unlock()

	if !d.valid {
		return fuse.ENOENT
	}

	*o = d.attr
	return nil
}

// Lookup : Check whether given object exists in the directory structure or not
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	Logger.LogDebug("FD : Dir Lookup called for %s", d.path)

	if _, ignore := ignoreList[name]; ignore {
		Logger.LogDebug("FD : Ignoring %s", d.path)
		return nil, fuse.ENOENT
	}

	d.dirlck.RLock()
	defer d.dirlck.RUnlock()

	path := filepath.Join(d.path, name)
	Logger.LogDebug("FD : Getting stats for %s", path)

	attr, err := BazilFS.client.GetAttr(path)
	if err != nil {
		Logger.LogErr("FD : Failed to get stats for %s", path)
		return nil, fuse.ENOENT
	}

	if n := nodeMap[path]; n != nil {
		if attr.IsDir() {
			d := n.(*Dir)
			if d.valid {
				d.SetDirAttr(&attr)
				return d, nil
			}
		} else {
			f := n.(*File)
			if f.valid {
				f.SetFileAttr(&attr)
				return f, nil
			}
		}
	}

	switch {
	case attr.IsDir():
		return BazilFS.newDirNode(path, &attr), nil
	default:
		return BazilFS.newFileNode(path, &attr), nil
	}

	return nil, fuse.ENOENT
}

// ReadDirAll : Get the list of objects from a directory
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	Logger.LogDebug("FD : Dir ReadDirAll for %s", d.path)

	d.dirlck.RLock()
	var out []fuse.Dirent

	blobs, err := BazilFS.client.ReadDir(d.path)
	Logger.LogDebug("FD : ReadDir came back with %d elements", len(blobs))

	if err != nil {
		Logger.LogErr("FD : Failed to read directory (%s)", err)
		return nil, err
	}

	for _, node := range blobs {
		de := fuse.Dirent{Name: node.Name}
		if node.IsDir() {
			de.Type = fuse.DT_Dir
		} else if node.IsSymlink() {
			de.Type = fuse.DT_Link
		} else {
			de.Type = fuse.DT_File
		}
		out = append(out, de)
	}

	d.dirlck.RUnlock()
	return out, nil
}

// Mkdir : Create a new directory
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	Logger.LogDebug("FD : Dir Mkdir for %s", d.path)

	d.dirlck.Lock()
	defer d.dirlck.Unlock()

	if exists := d.FileExists(req.Name); exists {
		return nil, fuse.EEXIST
	}

	path := filepath.Join(d.path, req.Name)
	n := BazilFS.newDirNode(path, &FSIntf.BlobAttr{
		Name:    path,
		Size:    4096,
		Mode:    os.ModeDir | Config.BlobfuseConfig.DefaultPerm,
		Modtime: time.Now(),
	})

	if err := BazilFS.client.CreateDir(path, req.Mode); err != nil {
		Logger.LogErr("FD : Failed to create directory %s (%s)", path, err)
		return nil, err
	}

	return n, nil
}

// Create : Create a new entry in directory...
func (d *Dir) Create(ctx context.Context,
	req *fuse.CreateRequest,
	resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	Logger.LogDebug("FD : Dir Create" + req.Name)

	d.dirlck.Lock()
	defer d.dirlck.Unlock()

	if exists := d.FileExists(req.Name); exists {
		return nil, nil, fuse.EEXIST
	}

	path := filepath.Join(d.path, req.Name)

	f := BazilFS.newFileNode(path, &FSIntf.BlobAttr{
		Size:    0,
		Mode:    Config.BlobfuseConfig.DefaultPerm,
		Modtime: time.Now(),
		Flags:   0,
		NodeID:  BazilFS.nextID(),
	})
	f.created = true
	resp.Attr = f.attr
	return f, f, nil
}

// Link : Create a new  link to directory
func (d *Dir) Link(ctx context.Context, req *fuse.LinkRequest, old fs.Node) (newNode fs.Node, err error) {
	Logger.LogDebug("FD : Dir Link" + req.NewName)

	nd := newNode.(*Dir)

	if d.attr.Inode == nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	} else if d.attr.Inode < nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
	} else {
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	}

	if exists := d.FileExists(req.NewName); !exists {
		Logger.LogErr("FD : Link already exists")
		return nil, fuse.ENOENT
	}

	newPath := filepath.Join(nd.path, req.NewName)

	if err := BazilFS.client.CreateLink(d.path, newPath); err != nil {
		Logger.LogErr("FD : Failed to create link (%s)", err)
		return nil, err
	}

	return nd, nil
}

// Symlink : Create a new symlink ...
func (d *Dir) Symlink(ctx context.Context, req *fuse.SymlinkRequest) (fs.Node, error) {
	Logger.LogDebug("FD : Dir Symlink" + req.NewName)

	nd := d
	nd.attr.Mode |= os.ModeSymlink

	if d.attr.Inode == nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	} else if d.attr.Inode < nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
	} else {
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	}

	if exists := d.FileExists(req.NewName); !exists {
		Logger.LogErr("FD : Link already exists")
		return nil, fuse.ENOENT
	}

	targetPath := filepath.Join(d.path, req.Target)
	newPath := filepath.Join(nd.path, req.NewName)
	Logger.LogDebug("FD : Symlink %s -> %s", newPath, targetPath)

	if err := BazilFS.client.CreateLink(targetPath, newPath); err != nil {
		Logger.LogErr("FD : Failed to create link (%s)", err)
		return nil, err
	}

	return nd, nil
}

// Rename : Rename a directory
func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	Logger.LogDebug("FD : Rename directory %s -> %s", req.OldName, req.NewName)

	nd := newDir.(*Dir)

	if d.attr.Inode == nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	} else if d.attr.Inode < nd.attr.Inode {
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
	} else {
		nd.dirlck.Lock()
		defer nd.dirlck.Unlock()
		d.dirlck.Lock()
		defer d.dirlck.Unlock()
	}

	if exists := d.FileExists(req.OldName); !exists {
		Logger.LogErr("FD : Can not rename, file does not exists %s", req.OldName)
		return fuse.ENOENT
	}

	oldPath := filepath.Join(d.path, req.OldName)
	newPath := filepath.Join(nd.path, req.NewName)
	Logger.LogDebug("FD : Dir Rename %s -> %s", oldPath, newPath)

	if err := BazilFS.client.RenameDir(oldPath, newPath); err != nil {
		Logger.LogErr("FD : Failed to rename %s (%s)", req.OldName, err)
		return err
	}
	od := nodeMap[oldPath].(*Dir)
	od.valid = false

	Logger.LogDebug("FD : Rename successful for %s to %s", d.path, nd.path)

	return nil
}

// Remove : Delete a directory
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	Logger.LogDebug("FD : Dir Remove %s", req.Name)

	d.dirlck.Lock()
	defer d.dirlck.Unlock()

	if exists := d.FileExists(req.Name); !exists {
		Logger.LogErr("FD : Cannot delete, file does not exists %s", req.Name)
		return fuse.ENOENT
	}

	path := filepath.Join(d.path, req.Name)
	if err := BazilFS.client.DeleteDir(path); err != nil {
		Logger.LogErr("FD : Failed to delete %s (%s)", req.Name, err)
		return err
	}
	od := nodeMap[path].(*Dir)
	od.valid = false

	return nil
}
