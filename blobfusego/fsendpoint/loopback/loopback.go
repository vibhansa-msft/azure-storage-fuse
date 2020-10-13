package loopback

import (
	"io/ioutil"
	"os"
	"syscall"

	FSFact "github.com/blobfusego/fswrapper/fscreator"
	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"
)

type loopbackFS struct {
	refCount int
	consumer FSIntf.FileSystem
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
type loopbackConfig struct {
	linkPath string
}

var loopbackCfg loopbackConfig

func (f *loopbackFS) InitFS() int {
	Logger.LogDebug("FS : InitFS called")
	loopbackCfg.linkPath = *Config.BlobfuseConfig.TmpPath
	return 0
}

func (f *loopbackFS) DeInitFs() int {
	Logger.LogDebug("FS : DeInitFs called")
	return 0
}

// Get the file system name
func (f *loopbackFS) GetName() string {
	return fsName
}

// Get the ref count
func (f *loopbackFS) GetCount() int {
	return instance.refCount
}

// Set the next component in pipeline for this system
func (f *loopbackFS) SetConsumer(cons FSIntf.FileSystem) int {
	Logger.LogDebug("FS : SetConsumer in " + fsName)
	instance.consumer = cons
	return 0
}

// Get the file system stats
func (f *loopbackFS) StatFS() int {
	return 0
}

// PrintPipeline : Print the current pipeline
func (f *loopbackFS) PrintPipeline() string {
	if instance.consumer != nil {
		return (fsName + " -> " + instance.consumer.PrintPipeline())
	}
	return (fsName + " -> X ")
}

// Directory level operations
func (f *loopbackFS) CreateDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) DeleteDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) OpenDir(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) CloseDir(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadDir(path string) (dirList []FSIntf.BlobAttr) {
	Logger.LogDebug("FS : ReadDir for " + path)

	files, err := ioutil.ReadDir(loopbackCfg.linkPath + "/" + path)
	if err != nil {
		Logger.LogErr("FS : Failed to Read Dir for %s (%v)", path, err)
		return dirList
	}

	for _, f := range files {
		//stat, _ := os.Stat(f.Name())
		stat := f.Sys().(*syscall.Stat_t)
		var attr FSIntf.BlobAttr
		attr.Name = f.Name()
		attr.Size = (uint64)(f.Size())
		attr.Mode = f.Mode()
		attr.Flags = 0
		//dirList[i].Modtime = stat.ModTime()
		attr.NodeID = stat.Ino
		if f.IsDir() {
			attr.Flags.Set(FSIntf.PropFlagIsDir)
		}

		dirList = append(dirList, attr)

	}

	return dirList
}

func (f *loopbackFS) RenameDir(path string, name string) int {
	panic("not implemented") // TODO: Implement
}

// File level operations
func (f *loopbackFS) CreateFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) DeleteFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) OpenFile(path string, mode int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) CloseFile(path string) {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) WriteFile(path string, offset int, length int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) FlushFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReleaseFile(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) UnlinkFile(path string) int {
	panic("not implemented") // TODO: Implement
}

// Symlink operations
func (f *loopbackFS) CreateLink(path string, dst string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) ReadLink(path string, link string) int {
	panic("not implemented") // TODO: Implement
}

// Filesystem level operations
func (f *loopbackFS) GetAttr(path string, attr *FSIntf.BlobAttr) error {
	Logger.LogDebug("FS : GetAttr called for %s", path)

	stat, err := os.Stat(loopbackCfg.linkPath + "/" + path)
	if err != nil {
		Logger.LogErr("FS : Failed to get stat of %s (%v)", path, err)
		return err
	}

	attr.Name = path
	attr.Size = (uint64)(stat.Size())
	attr.Mode = stat.Mode()
	attr.Modtime = stat.ModTime()

	meta := stat.Sys().(*syscall.Stat_t)
	attr.NodeID = meta.Ino
	if stat.IsDir() {
		attr.Flags.Set(FSIntf.PropFlagIsDir)
	}
	return nil
}

func (f *loopbackFS) SetAttr(path string) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) Chmod(path string, mod int) int {
	panic("not implemented") // TODO: Implement
}

func (f *loopbackFS) Chown(path string, owner string) int {
	panic("not implemented") // TODO: Implement
}
