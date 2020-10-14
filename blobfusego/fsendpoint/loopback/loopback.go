package loopback

import (
	"io/ioutil"
	"os"
	"path/filepath"

	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"
)

type loopbackFS struct {
	refCount int
	consumer FSIntf.FileSystem
	lfsPath  string
}

var instance *loopbackFS
var fsName = string("loopback")

var regObj = FSFact.FSManager{CreateObjFunc: CreateObj, ReleaseObjFunc: ReleaseObj}

func init() {
	FSFact.RegisterFileSystem(fsName, regObj)
}

////////////////////////////////////////
//	REQUIRED FOR FSCREATOR TO WPORK

// CreateObj : Create the loopback FS object for factory
func CreateObj() FSIntf.FileSystem {
	if instance == nil {
		instance = &loopbackFS{}
		instance.refCount = 0
		Logger.LogDebug("Created first instances of " + fsName)
	}
	instance.refCount++
	return instance
}

// ReleaseObj : Delete the loopback FS object
func ReleaseObj() {
	instance.refCount--
	if instance.refCount == 0 {
		instance = nil
		Logger.LogDebug("Released all instances of " + fsName)
	}
}

////////////////////////////////////////
// Init/DeInit the filesystem
func (fsys *loopbackFS) InitFS() int {
	Logger.LogDebug("FS : LoopbackFS InitFS")
	fsys.lfsPath = *Config.BlobfuseConfig.TmpPath
	return 0
}

func (fsys *loopbackFS) DeInitFs() int {
	return 0
}

// Set the next component in pipeline for this system
func (fsys *loopbackFS) SetClient(cons FSIntf.FileSystem) int {
	fsys.consumer = cons
	return 0
}

// Get the file system name
func (fsys *loopbackFS) GetName() string {
	return fsName
}

// Get the reference count
func (fsys *loopbackFS) GetCount() int {
	return fsys.refCount
}

// Print the pipeline
func (fsys *loopbackFS) PrintPipeline() string {
	if fsys.consumer != nil {
		return (fsName + " -> " + fsys.consumer.PrintPipeline())
	}
	return (fsName + " -> X ")
}

// Get the file system stats
func (fsys *loopbackFS) StatFS() error {
	return nil
}

// Directory level operations
func (fsys *loopbackFS) CreateDir(name string, mode os.FileMode) error {
	path := filepath.Join(fsys.lfsPath, name)
	return os.Mkdir(path, mode)
}

func (fsys *loopbackFS) DeleteDir(name string) error {
	path := filepath.Join(fsys.lfsPath, name)
	return os.RemoveAll(path)
}

func (fsys *loopbackFS) OpenDir(_ string) error {
	return nil
}

func (fsys *loopbackFS) CloseDir(_ string) error {
	return nil
}

func (fsys *loopbackFS) ReadDir(name string) (lst []FSIntf.BlobAttr, err error) {
	path := filepath.Join(fsys.lfsPath, name)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return lst, err
	}

	for _, file := range files {
		attr := FSIntf.BlobAttr{
			Name:    file.Name(),
			Size:    uint64(file.Size()),
			Mode:    file.Mode(),
			Modtime: file.ModTime(),
		}
		if file.IsDir() {
			attr.Flags.Set(FSIntf.PropFlagIsDir)
		}
		lst = append(lst, attr)
	}
	return lst, nil
}

func (fsys *loopbackFS) RenameDir(old string, new string) error {
	oldPath := filepath.Join(fsys.lfsPath, old)
	newPath := filepath.Join(fsys.lfsPath, new)

	return os.Rename(oldPath, newPath)
}

// File level operations
func (fsys *loopbackFS) CreateFile(name string, mod os.FileMode) error {
	path := filepath.Join(fsys.lfsPath, name)
	_, err := os.Create(path)
	return err
}

func (fsys *loopbackFS) DeleteFile(name string) error {
	path := filepath.Join(fsys.lfsPath, name)
	return os.Remove(path)
}

func (fsys *loopbackFS) OpenFile(_ string) error {
	return nil
}

func (fsys *loopbackFS) CloseFile(_ string) error {
	return nil
}

func (fsys *loopbackFS) ReadFile(name string, offset int64, len int64) (data []byte, err error) {
	path := filepath.Join(fsys.lfsPath, name)

	f, err := os.Open(path)
	if err != nil {
		return data, err
	}

	data = make([]byte, len)
	readLen, err := f.Read(data)
	f.Close()

	return data[:readLen], nil
}

func (fsys *loopbackFS) WriteFile(name string, offset int64, len int64, data []byte) (bytes int, err error) {
	path := filepath.Join(fsys.lfsPath, name)
	f, err := os.OpenFile(path, os.O_RDWR, 0644)

	if err != nil {
		return 0, err
	}

	if _, err := f.Seek(offset, 0); err != nil {
		f.Close()
		return 0, err
	}
	if bytes, err = f.WriteAt(data, len); err != nil {
		f.Close()
		return 0, err
	}

	f.Close()
	return bytes, nil

}

func (fsys *loopbackFS) TruncateFile(name string, _ int64) error {
	path := filepath.Join(fsys.lfsPath, name)
	f, err := os.OpenFile(path, os.O_TRUNC, 0777)

	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func (fsys *loopbackFS) FlushFile(_ string) error {
	return nil
}

func (fsys *loopbackFS) ReleaseFile(_ string) error {
	return nil
}

func (fsys *loopbackFS) UnlinkFile(_ string) error {
	return nil
}

// Symlink operations
func (fsys *loopbackFS) CreateLink(_ string, _ string) error {
	return nil
}

func (fsys *loopbackFS) ReadLink(_ string) (string, error) {
	return "", nil
}

// Filesystem level operations
func (fsys *loopbackFS) GetAttr(name string) (attr FSIntf.BlobAttr, err error) {
	path := filepath.Join(fsys.lfsPath, name)
	f, err := os.Stat(path)
	if err != nil {
		return attr, err
	}

	attr = FSIntf.BlobAttr{
		Name:    f.Name(),
		Size:    uint64(f.Size()),
		Mode:    f.Mode(),
		Modtime: f.ModTime(),
	}
	if f.IsDir() {
		attr.Flags.Set(FSIntf.PropFlagIsDir)
	}

	return attr, nil

}

func (fsys *loopbackFS) SetAttr(_ string, _ FSIntf.BlobAttr) error {
	return nil
}

func (fsys *loopbackFS) Chmod(_ string, _ os.FileMode) error {
	return nil
}

func (fsys *loopbackFS) Chown(_ string, _ string) error {
	return nil
}
